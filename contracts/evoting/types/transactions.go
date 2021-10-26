package types

import (
	"crypto/rand"
	"encoding/hex"
	"golang.org/x/xerrors"
)

type ElectionsMetadata struct {
	ElectionsIDs ElectionIDs
}

// ElectionIDs is a slice of hex-encoded election IDs
type ElectionIDs []string

// Contains checks if el is present
func (e ElectionIDs) Contains(el string) bool {
	for _, e1 := range e {
		if e1 == el {
			return true
		}
	}

	return false
}

// Add adds an election ID or returns an error if already present
func (e *ElectionIDs) Add(id string) error {
	if e.Contains(id) {
		return xerrors.Errorf("id '%s' already exist")
	}

	*e = append(*e, id)

	return nil
}

type CreateElectionTransaction struct {
	Title   string
	AdminID string
	Format  string
}

type OpenElectionTransaction struct {
	// ElectionID is hex-encoded
	ElectionID string
}

type CastVoteTransaction struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Ballot     Ciphertext
}

type CloseElectionTransaction struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
}

type ShuffleBallotsTransaction struct {
	ElectionID      string
	Round           int
	ShuffledBallots Ciphertexts
	Proof           []byte
	Signature       []byte
	Signer          CollectiveAuthorityMember
}

type DecryptBallotsTransaction struct {
	// ElectionID is hex-encoded
	ElectionID       string
	UserID           string
	DecryptedBallots []Ballot
}

type CancelElectionTransaction struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
}

// RandomID returns the hex encoding of a randomly created 32 byte ID.
func RandomID() (string, error) {
	buf := make([]byte, 32)
	n, err := rand.Read(buf)
	if err != nil || n != 32 {
		return "", xerrors.Errorf("failed to fill buffer with random data: %v", err)
	}

	return hex.EncodeToString(buf), nil
}
