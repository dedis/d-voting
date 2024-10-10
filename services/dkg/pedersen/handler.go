package pedersen

import (
	"bytes"
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/rs/zerolog"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino/minogrpc/session"
	jsondela "go.dedis.ch/dela/serde/json"

	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
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

// the receiving time out, after which we check if the DKG setup is done or not.
// Allows to exit the loop.
const recvTimeout = time.Second * 2

// the time after which we expect new messages (deals or responses) to be
// received.
const retryTimeout = time.Second * 1

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
	formFac serde.Factory

	log     zerolog.Logger
	running bool

	saveState func(*Handler)

	status *dkg.Status
}

// NewHandler creates a new handler
func NewHandler(me mino.Address, service ordering.Service, pool pool.Pool,
	txnmngr txn.Manager, pubSharesSigner crypto.Signer, handlerData HandlerData,
	context serde.Context, formFac serde.Factory, status *dkg.Status,
	saveState func(*Handler)) *Handler {

	privKey := handlerData.PrivKey
	pubKey := handlerData.PubKey
	startRes := handlerData.StartRes
	privShare := handlerData.PrivShare

	log := dela.Logger.With().Str("role", "DKG").Str("address", me.String()).Logger()

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
		formFac: formFac,

		log:     log,
		running: false,

		saveState: saveState,
		status:    status,
	}
}

// Stream implements mino.Handler. It allows one to stream messages to the
// players.
func (h *Handler) Stream(out mino.Sender, in mino.Receiver) error {
	// Note: one should never assume any synchronous properties on the messages.
	// For example, we can not expect to receive the start message from the
	// initiator of the DKG protocol first because some node could have received
	// this start message earlier than us, start their DKG work by sending
	// messages to the other nodes, and then we might get their messages before
	// the start message.

	// We make sure not additional request is accepted if a setup is in
	// progress.
	h.Lock()
	if !h.startRes.Done() && h.running {
		h.Unlock()
		return xerrors.Errorf("DKG is running")
	}
	if !h.startRes.Done() {
		// This is the first setup
		h.running = true
		defer func() {
			h.running = false
		}()
	}
	h.Unlock()

	deals := list.New()
	responses := list.New()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), recvTimeout)
		from, msg, err := in.Recv(ctx)
		cancel()

		if errors.Is(err, context.DeadlineExceeded) {
			if h.startRes.Done() {
				return nil
			}

			continue
		}

		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return xerrors.Errorf("failed to receive: %v", err)
		}

		h.log.Info().Msgf("received message from %s: %T\n", from, msg)

		// We expect a Start message or a decrypt request at first, but we might
		// receive other messages in the meantime, like a Deal.
		switch msg := msg.(type) {

		case types.Start:
			err := h.start(msg, deals, responses, from, out)
			if err != nil {
				return xerrors.Errorf("failed to start: %v", err)
			}

		case types.Deal:
			// This is a special case where a DKG started, some nodes received the
			// start signal and started sending their deals but we have not yet
			// received our start signal. In this case we collect the Deals while
			// waiting for the start signal.
			h.Lock()
			deals.PushBack(msg)
			h.Unlock()

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

			h.Lock()
			responses.PushBack(response)
			h.Unlock()

		case types.DecryptRequest:

			if !h.startRes.Done() {
				return xerrors.Errorf("you must first initialize DKG. Did you " +
					"call setup() first?")
			}

			err = h.handleDecryptRequest(msg.GetFormId())
			if err != nil {
				return xerrors.Errorf("could not send pubShares: %v", err)
			}

			return nil

		case types.GetPeerPubKey:
			response := types.NewGetPeerPubKeyResp(h.pubKey)
			errs := out.Send(response, from)
			err = <-errs
			if err != nil {
				return xerrors.Errorf("got an error while sending the get peer pubkey resp "+
					"reply: %v", err)
			}

		default:
			return xerrors.Errorf("expected Start message, decrypt request or "+
				"Deal as first message, got: %T", msg)
		}
	}
}

// start is called when the node has received its start message. Note that we
// might have already received some deals from other nodes in the meantime. The
// function handles the DKG creation protocol.
func (h *Handler) start(start types.Start, deals, resps *list.List, from mino.Address,
	out mino.Sender) error {

	if len(start.GetAddresses()) != len(start.GetPublicKeys()) {
		return xerrors.Errorf("there should be as many players as "+
			"pubKey: %d := %d", len(start.GetAddresses()), len(start.GetPublicKeys()))
	}

	// create the DKG
	t := threshold.ByzantineThreshold(len(start.GetPublicKeys()))
	d, err := pedersen.NewDistKeyGenerator(suite, h.privKey, start.GetPublicKeys(), t)
	if err != nil {
		return xerrors.Errorf("failed to create new DKG: %v", err)
	}

	h.dkg = d
	h.startRes.SetParticipants(start.GetAddresses())

	// asynchronously start the procedure. This allows for receiving messages
	// in the main for loop in the meantime.
	go h.doDKG(deals, resps, out, from)

	return nil
}

// doDKG calls the subsequent DKG steps
func (h *Handler) doDKG(deals, resps *list.List, out mino.Sender, from mino.Address) {
	h.log.Info().Str("action", "deal").Msg("new state")
	*h.status = dkg.Status{Status: dkg.Dealing}
	h.deal(out)

	h.log.Info().Str("action", "respond").Msg("new state")
	*h.status = dkg.Status{Status: dkg.Responding}
	h.respond(deals, out)

	h.log.Info().Str("action", "certify").Msg("new state")
	*h.status = dkg.Status{Status: dkg.Certifying}
	err := h.certify(resps, out)
	if err != nil {
		dela.Logger.Error().Msgf("failed to certify: %v", err)
		return
	}

	h.log.Info().Str("action", "finalize").Msg("new state")
	*h.status = dkg.Status{Status: dkg.Certified}

	// Send back the public DKG key
	distKey, err := h.dkg.DistKeyShare()
	if err != nil {
		dela.Logger.Error().Msgf("failed to get distr key: %v", err)
		return
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
		dela.Logger.Error().Msgf("got an error while sending pub key: %v", err)
		return
	}
	h.saveState(h)
}

func (h *Handler) deal(out mino.Sender) error {
	// Send my Deals to the other nodes. Note that we take an optimistic
	// approach and don't check if the deals are correctly sent to the node. The
	// DKG setup needs a full connectivity anyway, and for the moment everything
	// fails if this assumption breaks.

	deals, err := h.dkg.Deals()
	if err != nil {
		return xerrors.Errorf("failed to compute the deals: %v", err)
	}

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

		to := h.startRes.participants[i]

		h.log.Info().Str("to", to.String()).Msg("send deal")

		out.Send(dealMsg, to)
	}

	return nil
}

func (h *Handler) respond(deals *list.List, out mino.Sender) {
	numReceivedDeals := 0

	for numReceivedDeals < len(h.startRes.participants)-1 {
		h.Lock()
		deal := deals.Front()
		if deal != nil {
			deals.Remove(deal)
		}
		h.Unlock()

		if deal == nil {
			time.Sleep(retryTimeout)
			continue
		}

		err := h.handleDeal(deal.Value.(types.Deal), out)
		if err != nil {
			h.log.Warn().Msgf("failed to handle received deal: %v", err)
		}

		numReceivedDeals++
	}
}

func (h *Handler) certify(resps *list.List, out mino.Sender) error {

	for !h.dkg.Certified() {
		h.Lock()
		resp := resps.Front()
		if resp != nil {
			resps.Remove(resp)
		}
		h.Unlock()

		if resp == nil {
			time.Sleep(retryTimeout)
			continue
		}

		_, err := h.dkg.ProcessResponse(resp.Value.(*pedersen.Response))
		if err != nil {
			h.log.Warn().Msgf("%s failed to process response: %v", h.me, err)
		}
	}

	return nil
}

// handleDeal process the Deal and send the responses to the other nodes.
func (h *Handler) handleDeal(msg types.Deal, out mino.Sender) error {

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
		return xerrors.Errorf("failed to process deal: %v", err)
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

	for _, addr := range h.startRes.participants {
		if addr.Equal(h.me) {
			continue
		}

		errs := out.Send(resp, addr)

		err = <-errs
		if err != nil {
			return xerrors.Errorf("failed to send response to '%s': %v", addr, err)
		}
	}

	return nil
}

// handleDecryptRequest computes the public shares of a form and sends them
// to the chain to allow decryption to proceed.
func (h *Handler) handleDecryptRequest(formID string) error {
	shuffleInstances, err := h.getShuffleIfValid(formID)
	if err != nil {
		return xerrors.Errorf("failed to check if the shuffle is over: %v", err)
	}

	numberOfShuffles := len(shuffleInstances)
	numberOfBallots := len(shuffleInstances[numberOfShuffles-1].ShuffledBallots)
	publicShares := make([][]etypes.Pubshare, numberOfBallots)

	h.RLock()

	for i, ballot := range shuffleInstances[numberOfShuffles-1].ShuffledBallots {
		ballotShares := make([]etypes.Pubshare, len(ballot))

		for j, ciphertext := range ballot {
			S := suite.Point().Mul(h.privShare.V, ciphertext.K)

			partialVal := suite.Point().Sub(ciphertext.C, S)

			ballotShares[j] = partialVal
		}

		publicShares[i] = ballotShares
	}

	h.RUnlock()

	err = h.txmnger.Sync()
	if err != nil {
		return xerrors.Errorf("failed to sync manager: %v", err)
	}

	// loop until our transaction has been accepted, or enough nodes submitted
	// their pubShares
	for {
		form, err := etypes.FormFromStore(h.context, h.formFac, formID, h.service.GetStore())
		if err != nil {
			return xerrors.Errorf("could not get the form: %v", err)
		}

		//TODO: Works with current "shuffleThreshold", but the shuffle threshold
		// should be smaller in theory ? (1/3 + 1 vs 2/3 + 1 ? )
		nbrSubmissions := len(form.PubsharesUnits.Pubshares)

		if nbrSubmissions >= form.ShuffleThreshold {
			dela.Logger.Info().Msgf("decryption possible with shares from %d nodes",
				nbrSubmissions)
			return nil
		}

		tx, err := makeTx(h.context, &form, publicShares, h.privShare.I,
			h.txmnger, h.pubSharesSigner)

		if err != nil {
			return xerrors.Errorf("failed to make tx: %v", err)
		}

		// TODO: Define in term of size of form ? (same in shuffle)
		watchTimeout := 4 + rand.Intn(form.ShuffleThreshold)
		watchCtx, cancel := context.WithTimeout(context.Background(), time.Duration(watchTimeout)*time.Second)
		defer cancel()

		events := h.service.Watch(watchCtx)

		err = h.pool.Add(tx)
		if err != nil {
			return xerrors.Errorf("failed to add transaction to the pool: %v", err)
		}

		accepted, msg := watchTx(events, tx.GetID())

		if accepted {
			dela.Logger.Info().Msgf("pubShares accepted on the chain (index: %d)", h.privShare.I)
			return nil
		}

		err = h.txmnger.Sync()
		if err != nil {
			return xerrors.Errorf("failed to sync manager: %v", err)
		}

		dela.Logger.Info().Msgf("submission of pubShares denied: %s", msg)

		cancel()
	}
}

// getShuffleIfValid allows checking if enough shuffles have been made on the
// ballots.
func (h *Handler) getShuffleIfValid(formID string) ([]etypes.ShuffleInstance, error) {
	form, err := etypes.FormFromStore(h.context, h.formFac, formID, h.service.GetStore())
	if err != nil {
		return nil, xerrors.Errorf("could not get the form: %v", err)
	}

	if len(form.ShuffleInstances) == 0 {
		return nil, xerrors.New("form has no shuffles")
	}

	if form.Status != etypes.ShuffledBallots {
		return nil, xerrors.New("ballots have not been shuffled")
	}

	return form.ShuffleInstances, nil
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

	ret, err := json.Marshal(&struct {
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

	return ret, err
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
	err = hd.StartRes.UnmarshalJSON(aux.StartRes)
	if err != nil {
		return err
	}

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
		err = privShareV.UnmarshalBinary(privShareBuf.V)
		if err != nil {
			return err
		}
		privShare := &share.PriShare{
			I: privShareBuf.I,
			V: privShareV,
		}
		hd.PrivShare = privShare
	}

	// Unmarshal PubKey
	pubKey := suite.Point()
	err = pubKey.UnmarshalBinary(aux.PubKey)
	if err != nil {
		return err
	}
	hd.PubKey = pubKey

	// Unmarshal PrivKey
	privKey := suite.Scalar()
	err = privKey.UnmarshalBinary(aux.PrivKey)
	if err != nil {
		return err
	}
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

	ret, err := json.Marshal(&struct {
		DistKey      []byte   `json:",omitempty"`
		Participants [][]byte `json:",omitempty"`
	}{
		DistKey:      distKeyBuf,
		Participants: participantsBuf,
	})

	return ret, err
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
		// TODO: use addressFactory here
		f := session.AddressFactory{}
		var participants = make([]mino.Address, len(aux.Participants))
		for i, partStr := range aux.Participants {
			participants[i] = f.FromText(partStr)
		}
		s.SetParticipants(participants)
	} else {
		s.SetParticipants(nil)
	}

	return nil
}

// watchTx checks the transaction to find one that match txID. Returns if the
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

func makeTx(ctx serde.Context, form *etypes.Form, pubShares etypes.PubsharesUnit,
	index int,
	manager txn.Manager,
	pubSharesSigner crypto.Signer) (txn.Transaction, error) {

	pubShareTx := etypes.RegisterPubShares{
		FormID:    form.FormID,
		Pubshares: pubShares,
		Index:     index,
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
		return nil, xerrors.Errorf("could not sign the pubShares : %v", err)
	}

	pubKey, err := pubSharesSigner.GetPublicKey().MarshalBinary()
	if err != nil {
		return nil, xerrors.Errorf("could not marshal signer's public key: %v", err)
	}

	encodedSignature, err := signature.Serialize(jsondela.NewContext())
	if err != nil {
		return nil, xerrors.Errorf("Could not encode signature as []byte : %v ", err)
	}

	// Complete transaction:
	pubShareTx.Signature = encodedSignature
	pubShareTx.PublicKey = pubKey

	data, err := pubShareTx.Serialize(ctx)
	if err != nil {
		return nil, xerrors.Errorf("failed to serialize register pubShares: %v", err)
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
		Key:   evoting.FormArg,
		Value: data,
	}

	tx, err := manager.Make(args...)
	if err != nil {
		return nil, xerrors.Errorf("failed to use manager: %v", err.Error())
	}

	return tx, nil
}
