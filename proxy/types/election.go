package types

import (
	etypes "github.com/dedis/d-voting/contracts/evoting/types"
)

// CreateElectionRequest defines the HTTP request for creating an election
type CreateElectionRequest struct {
	AdminID       string
	Configuration etypes.Configuration
}

// CreateElectionResponse defines the HTTP response when creating an election
type CreateElectionResponse struct {
	ElectionID string // hex-encoded
}

// CastVoteRequest defines the HTTP request for casting a vote
type CastVoteRequest struct {
	UserID string
	// Marshalled representation of Ciphervote. It contains []{K:,C:}
	Ballot CiphervoteJSON
}

// CiphervoteJSON is the JSON representation of a ciphervote
type CiphervoteJSON []EGPairJSON

// EGPairJSON is the JSON representation of an ElGamal pair
type EGPairJSON struct {
	K []byte
	C []byte
}

// UpdateElectionRequest defines the HTTP request for updating an election
type UpdateElectionRequest struct {
	Action string
}

// GetElectionResponse defines the HTTP response when getting the election info
type GetElectionResponse struct {
	// ElectionID is hex-encoded
	ElectionID    string
	Configuration etypes.Configuration
	Status        uint16
	Pubkey        string
	Result        []etypes.Ballot
	Format        string
}

// GetElections defines the HTTP request for getting all elections infos.
type GetElections struct {
	// UserId string
	Token string
}

// LightElection represents a light version of the election
type LightElection struct {
	ElectionID string
	Title      string
	Status     uint16
	Pubkey     string
}

// GetElectionsResponse defines the HTTP response when getting all elections
// infos.
type GetElectionsResponse struct {
	Elections []LightElection
}

// HTTPError defines the standard error format
type HTTPError struct {
	Title   string
	Code    uint
	Message string
	Args    map[string]interface{}
}
