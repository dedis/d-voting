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
	EncryptedBallots map[string][]byte

	// ShuffledBallots is a map from shuffle round to shuffled ciphertexts
	ShuffledBallots map[int][][]byte

	// ShufleProofs is a map from shuffle round to shuffle proofs
	ShuffleProofs map[int][]byte

	DecryptedBallots []Ballot
	ShuffleThreshold int
	Members          []CollectiveAuthorityMember
	Format           string
}

// Ballot contains all information about a simple ballot
type Ballot struct {
	Vote string
}
