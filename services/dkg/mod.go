package dkg

import (
	"go.dedis.ch/kyber/v3"
)

// DKG defines the primitive to start a DKG protocol
type DKG interface {
	// Listen starts the RPC. This function should be called on each node that
	// wishes to participate in a DKG.
	Listen(electionID []byte) (Actor, error)

	// GetActor allows to retrieve the Actor corresponding a given electionID
	GetActor(electionID []byte) (Actor, error)
}

// Actor defines the primitives to use a DKG protocol
type Actor interface {
	// Setup must be first called by ONE of the actors to use the subsequent
	// functions. It creates the public distributed key and the private share on
	// each node. Each node represented by a player must first execute Listen().
	Setup(electionID []byte) (pubKey kyber.Point, err error)

	// GetPublicKey returns the collective public key. Returns an error it the
	// setup has not been done.
	GetPublicKey() (kyber.Point, error)

	Encrypt(message []byte) (K, C kyber.Point, remainder []byte, err error)
	Decrypt(K, C kyber.Point, electionID []byte) ([]byte, error)

	Reshare() error
}
