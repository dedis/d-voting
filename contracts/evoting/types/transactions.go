package types

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/xerrors"
)

type ElectionsMetadata struct {
	ElectionsIds []string
}

type CreateElectionTransaction struct {
	ElectionID       string
	Title            string
	AdminId          string
	ShuffleThreshold int
	Members          []CollectiveAuthorityMember
	Format           string
}

type CastVoteTransaction struct {
	ElectionID string
	UserId     string
	Ballot     []byte
}

type CloseElectionTransaction struct {
	ElectionID string
	UserId     string
}

type ShuffleBallotsTransaction struct {
	ElectionID      string
	Round           int
	ShuffledBallots [][]byte
	Proof           []byte
}

type DecryptBallotsTransaction struct {
	ElectionID       string
	UserId           string
	DecryptedBallots []Ballot
}

type CancelElectionTransaction struct {
	ElectionID string
	UserId     string
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
