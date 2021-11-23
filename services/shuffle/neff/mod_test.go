package neff

import (
	"encoding/hex"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	"strconv"
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	neffShuffleTypes "github.com/dedis/d-voting/services/shuffle/neff/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	shuffleKyber "go.dedis.ch/kyber/v3/shuffle"
	"go.dedis.ch/kyber/v3/util/random"
)

func TestNeffShuffle_Listen(t *testing.T) {

	NeffShuffle := NewNeffShuffle(fake.Mino{}, &FakeService{}, &FakePool{}, nil, fakeAuthorityFactory{}, fake.NewSigner())

	actor, err := NeffShuffle.Listen(fake.NewSigner())
	require.NoError(t, err)

	require.NotNil(t, actor)
}

func TestNeffShuffle_Shuffle(t *testing.T) {

	electionId := []byte("dummyId")

	actor := Actor{
		rpc:       fake.NewBadRPC(),
		mino:      fake.Mino{},
		service:   &FakeService{electionId: types.ID(hex.EncodeToString(electionId))},
		rosterFac: fakeAuthorityFactory{},
	}

	err := actor.Shuffle(electionId)
	require.EqualError(t, err, fake.Err("failed to stream"))

	rpc := fake.NewStreamRPC(fake.NewReceiver(), fake.NewBadSender())
	actor.rpc = rpc

	err = actor.Shuffle(electionId)
	require.EqualError(t, err, fake.Err("failed to start shuffle"))

	rpc = fake.NewStreamRPC(fake.NewBadReceiver(), fake.Sender{})
	actor.rpc = rpc

	// we no longer use the receiver:
	err = actor.Shuffle(electionId)
	require.NoError(t, err)

	recv := fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), nil))

	recv = fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), neffShuffleTypes.NewEndShuffle()))

	rpc = fake.NewStreamRPC(recv, fake.Sender{})
	actor.rpc = rpc

	err = actor.Shuffle(electionId)
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

	//AuthorityOf(serde.Context, []byte) (authority.Authority, error)
}

func (f fakeAuthorityFactory) AuthorityOf(ctx serde.Context, rosterBuf []byte) (authority.Authority, error) {
	fakeAuthority := &fakeAuthority{}
	return fakeAuthority, nil
}

type fakeAuthority struct {
	serde.Message
	serde.Fingerprinter
	crypto.CollectiveAuthority
}

func (f fakeAuthority) Apply(c authority.ChangeSet) authority.Authority {
	return nil
}

// Diff should return the change set to apply to get the given authority.
func (f fakeAuthority) Diff(a authority.Authority) authority.ChangeSet {
	return nil
}

func (f fakeAuthority) PublicKeyIterator() crypto.PublicKeyIterator {
	signers := make([]crypto.Signer, 2)
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
	return 2
}
