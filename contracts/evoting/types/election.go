package types

type ID string

// todo : status should be string ?
type status uint16

const (
	Initial         status = 0
	Open            status = 1
	Closed          status = 2
	ShuffledBallots status = 3
	// DecryptedBallots = 4
	ResultAvailable status = 5
	Canceled        status = 6
)

// Election contains all information about a simple election
type Election struct {
	Title      string
	ElectionID ID
	AdminId    string
	Status     status // Initial | Open | Closed | Shuffling | Decrypting
	Pubkey     []byte

	// EncryptedBallots is a map from User ID to their ballot ciphertext
	// EncryptedBallots map[string][]byte
	EncryptedBallots *EncryptedBallots

	// ShuffledBallots is a map from shuffle round to shuffled ciphertexts
	// ShuffledBallots map[int][][]byte
	// Dimensions: <round, ballots, ballot>
	ShuffledBallots [][][]byte

	// ShufleProofs is a map from shuffle round to shuffle proofs
	// ShuffleProofs map[int][]byte
	ShuffledProofs [][]byte

	DecryptedBallots []Ballot
	ShuffleThreshold int
	Members          []CollectiveAuthorityMember
	Format           string
}

// Ballot contains all information about a simple ballot
type Ballot struct {
	Vote string
}

// EncryptedBallots maintains a list of encrypted ballots with the associated
// user ID.
type EncryptedBallots struct {
	UserIDs []string
	Ballots [][]byte
}

// CastVote updates a user's vote or add a new vote and its associated user.
func (e *EncryptedBallots) CastVote(userID string, encryptedVote []byte) {
	for i, u := range e.UserIDs {
		if u == userID {
			e.Ballots[i] = encryptedVote
			return
		}
	}

	e.UserIDs = append(e.UserIDs, userID)
	e.Ballots = append(e.Ballots, encryptedVote)
}

// GetBallotFromUser returns the ballot associated to a user. Returns nil if
// user is not found.
func (e *EncryptedBallots) GetBallotFromUser(userID string) []byte {
	for i, u := range e.UserIDs {
		if u == userID {
			return e.Ballots[i]
		}
	}

	return nil
}

// DeleteUser removes a user and its associated votes if found.
func (e *EncryptedBallots) DeleteUser(userID string) bool {
	for i, u := range e.UserIDs {
		if u == userID {
			e.UserIDs = append(e.UserIDs[:i], e.UserIDs[i+1:]...)
			e.Ballots = append(e.Ballots[:i], e.Ballots[i+1:]...)
			return true
		}
	}

	return false
}
