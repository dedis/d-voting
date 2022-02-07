package pedersen

import (
	"encoding/hex"
	"sync"
	"time"

	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"

	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	ctypes "go.dedis.ch/dela/core/ordering/cosipbft/types"

	"github.com/dedis/d-voting/internal/tracing"
	"github.com/dedis/d-voting/services/dkg"
	_ "github.com/dedis/d-voting/services/dkg/pedersen/json"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	jsonserde "go.dedis.ch/dela/serde/json"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/net/context"
	"golang.org/x/xerrors"

	// Register the JSON format for the election
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
	RPC_NAME       = "dkgevoting"
)

// Pedersen allows one to initialize a new DKG protocol.
//
// - implements dkg.DKG
type Pedersen struct {
	sync.RWMutex

	mino      mino.Mino
	factory   serde.Factory
	service   ordering.Service
	rosterFac authority.Factory
	actors    map[string]dkg.Actor
}

// NewPedersen returns a new DKG Pedersen factory
func NewPedersen(m mino.Mino, service ordering.Service,
	rosterFac authority.Factory) *Pedersen {

	factory := types.NewMessageFactory(m.GetAddressFactory())
	actors := make(map[string]dkg.Actor)

	return &Pedersen{
		mino:      m,
		factory:   factory,
		service:   service,
		rosterFac: rosterFac,
		actors:    actors,
	}
}

// Listen implements dkg.DKG. It must be called on each node that participates
// in the DKG.
func (s *Pedersen) Listen(electionIDBuf []byte) (dkg.Actor, error) {

	electionID := hex.EncodeToString(electionIDBuf)

	_, exists := electionExists(s.service, electionIDBuf)
	if !exists {
		return nil, xerrors.Errorf("election %s was not found", electionID)
	}

	actor, exists := s.GetActor(electionIDBuf)
	if exists {
		return actor, xerrors.Errorf("actor already exists for electionID %s", electionID)
	}

	return s.NewActor(electionIDBuf, NewHandlerData())
}

// NewActor initializes a dkg.Actor with an RPC specific to the election with
// the given keypair
func (s *Pedersen) NewActor(electionIDBuf []byte, handlerData HandlerData) (dkg.Actor, error) {

	// hex-encoded string
	electionID := hex.EncodeToString(electionIDBuf)

	ctx := jsonserde.NewContext()
	ctx = serde.WithFactory(ctx, etypes.ElectionKey{}, etypes.ElectionFactory{})
	ctx = serde.WithFactory(ctx, ctypes.RosterKey{}, s.rosterFac)

	// link the actor to an RPC by the election ID
	h := NewHandler(s.mino.GetAddress(), s.service, handlerData, ctx)
	no := s.mino.WithSegment(electionID)
	rpc := mino.MustCreateRPC(no, RPC_NAME, h, s.factory)

	a := &Actor{
		rpc:        rpc,
		factory:    s.factory,
		service:    s.service,
		context:    ctx,
		handler:    h,
		electionID: electionID,
	}

	s.Lock()
	defer s.Unlock()
	s.actors[electionID] = a

	return a, nil
}

func (s *Pedersen) GetActor(electionIDBuf []byte) (dkg.Actor, bool) {
	s.RLock()
	defer s.RUnlock()
	actor, exists := s.actors[hex.EncodeToString(electionIDBuf)]
	return actor, exists
}

// Actor allows one to perform DKG operations like encrypt/decrypt a message
//
// - implements dkg.Actor
type Actor struct {
	rpc        mino.RPC
	factory    serde.Factory
	service    ordering.Service
	context    serde.Context
	handler    *Handler
	electionID string
}

// Setup implements dkg.Actor. It initializes the DKG protocol
// across all participating nodes.
func (a *Actor) Setup() (kyber.Point, error) {

	if a.handler.startRes.Done() {
		return nil, xerrors.Errorf("setup() was already called, only one call is allowed")
	}

	election, err := getElection(a.context, a.electionID, a.service)
	if err != nil {
		return nil, xerrors.Errorf("failed to get election: %v", err)
	}

	fac := a.context.GetFactory(ctypes.RosterKey{})
	rosterFac, ok := fac.(authority.Factory)
	if !ok {
		return nil, xerrors.Errorf("failed to get roster factory: %T", fac)
	}

	roster, err := rosterFac.AuthorityOf(a.context, election.RosterBuf)
	if err != nil {
		return nil, xerrors.Errorf("failed to deserialize roster: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, tracing.ProtocolKey, protocolNameSetup)

	sender, receiver, err := a.rpc.Stream(ctx, roster)
	if err != nil {
		return nil, xerrors.Errorf("failed to stream: %v", err)
	}

	addrs := make([]mino.Address, 0, roster.Len())
	addrIter := roster.AddressIterator()
	for addrIter.HasNext() {
		addrs = append(addrs, addrIter.GetNext())
	}

	// get the peer DKG pub keys
	getPeerKey := types.NewGetPeerPubKey()
	errs := sender.Send(getPeerKey, addrs...)
	err = <-errs
	if err != nil {
		return nil, xerrors.Errorf("failed to send getPeerKey message: %v", err)
	}

	lenAddrs := len(addrs)
	dkgPeerPubkeys := make([]kyber.Point, 0, lenAddrs)
	associatedAddrs := make([]mino.Address, 0, lenAddrs)

	if lenAddrs == 0 {
		return nil, xerrors.Errorf("the list of addresses is empty")
	}

	for i := 0; i < lenAddrs; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		from, msg, err := receiver.Recv(ctx)
		if err != nil {
			return nil, xerrors.Errorf("failed to receive peer pubkey: %v", err)
		}

		dela.Logger.Info().Msgf("received a response from %v", from)

		resp, ok := msg.(types.GetPeerPubKeyResp)
		if !ok {
			return nil, xerrors.Errorf("received an unexpected message: %T - %s", resp, resp)
		}

		dkgPeerPubkeys = append(dkgPeerPubkeys, resp.GetPublicKey())
		associatedAddrs = append(associatedAddrs, from)

		dela.Logger.Info().Msgf("Public key: %s", resp.GetPublicKey().String())
	}

	message := types.NewStart(associatedAddrs, dkgPeerPubkeys)

	errs = sender.Send(message, addrs...)
	err = <-errs
	if err != nil {
		return nil, xerrors.Errorf("failed to send start: %v", err)
	}

	dkgPubKeys := make([]kyber.Point, lenAddrs)

	for i := 0; i < lenAddrs; i++ {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		addr, msg, err := receiver.Recv(ctx)
		if err != nil {
			return nil, xerrors.Errorf("got an error from '%s' while "+
				"receiving: %v", addr, err)
		}

		doneMsg, ok := msg.(types.StartDone)
		if !ok {
			return nil, xerrors.Errorf("expected to receive a Done message, but "+
				"go the following: %T", msg)
		}

		dkgPubKeys[i] = doneMsg.GetPublicKey()

		// this is a simple check that every node sends back the same DKG pub
		// key.
		// TODO: handle the situation where a pub key is not the same
		if i != 0 && !dkgPubKeys[i-1].Equal(doneMsg.GetPublicKey()) {
			return nil, xerrors.Errorf("the public keys do not match: %v", dkgPubKeys)
		}
	}

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

// Decrypt implements dkg.Actor. It gets the private shares of the nodes and
// decrypts the message.
// TODO: perform a re-encryption instead of gathering the private shares, which
// should never happen.
func (a *Actor) Decrypt(K, C kyber.Point) ([]byte, error) {

	if !a.handler.startRes.Done() {
		return nil, xerrors.Errorf("setup() was not called")
	}

	players := mino.NewAddresses(a.handler.startRes.GetParticipants()...)

	ctx, cancel := context.WithTimeout(context.Background(), decryptTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, tracing.ProtocolKey, protocolNameDecrypt)

	sender, receiver, err := a.rpc.Stream(ctx, players)
	if err != nil {
		return nil, xerrors.Errorf("failed to create stream: %v", err)
	}

	iterator := players.AddressIterator()
	addrs := make([]mino.Address, 0, players.Len())
	for iterator.HasNext() {
		addrs = append(addrs, iterator.GetNext())
	}

	message := types.NewDecryptRequest(K, C, a.electionID)

	err = <-sender.Send(message, addrs...)
	if err != nil {
		return nil, xerrors.Errorf("failed to send decrypt request: %v", err)
	}

	pubShares := make([]*share.PubShare, len(addrs))

	for i := 0; i < len(addrs); i++ {
		_, message, err := receiver.Recv(ctx)
		if err != nil {
			return []byte{}, xerrors.Errorf("stream stopped unexpectedly: %v", err)
		}

		decryptReply, ok := message.(types.DecryptReply)
		if !ok {
			return []byte{}, xerrors.Errorf("got unexpected reply, expected "+
				"%T but got: %T", decryptReply, message)
		}

		pubShares[i] = &share.PubShare{
			I: int(decryptReply.I),
			V: decryptReply.V,
		}
	}

	res, err := share.RecoverCommit(suite, pubShares, len(addrs), len(addrs))
	if err != nil {
		return []byte{}, xerrors.Errorf("failed to recover commit: %v", err)
	}

	decryptedMessage, err := res.Data()
	if err != nil {
		return []byte{}, xerrors.Errorf("failed to get embedded data: %v", err)
	}

	return decryptedMessage, nil
}

// Reshare implements dkg.Actor. It recreates the DKG with an updated list of
// participants.
// TODO: to do
func (a *Actor) Reshare() error {
	return nil
}

// MarshalJSON implements dkg.Actor. It exports the data relevant to an Actor
// that is meant to be persistent.
func (a *Actor) MarshalJSON() ([]byte, error) {
	return a.handler.MarshalJSON()
}

func electionExists(service ordering.Service, electionIDBuf []byte) (ordering.Proof, bool) {
	proof, err := service.GetProof(electionIDBuf)
	if err != nil {
		return proof, false
	}

	// this is proof of absence
	if string(proof.GetValue()) == "" {
		return proof, false
	}

	return proof, true
}

// getElection gets the election from the service.
func getElection(ctx serde.Context, electionIDHex string, srv ordering.Service) (etypes.Election, error) {
	var election etypes.Election

	electionID, err := hex.DecodeString(electionIDHex)
	if err != nil {
		return election, xerrors.Errorf("failed to decode electionIDHex: %v", err)
	}

	proof, err := srv.GetProof(electionID)
	if err != nil {
		return election, xerrors.Errorf("failed to get proof: %v", err)
	}

	electionBuff := proof.GetValue()
	if len(electionBuff) == 0 {
		return election, xerrors.Errorf("election does not exist")
	}

	fac := ctx.GetFactory(etypes.ElectionKey{})
	if fac == nil {
		return election, xerrors.New("election factory not found")
	}

	message, err := fac.Deserialize(ctx, electionBuff)
	if err != nil {
		return election, xerrors.Errorf("failed to deserialize Election: %v", err)
	}

	election, ok := message.(etypes.Election)
	if !ok {
		return election, xerrors.Errorf("wrong message type: %T", message)
	}

	if electionIDHex != election.ElectionID {
		return election, xerrors.Errorf("electionID do not match: %q != %q",
			electionIDHex, election.ElectionID)
	}

	return election, nil
}
