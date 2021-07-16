package types

type LoginResponse struct {
	UserID string
	Token  string
}

type CollectiveAuthorityMember struct {
	Address   string
	PublicKey string
}

type CreateElectionRequest struct {
	Title            string
	AdminId          string
	Token            string
	Members          []CollectiveAuthorityMember
	ShuffleThreshold int
	Format           string
}

type CreateElectionResponse struct {
	ElectionID string
}

type CastVoteRequest struct {
	ElectionID string
	UserId     string
	Ballot     []byte
	Token      string
}

type CastVoteResponse struct {
}

// Wraps the ciphertext pairs
type Ciphertext struct {
	K []byte
	C []byte
}

type CloseElectionRequest struct {
	ElectionID string
	UserId     string
	Token      string
}

type CloseElectionResponse struct {
}

type ShuffleBallotsRequest struct {
	ElectionID string
	UserId     string
	Token      string
}

type ShuffleBallotsResponse struct {
	Message string
}

type DecryptBallotsRequest struct {
	ElectionID string
	UserId     string
	Token      string
}

type DecryptBallotsResponse struct {
}

type GetElectionResultRequest struct {
	ElectionID string
	// UserId   string
	Token string
}

type GetElectionResultResponse struct {
	Result []Ballot
}

type GetElectionInfoRequest struct {
	ElectionID string
	// UserId string
	Token string
}

type GetElectionInfoResponse struct {
	ElectionID string
	Title      string
	Status     uint16
	Pubkey     string
	Result     []Ballot
	Format     string
}

type GetAllElectionsInfoRequest struct {
	// UserId string
	Token string
}

type GetAllElectionsInfoResponse struct {
	// UserId         string
	AllElectionsInfo []GetElectionInfoResponse
}

type GetAllElectionsIdsRequest struct {
	// UserId string
	Token string
}

type GetAllElectionsIdsResponse struct {
	// UserId         string
	ElectionsIds []string
}

type CancelElectionRequest struct {
	ElectionID string
	UserId     string
	Token      string
}

type CancelElectionResponse struct {
}
