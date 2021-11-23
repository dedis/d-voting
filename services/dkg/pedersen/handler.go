package pedersen

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	evotingTypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/cosi/threshold"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	pedersen "go.dedis.ch/kyber/v3/share/dkg/pedersen"
	vss "go.dedis.ch/kyber/v3/share/vss/pedersen"
	"golang.org/x/xerrors"
)

// recvResponseTimeout is the maximum time a node will wait for a response
const recvResponseTimeout = time.Second * 10

// state is a struct contained in a handler that allows an actor to read the
// state of that handler. The actor should only use the getter functions to read
// the attributes.
type state struct {
	sync.Mutex
	distrKey     kyber.Point
	participants []mino.Address
}

func (s *state) Done() bool {
	s.Lock()
	defer s.Unlock()
	return s.distrKey != nil && s.participants != nil
}

func (s *state) GetDistKey() kyber.Point {
	s.Lock()
	defer s.Unlock()
	return s.distrKey
}

func (s *state) SetDistKey(key kyber.Point) {
	s.Lock()
	s.distrKey = key
	s.Unlock()
}

func (s *state) GetParticipants() []mino.Address {
	s.Lock()
	defer s.Unlock()
	return s.participants
}

func (s *state) SetParticipants(addrs []mino.Address) {
	s.Lock()
	s.participants = addrs
	s.Unlock()
}

// Handler represents the RPC executed on each node
//
// - implements mino.Handler
type Handler struct {
	mino.UnsupportedHandler
	sync.RWMutex
	dkg       *pedersen.DistKeyGenerator
	privKey   kyber.Scalar
	me        mino.Address
	privShare *share.PriShare
	startRes  *state
	service   ordering.Service
	evoting   bool
	pubkey    kyber.Point
}

// NewHandler creates a new handler
func NewHandler(privKey kyber.Scalar, me mino.Address, service ordering.Service, evoting bool, pubkey kyber.Point) *Handler {
	return &Handler{
		privKey:  privKey,
		me:       me,
		startRes: &state{},
		service:  service,
		evoting:  evoting,
		pubkey:   pubkey,
	}
}

// Stream implements mino.Handler. It allows one to stream messages to the
// players.
func (h *Handler) Stream(out mino.Sender, in mino.Receiver) error {
	// Note: one should never assume any synchronous properties on the messages.
	// For example we can not expect to receive the start message from the
	// initiator of the DKG protocol first because some node could have received
	// this start message earlier than us, start their DKG work by sending
	// messages to the other nodes, and then we might get their messages before
	// the start message.

	deals := []types.Deal{}
	responses := []*pedersen.Response{}

mainSwitch:
	from, msg, err := in.Recv(context.Background())
	if err != nil {
		return xerrors.Errorf("failed to receive: %v", err)
	}

	// We expect a Start message or a decrypt request at first, but we might
	// receive other messages in the meantime, like a Deal.
	switch msg := msg.(type) {

	case types.Start:
		err := h.start(msg, deals, responses, from, out, in)
		if err != nil {
			return xerrors.Errorf("failed to start: %v", err)
		}

	case types.Deal:
		// This is a special case where a DKG started, some nodes received the
		// start signal and started sending their deals but we have not yet
		// received our start signal. In this case we collect the Deals while
		// waiting for the start signal.
		deals = append(deals, msg)
		goto mainSwitch

	case types.Response:
		// This is a special case where a DKG started, some nodes received the
		// start signal and started sending their deals but we have not yet
		// received our start signal. In this case we collect the Response while
		// waiting for the start signal.
		response := &pedersen.Response{
			Index: msg.GetIndex(),
			Response: &vss.Response{
				SessionID: msg.GetResponse().GetSessionID(),
				Index:     msg.GetResponse().GetIndex(),
				Status:    msg.GetResponse().GetStatus(),
				Signature: msg.GetResponse().GetSignature(),
			},
		}
		responses = append(responses, response)
		goto mainSwitch

	case types.DecryptRequest:
		if !h.startRes.Done() {
			return xerrors.Errorf("you must first initialize DKG. Did you " +
				"call setup() first?")
		}

		if h.evoting {
			isShuffled, err := h.checkIsShuffled(msg.K, msg.C, msg.GetElectionId())
			if err != nil {
				return xerrors.Errorf("failed to check if the ciphertext has been shuffled: %v", err)
			}

			if !isShuffled {
				return xerrors.Errorf("the ciphertext has not been shuffled")
			}
		}

		// TODO: check if started before
		h.RLock()
		S := suite.Point().Mul(h.privShare.V, msg.K)
		h.RUnlock()

		partial := suite.Point().Sub(msg.C, S)

		h.RLock()
		decryptReply := types.NewDecryptReply(
			// TODO: check if using the private index is the same as the public
			// index.
			int64(h.privShare.I),
			partial,
		)
		h.RUnlock()

		errs := out.Send(decryptReply, from)
		err = <-errs
		if err != nil {
			return xerrors.Errorf("got an error while sending the decrypt "+
				"reply: %v", err)
		}

	case types.GetPeerPubKey:
		response := types.NewGetPeerPubKeyResp(h.pubkey)
		errs := out.Send(response, from)
		err = <-errs
		if err != nil {
			return xerrors.Errorf("got an error while sending the get peer pubkey resp "+
				"reply: %v", err)
		}
		goto mainSwitch

	default:
		return xerrors.Errorf("expected Start message, decrypt request or "+
			"Deal as first message, got: %T", msg)
	}

	return nil
}

// start is called when the node has received its start message. Note that we
// might have already received some deals from other nodes in the meantime. The
// function handles the DKG creation protocol.
func (h *Handler) start(start types.Start, receivedDeals []types.Deal,
	receivedResps []*pedersen.Response, from mino.Address, out mino.Sender,
	in mino.Receiver) error {

	if len(start.GetAddresses()) != len(start.GetPublicKeys()) {
		return xerrors.Errorf("there should be as many players as "+
			"pubKey: %d := %d", len(start.GetAddresses()), len(start.GetPublicKeys()))
	}

	// 1. Create the DKG
	threshold := threshold.ByzantineThreshold(len(start.GetPublicKeys()))
	d, err := pedersen.NewDistKeyGenerator(suite, h.privKey, start.GetPublicKeys(), threshold)
	if err != nil {
		return xerrors.Errorf("failed to create new DKG: %v", err)
	}
	h.dkg = d

	// 2. Send my Deals to the other nodes
	deals, err := d.Deals()
	if err != nil {
		return xerrors.Errorf("failed to compute the deals: %v", err)
	}

	// use a waitgroup to send all the deals asynchronously and wait
	var wg sync.WaitGroup
	wg.Add(len(deals))

	for i, deal := range deals {
		dealMsg := types.NewDeal(
			deal.Index,
			deal.Signature,
			types.NewEncryptedDeal(
				deal.Deal.DHKey,
				deal.Deal.Signature,
				deal.Deal.Nonce,
				deal.Deal.Cipher,
			),
		)

		errs := out.Send(dealMsg, start.GetAddresses()[i])
		go func(errs <-chan error) {
			err, more := <-errs
			if more {
				dela.Logger.Warn().Msgf("got an error while sending deal: %v", err)
			}
			wg.Done()
		}(errs)
	}

	wg.Wait()

	dela.Logger.Trace().Msgf("%s sent all its deals", h.me)

	numReceivedDeals := 0

	// Process the deals we received before the start message
	for _, deal := range receivedDeals {
		err = h.handleDeal(deal, from, start.GetAddresses(), out)
		if err != nil {
			dela.Logger.Warn().Msgf("%s failed to handle received deal "+
				"from %s: %v", h.me, from, err)
		}
		numReceivedDeals++
	}

	// If there are N nodes, then N nodes first send (N-1) Deals. Then each node
	// send a response to every other nodes. So the number of responses a node
	// get is (N-1) * (N-1), where (N-1) should equal len(deals).
	for numReceivedDeals < len(deals) {
		from, msg, err := in.Recv(context.Background())
		if err != nil {
			return xerrors.Errorf("failed to receive after sending deals: %v", err)
		}

		switch msg := msg.(type) {

		case types.Deal:
			// 4. Process the Deal and Send the response to all the other nodes
			err = h.handleDeal(msg, from, start.GetAddresses(), out)
			if err != nil {
				dela.Logger.Warn().Msgf("%s failed to handle received deal "+
					"from %s: %v", h.me, from, err)
				return xerrors.Errorf("failed to handle deal from '%s': %v", from, err)
			}
			numReceivedDeals++

		case types.Response:
			// 5. Processing responses
			dela.Logger.Trace().Msgf("%s received response from %s", h.me, from)
			response := &pedersen.Response{
				Index: msg.GetIndex(),
				Response: &vss.Response{
					SessionID: msg.GetResponse().GetSessionID(),
					Index:     msg.GetResponse().GetIndex(),
					Status:    msg.GetResponse().GetStatus(),
					Signature: msg.GetResponse().GetSignature(),
				},
			}
			receivedResps = append(receivedResps, response)

		default:
			return xerrors.Errorf("unexpected message: %T", msg)
		}
	}

	h.startRes.SetParticipants(start.GetAddresses())

	err = h.certify(receivedResps, out, in, from)
	if err != nil {
		return xerrors.Errorf("failed to certify: %v", err)
	}

	return nil
}

func (h *Handler) certify(resps []*pedersen.Response, out mino.Sender,
	in mino.Receiver, from mino.Address) error {

	for _, response := range resps {
		_, err := h.dkg.ProcessResponse(response)
		if err != nil {
			dela.Logger.Warn().Msgf("%s failed to process response: %v", h.me, err)
		}
	}

	for !h.dkg.Certified() {
		ctx, cancel := context.WithTimeout(context.Background(),
			recvResponseTimeout)
		defer cancel()

		from, msg, err := in.Recv(ctx)
		if err != nil {
			return xerrors.Errorf("failed to receive after sending deals: %v", err)
		}

		switch msg := msg.(type) {

		case types.Response:
			// 5. Processing responses
			dela.Logger.Trace().Msgf("%s received response from %s", h.me, from)
			response := &pedersen.Response{
				Index: msg.GetIndex(),
				Response: &vss.Response{
					SessionID: msg.GetResponse().GetSessionID(),
					Index:     msg.GetResponse().GetIndex(),
					Status:    msg.GetResponse().GetStatus(),
					Signature: msg.GetResponse().GetSignature(),
				},
			}

			_, err = h.dkg.ProcessResponse(response)
			if err != nil {
				dela.Logger.Warn().Msgf("%s, failed to process response "+
					"from '%s': %v", h.me, from, err)
			}

		default:
			return xerrors.Errorf("expected a response, got: %T", msg)
		}
	}

	dela.Logger.Trace().Msgf("%s is certified", h.me)

	// 6. Send back the public DKG key
	distrKey, err := h.dkg.DistKeyShare()
	if err != nil {
		return xerrors.Errorf("failed to get distr key: %v", err)
	}

	// 7. Update the state before sending to acknowledgement to the
	// orchestrator, so that it can process decrypt requests right away.
	h.startRes.SetDistKey(distrKey.Public())

	h.Lock()
	h.privShare = distrKey.PriShare()
	h.Unlock()

	done := types.NewStartDone(distrKey.Public())
	err = <-out.Send(done, from)
	if err != nil {
		return xerrors.Errorf("got an error while sending pub key: %v", err)
	}

	return nil
}

// handleDeal process the Deal and send the responses to the other nodes.
func (h *Handler) handleDeal(msg types.Deal, from mino.Address, addrs []mino.Address,
	out mino.Sender) error {

	dela.Logger.Trace().Msgf("%s received deal from %s", h.me, from)

	deal := &pedersen.Deal{
		Index: msg.GetIndex(),
		Deal: &vss.EncryptedDeal{
			DHKey:     msg.GetEncryptedDeal().GetDHKey(),
			Signature: msg.GetEncryptedDeal().GetSignature(),
			Nonce:     msg.GetEncryptedDeal().GetNonce(),
			Cipher:    msg.GetEncryptedDeal().GetCipher(),
		},
		Signature: msg.GetSignature(),
	}

	response, err := h.dkg.ProcessDeal(deal)
	if err != nil {
		return xerrors.Errorf("failed to process deal from %s: %v",
			h.me, err)
	}

	resp := types.NewResponse(
		response.Index,
		types.NewDealerResponse(
			response.Response.Index,
			response.Response.Status,
			response.Response.SessionID,
			response.Response.Signature,
		),
	)

	for _, addr := range addrs {
		if addr.Equal(h.me) {
			continue
		}

		errs := out.Send(resp, addr)
		err = <-errs
		if err != nil {
			dela.Logger.Warn().Msgf("got an error while sending "+
				"response: %v", err)
			return xerrors.Errorf("failed to send response to '%s': %v", addr, err)
		}

	}

	return nil
}

// checkIsShuffled allows to check if the ciphertext to decrypt has been
// previously shuffled
func (h *Handler) checkIsShuffled(K kyber.Point, C kyber.Point, electionId string) (bool, error) {

	electionIDBuff, err := hex.DecodeString(electionId)
	if err != nil {
		return false, xerrors.Errorf("failed to decode electionID: %v", err)
	}

	proof, err := h.service.GetProof(electionIDBuff)
	if err != nil {
		return false, xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election := new(evotingTypes.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return false, xerrors.Errorf("failed to unmarshal Election: %v", err)
	}

	for _, ct := range election.ShuffleInstances[election.ShuffleThreshold-1].ShuffledBallots {
		kPrime, cPrime, err := ct.GetPoints()
		if err != nil {
			return false, xerrors.Errorf("failed to get points: %v", err)
		}

		if kPrime.Equal(K) && cPrime.Equal(C) {
			return true, nil
		}
	}

	return false, nil

}
