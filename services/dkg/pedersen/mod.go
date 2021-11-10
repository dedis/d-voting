package pedersen

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"time"

	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"

	electionTypes "github.com/dedis/d-voting/contracts/evoting/types"

	"github.com/dedis/d-voting/internal/tracing"
	"github.com/dedis/d-voting/services/dkg"
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
)

// suite is the Kyber suite for Pedersen.
var suite = suites.MustFind("Ed25519")

var (
	// protocolNameSetup denotes the value of the protocol span tag associated
	// with the `dkg-setup` protocol.
	protocolNameSetup = "dkg-setup"
	// protocolNameDecrypt denotes the value of the protocol span tag
	// associated with the `dkg-decrypt` protocol.
	protocolNameDecrypt = "dkg-decrypt"
)

const (
	setupTimeout   = time.Second * 300
	decryptTimeout = time.Second * 100
)

// Pedersen allows one to initialize a new DKG protocol.
//
// - implements dkg.DKG
type Pedersen struct {
	privKey kyber.Scalar
	mino    mino.Mino
	factory serde.Factory
	actor   dkg.Actor
	evoting bool

	// set be the SetMissingStuff()
	service   ordering.Service
	rosterFac authority.Factory

	pubkey kyber.Point
}

// NewPedersen returns a new DKG Pedersen factory
func NewPedersen(m mino.Mino, evoting bool, service ordering.Service,
	rosterFac authority.Factory) (*Pedersen, kyber.Point) {

	factory := types.NewMessageFactory(m.GetAddressFactory())

	privkey := suite.Scalar().Pick(suite.RandomStream())
	pubkey := suite.Point().Mul(privkey, nil)

	return &Pedersen{
		privKey:   privkey,
		mino:      m,
		factory:   factory,
		evoting:   evoting,
		service:   service,
		rosterFac: rosterFac,
		pubkey:    pubkey,
	}, pubkey
}

// Listen implements dkg.DKG. It must be called on each node that participates
// in the DKG. Creates the RPC.
// TODO: Listen is not a good name; it's when you open the stream that you listen
// Something like NewActor
func (s *Pedersen) Listen(electionID []byte) (dkg.Actor, error) {
	h := NewHandler(s.privKey, s.mino.GetAddress(), s.service, s.evoting, s.pubkey)

	// Link the actor to an RPC by the election ID
	no := s.mino.WithSegment(hex.EncodeToString(electionID))
	rpc := mino.MustCreateRPC(no, "dkgevoting", h, s.factory) 

	a := &Actor{
		rpc:       rpc,
		factory:   s.factory,
		startRes:  h.startRes,
		service:   s.service,
		rosterFac: s.rosterFac,
		context:   jsonserde.NewContext(),
	}

	s.actor = a

	return a, nil
}

// GetLastActor ...
func (s *Pedersen) GetLastActor() (dkg.Actor, error) {
	if s.actor != nil {
		return s.actor, nil
	} else {
		return nil, xerrors.Errorf("listen has not been called")
	}
}

// Actor allows one to perform DKG operations like encrypt/decrypt a message
//
// - implements dkg.Actor
type Actor struct {
	rpc       mino.RPC
	factory   serde.Factory
	startRes  *state
	service   ordering.Service
	rosterFac authority.Factory
	context   serde.Context
}

// Setup implement dkg.Actor. It initializes the DKG.
func (a *Actor) Setup(electionID []byte) (kyber.Point, error) {
	if a.startRes.Done() {
		return nil, xerrors.Errorf("startRes is already done, only one setup call is allowed")
	}

	proof, err := a.service.GetProof(electionID)
	if err != nil {
		return nil, xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election := new(electionTypes.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal Election: %v", err)
	}

	roster, err := a.rosterFac.AuthorityOf(a.context, election.RosterBuf)
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
			return nil, xerrors.Errorf("the public keys does not match: %v", dkgPubKeys)
		}
	}

	return dkgPubKeys[0], nil
}

// GetPublicKey implements dkg.Actor
func (a *Actor) GetPublicKey() (kyber.Point, error) {
	if !a.startRes.Done() {
		return nil, xerrors.Errorf("DKG has not been initialized")
	}

	return a.startRes.GetDistKey(), nil
}

// Encrypt implements dkg.Actor. It uses the DKG public key to encrypt a
// message.
func (a *Actor) Encrypt(message []byte) (K, C kyber.Point, remainder []byte,
	err error) {

	if !a.startRes.Done() {
		return nil, nil, nil, xerrors.Errorf("you must first initialize DKG. " +
			"Did you call setup() first?")
	}

	// Embed the message (or as much of it as will fit) into a curve point.
	M := suite.Point().Embed(message, random.New())
	max := suite.Point().EmbedLen()
	if max > len(message) {
		max = len(message)
	}
	remainder = message[max:]
	// ElGamal-encrypt the point to produce ciphertext (K,C).
	k := suite.Scalar().Pick(random.New())             // ephemeral private key
	K = suite.Point().Mul(k, nil)                      // ephemeral DH public key
	S := suite.Point().Mul(k, a.startRes.GetDistKey()) // ephemeral DH shared secret
	C = S.Add(S, M)                                    // message blinded with secret

	return K, C, remainder, nil
}

// Decrypt implements dkg.Actor. It gets the private shares of the nodes and
// decrypt the  message.
// TODO: perform a re-encryption instead of gathering the private shares, which
// should never happen.
func (a *Actor) Decrypt(K, C kyber.Point, electionID []byte) ([]byte, error) {

	if !a.startRes.Done() {
		return nil, xerrors.Errorf("you must first initialize DKG. " +
			"Did you call setup() first?")
	}

	players := mino.NewAddresses(a.startRes.GetParticipants()...)

	ctx, cancel := context.WithTimeout(context.Background(), decryptTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, tracing.ProtocolKey, protocolNameDecrypt)

	sender, receiver, err := a.rpc.Stream(ctx, players)
	if err != nil {
		return nil, xerrors.Errorf("failed to create stream: %v", err)
	}

	players = mino.NewAddresses(a.startRes.GetParticipants()...)
	iterator := players.AddressIterator()

	addrs := make([]mino.Address, 0, players.Len())
	for iterator.HasNext() {
		addrs = append(addrs, iterator.GetNext())
	}

	message := types.NewDecryptRequest(K, C, hex.EncodeToString(electionID))

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
