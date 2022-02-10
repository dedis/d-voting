package shuffle

import (
	"go.dedis.ch/dela/core/txn"
)

// Shuffle defines the primitive to start a shuffle protocol
type Shuffle interface {
	// Listen starts the RPC. This function should be called on each node that
	// wishes to participate in a shuffle.
	Listen(txmngr txn.Manager) (Actor, error)
}

// Actor defines the primitives to use a shuffle protocol
type Actor interface {
	// Shuffle must be called by ONE of the actor to shuffle the list of ElGamal
	// pairs. Each node represented by a player must first execute Listen().
	Shuffle(electionID []byte) (err error)
}
