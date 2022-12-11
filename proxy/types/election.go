package types

import (
	etypes "github.com/dedis/d-voting/contracts/evoting/types"
)

// TransactionStatus is the status of a transaction
type TransactionStatus byte

const (
	// UnknownTransactionStatus is the basic status of a transaction
	UnknownTransactionStatus TransactionStatus = 0
	// IncludedTransaction is the status of a transaction that has been included
	IncludedTransaction TransactionStatus = 1
	// RejectedTransaction is the status of a transaction will never be included
	RejectedTransaction TransactionStatus = 2
)





// CreateFormRequest defines the HTTP request for creating a form
type CreateFormRequest struct {
	AdminID       string
	Configuration etypes.Configuration
}

// CreateFormResponse defines the HTTP response when creating a form
type CreateFormResponse struct {
	FormID string // hex-encoded
	Token string
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

// TransactionInfo defines the information of a transaction 
type TransactionInfo struct {
	Status TransactionStatus // 0 if not yet included, 1 if included, 2 if rejected
	TransactionID []byte
	LastBlockIdx uint64  // last block of the chain when the transaction was added to the pool
	Time 	   int64 // time when the transaction was added to the pool
	Hash    []byte // signature of the transaction
	Signature []byte // signature of the transaction
}

// TransactionInfoToSend defines the HTTP response when sending 
// transaction infos to the client so that he can use the status
// of the transaction to know if it has been included or not
// and if it has not been included, he can just use the token
// and ask again later
type TransactionInfoToSend struct {
	Status TransactionStatus // 0 if not yet included, 1 if included, 2 if rejected
	Token string
}




// UpdateFormRequest defines the HTTP request for updating a form
type UpdateFormRequest struct {
	Action string
}

// GetFormResponse defines the HTTP response when getting the form info
type GetFormResponse struct {
	// FormID is hex-encoded
	FormID      string
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
	Title      string
	Status     uint16
	Pubkey     string
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
