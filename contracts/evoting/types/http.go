package types

import (
	"go.dedis.ch/kyber/v3/suites"
)

var suite = suites.MustFind("Ed25519")

type LoginResponse struct {
	UserID string
	Token  string
}

type CollectiveAuthorityMember struct {
	Address   string
	PublicKey string
}

type CreateElectionRequest struct {
	Configuration Configuration
	AdminID       string
}

type OpenElectionRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
}

type CreateElectionResponse struct {
	// ElectionID is hex-encoded
	ElectionID string
}

type CastVoteRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	// Marshalled representation of Ciphervote
	Ballot []byte
	Token  string
}

type CastVoteResponse struct {
}

type CloseElectionRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Token      string
}

type CloseElectionResponse struct {
}

type ShuffleBallotsRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Token      string
}

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

type DecryptBallotsRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Token      string
}

type DecryptBallotsResponse struct {
}

type GetElectionResultRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	// UserId   string
	Token string
}

type GetElectionResultResponse struct {
	Result []Ballot
}

type GetElectionInfoRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	// UserId string
	Token string
}

type GetElectionInfoResponse struct {
	// ElectionID is hex-encoded
	ElectionID    string
	Configuration Configuration
	Status        uint16
	Pubkey        string
	Result        []Ballot
	Format        string
}

type GetAllElectionsInfoRequest struct {
	// UserId string
	Token string
}

type GetAllElectionsInfoResponse struct {
	// UserId         string
	AllElectionsInfo []GetElectionInfoResponse
}

type GetAllElectionsIDsRequest struct {
	// UserId string
	Token string
}

type GetAllElectionsIDsResponse struct {
	// UserId         string
	// ElectionsIDs is a slice of hex-encoded election IDs
	ElectionsIDs []string
}

type CancelElectionRequest struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Token      string
}

type CancelElectionResponse struct {
}
