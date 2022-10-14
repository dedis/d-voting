package dkg

import (
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/kyber/v3"
)

// StatusCode is the type used to define a DKG status
type StatusCode uint16

// Status defines a struct to hold the status and error context if any.
type Status struct {
	Status StatusCode
	// The following is mostly usefull to return context to a frontend in case
	// of error.
	Err  error
	Args map[string]interface{}
}

const (
	// Initialized is when the actor has been initialized
	Initialized StatusCode = 0
	// Setup is when the actor was set up
	Setup StatusCode = 1
	// Failed is when the actor failed to set up
	Failed StatusCode = 2
	// Dealing is when the actor is sending its deals
	Dealing = 3
	// Responding is when the actor sends its responses on the deals
	Responding = 4
	// Certifying is when the actor is validating its responses
	Certifying = 5
	// Certified is then the actor is certified
	Certified = 6
)

// DKG defines the primitive to start a DKG protocol
type DKG interface {
	// Listen starts the RPC. This function should be called on each node that
	// wishes to participate in a DKG. formID is NOT hex-encoded.
	Listen(formID []byte, txmngr txn.Manager) (Actor, error)

	// GetActor allows to retrieve the Actor corresponding to a given
	// formID. formID is NOT hex-encoded.
	GetActor(formID []byte) (Actor, bool)
}

// Actor defines the primitives to use a DKG protocol
//
// An actor is directly linked to a form; one should not be able to create
// an Actor for a form that does not exist
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

	// ComputePubshares sends a decryption request to all nodes. Nodes will then
	// publish their public shares on the smart contract.
	ComputePubshares() error

	// MarshalJSON returns a JSON-encoded bytestring containing all the actor
	// data that is meant to be persistent.
	MarshalJSON() ([]byte, error)

	// Status returns the actor's status
	Status() Status
}
