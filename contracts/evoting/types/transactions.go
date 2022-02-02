package types

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

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
		return xerrors.Errorf("id %q already exist", id)
	}

	*e = append(*e, id)

	return nil
}

type CreateElectionTransaction struct {
	Configuration Configuration
	AdminID       string
}

type OpenElectionTransaction struct {
	// ElectionID is hex-encoded
	ElectionID string
}

type CastVoteTransaction struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Ballot     EncryptedBallot
}

type CloseElectionTransaction struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
}

type ShuffleBallotsTransaction struct {
	ElectionID      string
	Round           int
	ShuffledBallots EncryptedBallots
	// RandomVector is the vector to be used to generate the proof of the next
	// shuffle
	RandomVector RandomVector
	// Proof is the proof corresponding to the shuffle of this transaction
	Proof []byte
	// Signature is the signature of the result of HashShuffle() with the private
	// key corresponding to PublicKey
	Signature []byte
	//PublicKey is the public key of the signer.
	PublicKey []byte
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

// HashShuffle hashes a given shuffle so that it can be signed or a signature
// can be verified, using a common template. Election is NOT hex encoded.
func (s ShuffleBallotsTransaction) HashShuffle(electionID []byte) ([]byte, error) {
	hash := sha256.New()

	hash.Write(electionID)

	shuffledBallots, err := json.Marshal(s.ShuffledBallots)
	if err != nil {
		return nil, xerrors.Errorf("could not marshal shuffled ballots : %v", err)
	}

	hash.Write(shuffledBallots)

	return hash.Sum(nil), nil
}
