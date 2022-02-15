package types

import (
	"go.dedis.ch/kyber/v3/suites"
)

var suite = suites.MustFind("Ed25519")

// LoginResponse defines the HTPP request for login
type LoginResponse struct {
	UserID string
	Token  string
}

// CreateElectionRequest defines the HTTP request for creating an election
type CreateElectionRequest struct {
	Configuration Configuration
	AdminID       string
}

// OpenElectionRequest defines the HTTP request for opening an election
type OpenElectionRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
}

// CreateElectionResponse defines the HTTP response when creating an election
type CreateElectionResponse struct {
	// ElectionID is hex-encoded
	ElectionID string
}

// CastVoteRequest defines the HTTP request for casting a vote
type CastVoteRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	// Marshalled representation of Ciphervote
	Ballot []byte
	Token  string
}

// CastVoteResponse degines the HTTP response when casting a vote
type CastVoteResponse struct {
}

// CloseElectionRequest degines the HTTP request for closing an election
type CloseElectionRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Token      string
}

// CloseElectionResponse defines the HTTP response when closing an election
type CloseElectionResponse struct {
}

// ShuffleBallotsRequest defines the HTTP request for shuffling the ballots
type ShuffleBallotsRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Token      string
}

// ShuffleBallotsResponse defines the HTTP response when shuffling the ballots
type ShuffleBallotsResponse struct {
	Message string
}

type BeginDecryptionRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Token      string
}

type BeginDecryptionResponse struct {
	Message string
}

// DecryptBallotsRequest defines the HTTP request for decrypting the ballots
type DecryptBallotsRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Token      string
}

// DecryptBallotsResponse defines the HTTP response when decrypting the ballots
type DecryptBallotsResponse struct {
}

// GetElectionResultRequest defines the HTTP request for getting the election
// result.
type GetElectionResultRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	// UserId   string
	Token string
}

// GetElectionResultResponse defines the HTTP response when getting the election
// result.
type GetElectionResultResponse struct {
	Result []Ballot
}

// GetElectionInfoRequest defines the HTTP request for getting the election info
type GetElectionInfoRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	// UserId string
	Token string
}

// GetElectionInfoResponse defines the HTTP response when getting the election
// info
type GetElectionInfoResponse struct {
	// ElectionID is hex-encoded
	ElectionID    string
	Configuration Configuration
	Status        uint16
	Pubkey        string
	Result        []Ballot
	Format        string
}

// GetAllElectionsInfoRequest defines the HTTP request for getting all elections
// infos.
type GetAllElectionsInfoRequest struct {
	// UserId string
	Token string
}

// GetAllElectionsInfoResponse defines the HTTP response when getting all
// elections infos.
type GetAllElectionsInfoResponse struct {
	// UserId         string
	AllElectionsInfo []GetElectionInfoResponse
}

// GetAllElectionsIDsRequest defines the HTTP request for getting all election
// IDs.
type GetAllElectionsIDsRequest struct {
	// UserId string
	Token string
}

// GetAllElectionsIDsResponse defines the HTTP response when getting all
// election IDs.
type GetAllElectionsIDsResponse struct {
	// UserId         string
	// ElectionsIDs is a slice of hex-encoded election IDs
	ElectionsIDs []string
}

// CancelElectionRequest defines the HTTP request for canceling an election
type CancelElectionRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Token      string
}

// CancelElectionResponse defines the HTTP response when canceling an election
type CancelElectionResponse struct {
}
