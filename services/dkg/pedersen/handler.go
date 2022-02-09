package pedersen

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"github.com/dedis/d-voting/contracts/evoting"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/crypto"
	jsondela "go.dedis.ch/dela/serde/json"
	"sync"
	"time"

	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/cosi/threshold"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	pedersen "go.dedis.ch/kyber/v3/share/dkg/pedersen"
	vss "go.dedis.ch/kyber/v3/share/vss/pedersen"
	"golang.org/x/xerrors"
)

// recvResponseTimeout is the maximum time a node will wait for a response
const recvResponseTimeout = time.Second * 10

// Handler represents the RPC executed on each node
//
// - implements mino.Handler
type Handler struct {
	mino.UnsupportedHandler
	sync.RWMutex

	me              mino.Address
	service         ordering.Service
	dkg             *pedersen.DistKeyGenerator
	pool            pool.Pool
	txmnger         txn.Manager
	pubSharesSigner crypto.Signer

	// These are persistent, see HandlerData
	startRes  *state
	privShare *share.PriShare
	privKey   kyber.Scalar
	pubKey    kyber.Point

	context serde.Context
}

// NewHandler creates a new handler
func NewHandler(me mino.Address, service ordering.Service, pool pool.Pool,
	txnmngr txn.Manager, pubSharesSigner crypto.Signer, handlerData HandlerData,
	context serde.Context) *Handler {

	privKey := handlerData.PrivKey
	pubKey := handlerData.PubKey
	startRes := handlerData.StartRes
	privShare := handlerData.PrivShare

	return &Handler{
		me:              me,
		service:         service,
		pool:            pool,
		txmnger:         txnmngr,
		pubSharesSigner: pubSharesSigner,

		startRes:  startRes,
		privShare: privShare,
		privKey:   privKey,
		pubKey:    pubKey,

		context: context,
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

		shuffleInstances, err := h.getShuffleIfValid(msg.GetElectionId())
		if err != nil {
			return xerrors.Errorf("failed to check if the shuffle is over: %v", err)
		}

		publicShares := make([][]etypes.PubShare, len(shuffleInstances))
		numberOfShuffles := len(shuffleInstances)

		// TODO: check if started before
		h.RLock()

		for _, ballot := range shuffleInstances[numberOfShuffles-1].ShuffledBallots {
			ballotShares := make([]etypes.PubShare, len(ballot))

			for _, ciphertext := range ballot {
				S := suite.Point().Mul(h.privShare.V, ciphertext.K)

				partialVal := suite.Point().Sub(ciphertext.C, S)

				ballotShares = append(ballotShares, partialVal)
			}

			publicShares = append(publicShares, ballotShares)
		}

		h.RUnlock()

		// loop until our transaction has been accepted, or enough nodes
		// submitted their pubShares
		for {
			election, err := getElection(h.context, msg.GetElectionId(), h.service)
			if err != nil {
				return xerrors.Errorf("could not get the election: %v", err)
			}

			//TODO: Works with current "shuffleThreshold", but the shuffle threshold
			// should be smaller in theory ? (1/3 + 1 vs 2/3 + 1 ? )
			differentPubShares := 0
			for _, submission := range election.PubSharesArchive {
				if submission != nil {
					differentPubShares++
				}
			}
			if differentPubShares >= election.ShuffleThreshold {
				dela.Logger.Info().Msgf("decryption possible with shares from %d nodes",
					differentPubShares)
				return nil
			}

			tx, err := makeTx(&election, publicShares, h.privShare.I, h.txmnger, h.pubSharesSigner)
			if err != nil {
				return xerrors.Errorf("failed to make tx: %v", err)
			}

			//TODO: Define in term of size of election ? (same in shuffle)
			watchTimeout := time.Second * 5
			watchCtx, cancel := context.WithTimeout(context.Background(), watchTimeout)
			defer cancel()

			events := h.service.Watch(watchCtx)

			err = h.pool.Add(tx)
			if err != nil {
				return xerrors.Errorf("failed to add transaction to the pool: %v", err)
			}

			accepted, msg := watchTx(events, tx.GetID())

			if !accepted {
				err = h.txmnger.Sync()
				if err != nil {
					return xerrors.Errorf("failed to sync manager: %v", err)
				}
			}

			if accepted {
				dela.Logger.Info().Msgf("our pubShares have been accepted on the chain, "+
					"total # of submissions = %d, "+
					"index: %v", len(election.PubSharesArchive), h.privShare.I)
				return nil
			}

			dela.Logger.Info().Msgf("submission of pubShares denied: %v", msg)

			cancel()
		}

	case types.GetPeerPubKey:
		response := types.NewGetPeerPubKeyResp(h.pubKey)
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

	// create the DKG
	thrshold := threshold.ByzantineThreshold(len(start.GetPublicKeys()))
	d, err := pedersen.NewDistKeyGenerator(suite, h.privKey, start.GetPublicKeys(), thrshold)
	if err != nil {
		return xerrors.Errorf("failed to create new DKG: %v", err)
	}
	h.dkg = d

	// send my Deals to the other nodes
	deals, err := h.dkg.Deals()
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
			// Process the Deal and Send the response to all the other nodes
			err = h.handleDeal(msg, from, start.GetAddresses(), out)
			if err != nil {
				dela.Logger.Warn().Msgf("%s failed to handle received deal "+
					"from %s: %v", h.me, from, err)
				return xerrors.Errorf("failed to handle deal from '%s': %v", from, err)
			}
			numReceivedDeals++

		case types.Response:
			// Process responses
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
			// Processing responses
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

	// Send back the public DKG key
	distKey, err := h.dkg.DistKeyShare()
	if err != nil {
		return xerrors.Errorf("failed to get distr key: %v", err)
	}

	// Update the state before sending to acknowledgement to the
	// orchestrator, so that it can process decrypt requests right away.
	h.startRes.SetDistKey(distKey.Public())

	h.Lock()
	h.privShare = distKey.PriShare()
	h.Unlock()

	done := types.NewStartDone(distKey.Public())
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

// getShuffleIfValid allows checking if enough shuffles have been made on the
// ballots.
func (h *Handler) getShuffleIfValid(electionID string) ([]etypes.ShuffleInstance, error) {
	election, err := getElection(h.context, electionID, h.service)
	if err != nil {
		return nil, xerrors.Errorf("could not get the election: %v", err)
	}

	if len(election.ShuffleInstances) == 0 {
		return nil, xerrors.New("election has no shuffles")
	}

	if election.Status != etypes.ShuffledBallots {
		return nil, xerrors.New("ballots have not been shuffled")
	}

	return election.ShuffleInstances, nil
}

// MarshalJSON returns a JSON-encoded bytestring containing all the data in the
// Handler that is meant to be persistent. It allows for saving the data to
// disk.
func (h *Handler) MarshalJSON() ([]byte, error) {
	handlerData := HandlerData{
		StartRes:  h.startRes,
		PrivShare: h.privShare,
		PrivKey:   h.privKey,
		PubKey:    h.pubKey,
	}

	return handlerData.MarshalJSON()
}

// HandlerData is used to synchronise actors between the DKG and the filesystem.
type HandlerData struct {
	StartRes  *state
	PrivShare *share.PriShare
	PubKey    kyber.Point
	PrivKey   kyber.Scalar
}

// NewHandlerData generates new actor data.
func NewHandlerData() HandlerData {
	privKey := suite.Scalar().Pick(suite.RandomStream())
	pubKey := suite.Point().Mul(privKey, nil)

	return HandlerData{
		StartRes: &state{},
		PubKey:   pubKey,
		PrivKey:  privKey,
	}
}

// MarshalJSON returns a JSON-encoded bytestring containing all the data in
// the Handler that is meant to be persistent.
// It allows for saving the data to disk.
func (hd *HandlerData) MarshalJSON() ([]byte, error) {
	// Marshal StartRes
	startResBuf, err := hd.StartRes.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// Marshal PrivShare
	var privShareBuf []byte
	if hd.PrivShare != nil {
		privShareVBuf, err := hd.PrivShare.V.MarshalBinary()
		if err != nil {
			return nil, err
		}
		privShareBuf, err = json.Marshal(&struct {
			I int    `json:",omitempty"`
			V []byte `json:",omitempty"`
		}{
			I: hd.PrivShare.I,
			V: privShareVBuf,
		})
		if err != nil {
			return nil, err
		}
	}

	// Marshal PubKey
	pubKeyBuf, err := hd.PubKey.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Marshal PrivKey
	privKeyBuf, err := hd.PrivKey.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return json.Marshal(&struct {
		StartRes  []byte `json:",omitempty"`
		PrivShare []byte `json:",omitempty"`
		PubKey    []byte
		PrivKey   []byte
	}{
		StartRes:  startResBuf,
		PrivShare: privShareBuf,
		PubKey:    pubKeyBuf,
		PrivKey:   privKeyBuf,
	})
}

// UnmarshalJSON fills a HandlerData with previously marshalled data.
func (hd *HandlerData) UnmarshalJSON(data []byte) error {
	aux := &struct {
		StartRes  []byte `json:",omitempty"`
		PrivShare []byte `json:",omitempty"`
		PubKey    []byte
		PrivKey   []byte
	}{}
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	// Unmarshal StartRes
	hd.StartRes = &state{}
	hd.StartRes.UnmarshalJSON(aux.StartRes)

	// Unmarshal PrivShare
	if aux.PrivShare == nil {
		hd.PrivShare = nil
	} else {
		privShareBuf := &struct {
			I int
			V []byte
		}{}
		err = json.Unmarshal(aux.PrivShare, privShareBuf)
		if err != nil {
			return err
		}
		privShareV := suite.Scalar()
		privShareV.UnmarshalBinary(privShareBuf.V)
		privShare := &share.PriShare{
			I: privShareBuf.I,
			V: privShareV,
		}
		hd.PrivShare = privShare
	}

	// Unmarshal PubKey
	pubKey := suite.Point()
	pubKey.UnmarshalBinary(aux.PubKey)
	hd.PubKey = pubKey

	// Unmarshal PrivKey
	privKey := suite.Scalar()
	privKey.UnmarshalBinary(aux.PrivKey)
	hd.PrivKey = privKey

	return nil
}

// state is a struct contained in a handler that allows an actor to read the
// state of that handler. The actor should only use the getter functions to read
// the attributes.
type state struct {
	sync.Mutex
	distKey      kyber.Point
	participants []mino.Address
}

func (s *state) Done() bool {
	s.Lock()
	defer s.Unlock()
	return s.distKey != nil && s.participants != nil
}

func (s *state) GetDistKey() kyber.Point {
	s.Lock()
	defer s.Unlock()
	return s.distKey
}

func (s *state) SetDistKey(key kyber.Point) {
	s.Lock()
	defer s.Unlock()
	s.distKey = key
}

func (s *state) GetParticipants() []mino.Address {
	s.Lock()
	defer s.Unlock()
	return s.participants
}

func (s *state) SetParticipants(addrs []mino.Address) {
	s.Lock()
	defer s.Unlock()
	s.participants = addrs
}

func (s *state) MarshalJSON() ([]byte, error) {
	s.Lock()
	defer s.Unlock()

	var distKeyBuf []byte
	var participantsBuf [][]byte
	var err error

	if s.distKey != nil {
		distKeyBuf, err = s.distKey.MarshalBinary()
		if err != nil {
			return nil, err
		}

		participantsBuf = make([][]byte, len(s.participants))
		for i, p := range s.participants {
			pBuf, err := p.MarshalText()
			if err != nil {
				return nil, err
			}
			participantsBuf[i] = pBuf
		}
	}

	return json.Marshal(&struct {
		DistKey      []byte   `json:",omitempty"`
		Participants [][]byte `json:",omitempty"`
	}{
		DistKey:      distKeyBuf,
		Participants: participantsBuf,
	})
}

func (s *state) UnmarshalJSON(data []byte) error {
	aux := &struct {
		DistKey      []byte
		Participants [][]byte
	}{}
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	if aux.DistKey != nil {
		distKey := suite.Point()
		err = distKey.UnmarshalBinary(aux.DistKey)
		if err != nil {
			return err
		}
		s.SetDistKey(distKey)
	} else {
		s.SetDistKey(nil)
	}

	if aux.Participants != nil {
		// TODO: Is using a fake implementation a problem?
		f := fake.NewBadMino().GetAddressFactory()
		var participants = make([]mino.Address, len(aux.Participants))
		for i := 0; i < len(aux.Participants); i++ {
			participants[i] = f.FromText(aux.Participants[i])
		}
		s.SetParticipants(participants)
	} else {
		s.SetParticipants(nil)
	}

	return nil
}

// watchTx checks the transaction to find one that match txID. Return if the
// transaction has been accepted or not. Will also return false if/when the
// events chan is closed, which is expected to happen.
func watchTx(events <-chan ordering.Event, txID []byte) (bool, string) {
	for event := range events {
		for _, res := range event.Transactions {
			if !bytes.Equal(res.GetTransaction().GetID(), txID) {
				continue
			}

			dela.Logger.Info().Hex("id", txID).Msg("transaction included in the block")

			accepted, msg := res.GetStatus()
			if accepted {
				return true, ""
			}

			return false, msg
		}
	}

	return false, "watch timeout"
}

func makeTx(election *etypes.Election, pubShares etypes.PubShares, index int, manager txn.Manager,
	pubSharesSigner crypto.Signer) (txn.Transaction, error) {

	pubShareTx := etypes.RegisterPubShares{
		ElectionID: election.ElectionID,
		PubShares:  pubShares,
		Index:      index,
	}

	h := sha256.New()

	err := pubShareTx.Fingerprint(h)
	if err != nil {
		return nil, xerrors.Errorf("failed to get fingerprint: %v", err)
	}

	hash := h.Sum(nil)

	// Sign the pubShares :
	signature, err := pubSharesSigner.Sign(hash)
	if err != nil {
		return nil, xerrors.Errorf("Could not sign the pubShares : %v", err)
	}

	encodedSignature, err := signature.Serialize(jsondela.NewContext())
	if err != nil {
		return nil, xerrors.Errorf("Could not encode signature as []byte : %v ", err)
	}

	// Complete transaction:
	pubShareTx.Signature = encodedSignature

	js, err := json.Marshal(pubShareTx)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal "+
			"RegisterPubSharesTransaction: %v", err)
	}

	args := make([]txn.Arg, 3)
	args[0] = txn.Arg{
		Key:   native.ContractArg,
		Value: []byte(evoting.ContractName),
	}
	args[1] = txn.Arg{
		Key:   evoting.CmdArg,
		Value: []byte(evoting.CmdRegisterPubShares),
	}
	args[2] = txn.Arg{
		Key:   evoting.ElectionArg,
		Value: js,
	}

	tx, err := manager.Make(args...)
	if err != nil {
		return nil, xerrors.Errorf("failed to use manager: %v", err.Error())
	}

	return tx, nil
}
