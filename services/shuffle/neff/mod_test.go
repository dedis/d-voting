package neff

import (
	"encoding/hex"
	"strconv"
	"testing"

	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"

	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/shuffle/neff/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	shuffleKyber "go.dedis.ch/kyber/v3/shuffle"
	"go.dedis.ch/kyber/v3/util/random"

	ctypes "go.dedis.ch/dela/core/ordering/cosipbft/types"
	"go.dedis.ch/dela/serde/json"
)

var serdecontext = serde.WithFactory(serde.WithFactory(json.NewContext(), etypes.ElectionKey{},
	etypes.ElectionFactory{}), ctypes.RosterKey{}, fake.Factory{})

func TestNeffShuffle_Listen(t *testing.T) {

	NeffShuffle := NewNeffShuffle(fake.Mino{}, &fake.Service{}, &fake.Pool{}, nil, fakeAuthorityFactory{}, fake.NewSigner())

	actor, err := NeffShuffle.Listen(fakeManager{})
	require.NoError(t, err)

	require.NotNil(t, actor)
}

func TestNeffShuffle_Shuffle(t *testing.T) {

	electionID := "deadbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	rosterLen := 2
	roster := authority.FromAuthority(fake.NewAuthority(rosterLen, fake.NewSigner))

	election := fake.NewElection(electionID)
	election.Roster = roster

	shuffledBallots := append([]etypes.Ciphervote{}, election.Suffragia.Ciphervotes...)
	election.ShuffleInstances = append(election.ShuffleInstances, etypes.ShuffleInstance{ShuffledBallots: shuffledBallots})

	election.ShuffleThreshold = 1

	service := fake.NewService(electionID, election, serdecontext)

	actor := Actor{
		rpc:     fake.NewBadRPC(),
		mino:    fake.Mino{},
		service: service,
		context: serdecontext,
	}

	err = actor.Shuffle(electionIDBuf)
	require.EqualError(t, err, fake.Err("failed to stream"))

	rpc := fake.NewStreamRPC(fake.NewReceiver(), fake.NewBadSender())
	actor.rpc = rpc

	err = actor.Shuffle(electionIDBuf)
	require.EqualError(t, err, fake.Err("failed to start shuffle"))

	rpc = fake.NewStreamRPC(fake.NewBadReceiver(), fake.Sender{})
	actor.rpc = rpc

	// we no longer use the receiver:
	err = actor.Shuffle(electionIDBuf)
	require.NoError(t, err)

	recv := fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), types.NewEndShuffle()))

	rpc = fake.NewStreamRPC(recv, fake.Sender{})
	actor.rpc = rpc

	err = actor.Shuffle(electionIDBuf)
	require.NoError(t, err)
}

func TestNeffShuffle_Verify(t *testing.T) {

	actor := Actor{}

	rand := suite.RandomStream()
	h := suite.Scalar().Pick(rand)
	H := suite.Point().Mul(h, nil)

	k := 3
	X := make([]kyber.Point, k)
	Y := make([]kyber.Point, k)

	for i := 0; i < k; i++ {
		// Embed the message into a curve point
		message := "Test" + strconv.Itoa(i)
		M := suite.Point().Embed([]byte(message), random.New())

		// ElGamal-encrypt the point to produce ciphertext (K,C).
		k := suite.Scalar().Pick(random.New()) // ephemeral private key
		K := suite.Point().Mul(k, nil)         // ephemeral DH public key
		S := suite.Point().Mul(k, H)           // ephemeral DH shared secret
		C := S.Add(S, M)                       // message blinded with secret
		X[i] = K
		Y[i] = C
	}

	Kbar, Cbar, prover := shuffleKyber.Shuffle(suite, nil, H, X, Y, rand)
	shuffleProof, _ := proof.HashProve(suite, protocolName, prover)

	err := actor.Verify(suite.String(), Y, Y, H, Kbar, Cbar, shuffleProof)
	require.EqualError(t, err, "invalid PairShuffleProof")

	err = actor.Verify(suite.String(), X, Y, H, Kbar, Cbar, shuffleProof)
	require.NoError(t, err)
}

// -----------------------------------------------------------------------------
// Utility functions

type fakeAuthorityFactory struct {
	serde.Factory
}

func (f fakeAuthorityFactory) AuthorityOf(ctx serde.Context, rosterBuf []byte) (authority.Authority, error) {
	fakeAuthority := &fakeAuthority{}
	return fakeAuthority, nil
}

type fakeAuthority struct {
	serde.Message
	serde.Fingerprinter
	crypto.CollectiveAuthority

	len int
}

func (f fakeAuthority) Apply(c authority.ChangeSet) authority.Authority {
	return nil
}

// Diff should return the change set to apply to get the given authority.
func (f fakeAuthority) Diff(a authority.Authority) authority.ChangeSet {
	return nil
}

func (f fakeAuthority) PublicKeyIterator() crypto.PublicKeyIterator {
	signers := make([]crypto.Signer, f.len)
	signers[0] = fake.NewSigner()

	return fake.NewPublicKeyIterator(signers)
}

func (f fakeAuthority) AddressIterator() mino.AddressIterator {
	addrs := make([]mino.Address, f.Len())
	for i := 0; i < f.Len(); i++ {
		addrs[i] = fake.NewAddress(i)
	}
	return fake.NewAddressIterator(addrs)
}

func (f fakeAuthority) Len() int {
	return f.len
}

// fakeManager is a fake manager
//
// - implements txn.Manager
type fakeManager struct {
	txn.Manager
}

func (fakeManager) Sync() error {
	return nil
}

func (fakeManager) Make(args ...txn.Arg) (txn.Transaction, error) {
	return nil, fake.GetError()
}
