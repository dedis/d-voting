package shuffle

import (
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/kyber/v3"
)

// Shuffle defines the primitive to start a shuffle protocol
type Shuffle interface {
	// Listen starts the RPC. This function should be called on each node that
	// wishes to participate in a shuffle.
	Listen(signer crypto.Signer) (Actor, error)
}

// Actor defines the primitives to use a shuffle protocol
type Actor interface {

	// Shuffle must be called by ONE of the actor to shuffle the list of ElGamal
	// pairs.
	// Each node represented by a player must first execute Listen().
	Shuffle(co crypto.CollectiveAuthority, electionId string) (err error)

	// Verify allows to verify a shuffle
	Verify(suiteName string, Ks []kyber.Point, Cs []kyber.Point, pubKey kyber.Point, KsShuffled []kyber.Point,
		CsShuffled []kyber.Point, proof []byte) (err error)
}
