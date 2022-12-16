package pedersen

import (
	"encoding/hex"
	"sync"
	"time"

	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/crypto"

	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"

	"github.com/dedis/d-voting/contracts/evoting"
	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/rs/zerolog"

	"github.com/dedis/d-voting/internal/tracing"
	"github.com/dedis/d-voting/services/dkg"

	// Register the JSON types for Pedersen
	_ "github.com/dedis/d-voting/services/dkg/pedersen/json"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	jsonserde "go.dedis.ch/dela/serde/json"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/net/context"
	"golang.org/x/xerrors"

	// Register the JSON format for the form
	_ "github.com/dedis/d-voting/contracts/evoting/json"
)

// suite is the Kyber suite for Pedersen.
var suite = suites.MustFind("Ed25519")

var (
	// protocolNameSetup denotes the value of the protocol span tag
	// associated with the `dkg-setup` protocol.
	protocolNameSetup = "dkg-setup"
	// protocolNameDecrypt denotes the value of the protocol span tag
	// associated with the `dkg-decrypt` protocol.
	protocolNameDecrypt = "dkg-decrypt"
)

const (
	setupTimeout   = time.Second * 300
	decryptTimeout = time.Second * 100

	// RPC defines the RPC name used for mino
	RPC = "dkgevoting"
)

// Pedersen allows one to initialize a new DKG protocol.
//
// - implements dkg.DKG
type Pedersen struct {
	sync.RWMutex

	mino    mino.Mino
	factory serde.Factory
	service ordering.Service
	formFac serde.Factory
	pool    pool.Pool
	signer  crypto.Signer
	actors  map[string]dkg.Actor
}

// NewPedersen returns a new DKG Pedersen factory
func NewPedersen(m mino.Mino, service ordering.Service, pool pool.Pool,
	formFac serde.Factory, signer crypto.Signer) *Pedersen {

	factory := types.NewMessageFactory(m.GetAddressFactory())
	actors := make(map[string]dkg.Actor)

	return &Pedersen{
		mino:    m,
		factory: factory,
		service: service,
		pool:    pool,
		actors:  actors,
		signer:  signer,
		formFac: formFac,
	}
}

// Listen implements dkg.DKG. It must be called on each node that participates
// in the DKG.
func (s *Pedersen) Listen(formIDBuf []byte, txmngr txn.Manager) (dkg.Actor, error) {

	formID := hex.EncodeToString(formIDBuf)

	_, exists := formExists(s.service, formIDBuf)
	if !exists {
		return nil, xerrors.Errorf("form %s was not found", formID)
	}

	actor, exists := s.GetActor(formIDBuf)
	if exists {
		return actor, xerrors.Errorf("actor already exists for formID %s", formID)
	}

	return s.NewActor(formIDBuf, s.pool, txmngr, NewHandlerData())
}

// NewActor initializes a dkg.Actor with an RPC specific to the form with
// the given keypair
func (s *Pedersen) NewActor(formIDBuf []byte, pool pool.Pool, txmngr txn.Manager,
	handlerData HandlerData) (dkg.Actor,
	error) {

	// hex-encoded string
	formID := hex.EncodeToString(formIDBuf)

	ctx := jsonserde.NewContext()

	status := &dkg.Status{Status: dkg.Initialized}

	// link the actor to an RPC by the form ID
	h := NewHandler(s.mino.GetAddress(), s.service, pool, txmngr, s.signer,
		handlerData, ctx, s.formFac, status)

	no := s.mino.WithSegment(formID)
	rpc := mino.MustCreateRPC(no, RPC, h, s.factory)

	log := dela.Logger.With().Str("role", "DKG actor").Logger()

	a := &Actor{
		rpc:     rpc,
		factory: s.factory,
		service: s.service,
		context: ctx,
		formFac: s.formFac,
		handler: h,
		formID:  formID,
		status:  status,
		log:     log,
	}

	evoting.PromFormDkgStatus.WithLabelValues(formID).Set(float64(dkg.Initialized))

	s.Lock()
	defer s.Unlock()
	s.actors[formID] = a

	return a, nil
}

// GetActor implements dkg.DKG
func (s *Pedersen) GetActor(formIDBuf []byte) (dkg.Actor, bool) {
	s.RLock()
	defer s.RUnlock()
	actor, exists := s.actors[hex.EncodeToString(formIDBuf)]
	return actor, exists
}

// Actor allows one to perform DKG operations like encrypt/decrypt a message
//
// - implements dkg.Actor
type Actor struct {
	rpc     mino.RPC
	factory serde.Factory
	service ordering.Service
	context serde.Context
	formFac serde.Factory
	handler *Handler
	formID  string
	status  *dkg.Status
	log     zerolog.Logger
}

func (a *Actor) setErr(err error, args map[string]interface{}) {
	*a.status = dkg.Status{
		Status: dkg.Failed,
		Err:    err,
		Args:   args,
	}

	evoting.PromFormDkgStatus.WithLabelValues(a.formID).Set(float64(dkg.Failed))
}

// Setup implements dkg.Actor. It initializes the DKG protocol across all
// participating nodes. This function updates the actor's status in case of
// error to allow asynchronous call of this function.
func (a *Actor) Setup() (kyber.Point, error) {
	a.log.Info().Msg("setup")

	if a.handler.startRes.Done() {
		err := xerrors.New("setup() was already called, only one call is allowed")
		a.setErr(err, nil)
		return nil, err
	}

	form, err := a.getForm()
	if err != nil {
		err := xerrors.Errorf("failed to get form: %v", err)
		a.setErr(err, nil)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, tracing.ProtocolKey, protocolNameSetup)

	sender, receiver, err := a.rpc.Stream(ctx, form.Roster)
	if err != nil {
		err := xerrors.Errorf("failed to stream: %v", err)
		a.setErr(err, nil)
		return nil, err
	}

	if form.Roster.Len() == 0 {
		err := xerrors.Errorf("the roster is empty")
		a.setErr(err, nil)
		return nil, err
	}

	addrs := make([]mino.Address, 0, form.Roster.Len())
	addrIter := form.Roster.AddressIterator()
	for addrIter.HasNext() {
		addrs = append(addrs, addrIter.GetNext())
	}

	a.log.Info().Msgf("sending getkey request to %v", addrs)

	// get the peer DKG pub keys
	getPeerKey := types.NewGetPeerPubKey()
	errs := sender.Send(getPeerKey, addrs...)

	err = <-errs
	if err != nil {
		err := xerrors.Errorf("failed to send getPeerKey message: %v", err)
		a.setErr(err, nil)
		return nil, err
	}

	lenAddrs := len(addrs)
	dkgPeerPubkeys := make([]kyber.Point, 0, lenAddrs)
	associatedAddrs := make([]mino.Address, 0, lenAddrs)

	if lenAddrs == 0 {
		err := xerrors.Errorf("the list of addresses is empty")
		a.setErr(err, nil)
		return nil, err
	}

	for i := 0; i < lenAddrs; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		from, msg, err := receiver.Recv(ctx)
		if err != nil {
			err := xerrors.Errorf("failed to receive peer pubkey: %v", err)
			a.setErr(err, nil)
			return nil, err
		}

		dela.Logger.Info().Msgf("received a response from %v", from)

		resp, ok := msg.(types.GetPeerPubKeyResp)
		if !ok {
			err := xerrors.Errorf("received an unexpected message: %T - %s", resp, resp)
			a.setErr(err, nil)
			return nil, err
		}

		dkgPeerPubkeys = append(dkgPeerPubkeys, resp.GetPublicKey())
		associatedAddrs = append(associatedAddrs, from)

		dela.Logger.Info().Msgf("Public key: %s", resp.GetPublicKey().String())
	}

	message := types.NewStart(associatedAddrs, dkgPeerPubkeys)

	a.log.Info().Msgf("sending start to %s", addrs)

	errs = sender.Send(message, addrs...)
	err = <-errs
	if err != nil {
		err := xerrors.Errorf("failed to send start: %v", err)
		a.setErr(err, nil)
		return nil, err
	}

	dkgPubKeys := make([]kyber.Point, lenAddrs)

	for i := 0; i < lenAddrs; i++ {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
		defer cancel()

		addr, msg, err := receiver.Recv(ctx)
		if err != nil {
			err := xerrors.Errorf("got an error from '%s' while receiving: %v", addr, err)
			a.setErr(err, nil)
			return nil, err
		}

		doneMsg, ok := msg.(types.StartDone)
		if !ok {
			err := xerrors.Errorf("expected to receive a Done message, but "+
				"go the following: %T", msg)
			a.setErr(err, nil)
			return nil, err
		}

		dkgPubKeys[i] = doneMsg.GetPublicKey()

		// this is a simple check that every node sends back the same DKG pub
		// key.
		if i != 0 && !dkgPubKeys[i-1].Equal(doneMsg.GetPublicKey()) {
			err := xerrors.Errorf("the public keys do not match: %v", dkgPubKeys)
			a.setErr(err, nil)
			return nil, err
		}

		a.log.Info().Msgf("ok for %s", addr.String())
	}

	*a.status = dkg.Status{Status: dkg.Setup}
	evoting.PromFormDkgStatus.WithLabelValues(a.formID).Set(float64(dkg.Setup))

	return dkgPubKeys[0], nil
}

// GetPublicKey implements dkg.Actor
func (a *Actor) GetPublicKey() (kyber.Point, error) {
	if !a.handler.startRes.Done() {
		return nil, xerrors.Errorf("dkg has not been initialized")
	}

	return a.handler.startRes.GetDistKey(), nil
}

// Encrypt implements dkg.Actor. It uses the DKG public key to encrypt a
// message.
func (a *Actor) Encrypt(message []byte) (K, C kyber.Point, remainder []byte,
	err error) {

	if !a.handler.startRes.Done() {
		return nil, nil, nil, xerrors.Errorf("setup() was not called")
	}

	// Embed the message (or as much of it as will fit) into a curve point.
	M := suite.Point().Embed(message, random.New())
	max := suite.Point().EmbedLen()
	if max > len(message) {
		max = len(message)
	}
	remainder = message[max:]
	// ElGamal-encrypt the point to produce ciphertext (K,C).
	k := suite.Scalar().Pick(random.New())                     // ephemeral private key
	K = suite.Point().Mul(k, nil)                              // ephemeral DH public key
	S := suite.Point().Mul(k, a.handler.startRes.GetDistKey()) // ephemeral DH shared secret
	C = S.Add(S, M)                                            // message blinded with secret

	return K, C, remainder, nil
}

// ComputePubshares implements dkg.Actor. It sends a decrypt request to all
// the nodes taking part.
func (a *Actor) ComputePubshares() error {

	if !a.handler.startRes.Done() {
		return xerrors.Errorf("setup() was not called")
	}

	players := mino.NewAddresses(a.handler.startRes.GetParticipants()...)
	if players.Len() == 0 {
		return xerrors.Errorf("the list of Participants is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), decryptTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, tracing.ProtocolKey, protocolNameDecrypt)

	sender, _, err := a.rpc.Stream(ctx, players)
	if err != nil {
		return xerrors.Errorf("failed to create stream: %v", err)
	}

	iterator := players.AddressIterator()
	addrs := make([]mino.Address, 0, players.Len())
	for iterator.HasNext() {
		addrs = append(addrs, iterator.GetNext())
	}

	message := types.NewDecryptRequest(a.formID)

	err = <-sender.Send(message, addrs...)
	if err != nil {
		dela.Logger.Warn().Msgf("failed to send decrypt request: %v", err)
		//return xerrors.Errorf("failed to send decrypt request: %v", err)
	}

	return nil
}

// MarshalJSON implements dkg.Actor. It exports the data relevant to an Actor
// that is meant to be persistent.
func (a *Actor) MarshalJSON() ([]byte, error) {
	return a.handler.MarshalJSON()
}

// Status implements dkg.Actor
func (a *Actor) Status() dkg.Status {
	return *a.status
}

func formExists(service ordering.Service, formIDBuf []byte) (ordering.Proof, bool) {
	proof, err := service.GetProof(formIDBuf)
	if err != nil {
		return proof, false
	}

	// this is proof of absence
	if string(proof.GetValue()) == "" {
		return proof, false
	}

	return proof, true
}

// getForm gets the form from the service.
func (a Actor) getForm() (etypes.Form, error) {
	var form etypes.Form

	formID, err := hex.DecodeString(a.formID)
	if err != nil {
		return form, xerrors.Errorf("failed to decode formIDHex: %v", err)
	}

	proof, exists := formExists(a.service, formID)
	if !exists {
		return form, xerrors.Errorf("form does not exist: %v", err)
	}

	message, err := a.formFac.Deserialize(a.context, proof.GetValue())
	if err != nil {
		return form, xerrors.Errorf("failed to deserialize Form: %v", err)
	}

	form, ok := message.(etypes.Form)
	if !ok {
		return form, xerrors.Errorf("wrong message type: %T", message)
	}

	if a.formID != form.FormID {
		return form, xerrors.Errorf("formID do not match: %q != %q",
			a.formID, form.FormID)
	}

	return form, nil
}
