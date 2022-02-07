package types

import (
	"fmt"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
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
	Ballot     EncryptedBallot
	Token      string
}

type CastVoteResponse struct {
}

// Ciphertext wraps K, C kyber points into their binary representation
type Ciphertext struct {
	K []byte
	C []byte
}

// GetPoints returns the kyber.Point curve points
func (ct Ciphertext) GetPoints() (k kyber.Point, c kyber.Point, err error) {
	k = suite.Point()
	c = suite.Point()

	err = k.UnmarshalBinary(ct.K)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to unmarshal K: %v", err)
	}

	err = c.UnmarshalBinary(ct.C)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to unmarshal C: %v", err)
	}

	return k, c, nil
}

// Copy returns a copy of Ciphertext
func (ct Ciphertext) Copy() Ciphertext {
	return Ciphertext{
		K: append([]byte{}, ct.K...),
		C: append([]byte{}, ct.C...),
	}
}

// String returns a string representation of a ciphertext
func (ct Ciphertext) String() string {
	return fmt.Sprintf("{K: %x, C: %x}", ct.K, ct.C)
}

// FromPoints fills the ciphertext with k, c
func (ct *Ciphertext) FromPoints(k, c kyber.Point) error {
	buf, err := k.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to marshall k: %v", err)
	}

	ct.K = buf

	buf, err = c.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to marshall c: %v", err)
	}

	ct.C = buf

	return nil
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
