package types

// NewDKGRequest defines the request to create a new DGK
type NewDKGRequest struct {
	ElectionID string // hex-encoded
}

// UpdateDKG defines the input used to update dkg
type UpdateDKG struct {
	Action string
}
