package types

import (
	etypes "github.com/dedis/d-voting/contracts/evoting/types"
)

// CreateFormRequest defines the HTTP request for creating a form
type CreateFormRequest struct {
	UserID        string
	Configuration etypes.Configuration
}

// PermissionOperationRequest defines the HTTP request for performing
// an operation request
type PermissionOperationRequest struct {
	TargetUserID     string
	PerformingUserID string
}

// CreateFormResponse defines the HTTP response when creating a form
type CreateFormResponse struct {
	FormID string // hex-encoded
	Token  string
}

// CastVoteRequest defines the HTTP request for casting a vote
type CastVoteRequest struct {
	VoterID string
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

// UpdateFormRequest defines the HTTP request for updating a form
type UpdateFormRequest struct {
	Action string
	UserID string
}

// GetFormResponse defines the HTTP response when getting the form info
type GetFormResponse struct {
	// FormID is hex-encoded
	FormID          string
	Configuration   etypes.Configuration
	Status          uint16
	Pubkey          string
	Result          []etypes.Ballot
	Roster          []string
	ChunksPerBallot int
	BallotSize      int
	Voters          []string
}

// LightForm represents a light version of the form
type LightForm struct {
	FormID string
	Title  etypes.Title
	Status uint16
	Pubkey string
}

// GetFormsResponse defines the HTTP response when getting all forms
// infos.
type GetFormsResponse struct {
	Forms []LightForm
}

// HTTPError defines the standard error format
type HTTPError struct {
	Title   string
	Code    uint
	Message string
	Args    map[string]interface{}
}
