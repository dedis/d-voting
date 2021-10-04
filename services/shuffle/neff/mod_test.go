package neff

import (
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

	NeffShuffle := NewNeffShuffle(fake.Mino{}, &FakeService{}, &FakePool{}, nil, fake.NewSigner())

	actor, err := NeffShuffle.Listen()
	require.NoError(t, err)

	require.NotNil(t, actor)
}

func TestNeffShuffle_Shuffle(t *testing.T) {

	electionId := "dummyId"

	actor := Actor{
		rpc:  fake.NewBadRPC(),
		mino: fake.Mino{},
	}

	fakeAuthority := fake.NewAuthority(1, fake.NewSigner)

	err := actor.Shuffle(fakeAuthority, electionId)
	require.EqualError(t, err, fake.Err("failed to stream"))

	rpc := fake.NewStreamRPC(fake.NewReceiver(), fake.NewBadSender())
	actor.rpc = rpc

	err = actor.Shuffle(fakeAuthority, electionId)
	require.EqualError(t, err, fake.Err("failed to send first message"))

	rpc = fake.NewStreamRPC(fake.NewBadReceiver(), fake.Sender{})
	actor.rpc = rpc

	err = actor.Shuffle(fakeAuthority, electionId)
	require.EqualError(t, err, fake.Err("got an error from '<nil>' while receiving"))

	recv := fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), nil))

	rpc = fake.NewStreamRPC(recv, fake.Sender{})
	actor.rpc = rpc

	err = actor.Shuffle(fakeAuthority, electionId)
	require.EqualError(t, err, "expected to receive an EndShuffle message, but go the following: <nil>")

	recv = fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), neffShuffleTypes.NewEndShuffle()))

	rpc = fake.NewStreamRPC(recv, fake.Sender{})
	actor.rpc = rpc

	err = actor.Shuffle(fakeAuthority, electionId)
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
