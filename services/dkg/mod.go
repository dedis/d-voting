package dkg

import (
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/kyber/v3"
)

// DKG defines the primitive to start a DKG protocol
type DKG interface {
	// Listen starts the RPC. This function should be called on each node that
	// wishes to participate in a DKG. electionID is NOT hex-encoded.
	Listen(electionID []byte, txmngr txn.Manager) (Actor, error)

	// GetActor allows to retrieve the Actor corresponding to a given
	// electionID. electionID is NOT hex-encoded.
	GetActor(electionID []byte) (Actor, bool)
}

// Actor defines the primitives to use a DKG protocol
//
// An actor is directly linked to an election; one should not be able to create
// an Actor for an election that does not exist
type Actor interface {
	// Setup must be first called by ONE of the actors to use the subsequent
	// functions. It creates the public distributed key and the private share on
	// each node. Each node represented by a player must first execute Listen().
	// Returns an error if Setup was already done.
	Setup() (pubKey kyber.Point, err error)

	// GetPublicKey returns the collective public key. Returns an error if the
	// setup has not been done.
	GetPublicKey() (kyber.Point, error)

	Encrypt(message []byte) (K, C kyber.Point, remainder []byte, err error)

	// ComputePubshares sends a decryption request to all nodes in order to gather
	// the public shares
	ComputePubshares() error

	// MarshalJSON returns a JSON-encoded bytestring containing all the actor
	// data that is meant to be persistent.
	MarshalJSON() ([]byte, error)
}
