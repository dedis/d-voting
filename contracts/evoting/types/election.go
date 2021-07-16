package types

type ID string

// todo : status should be string ?
type status uint16

const (
	// Initial = 0
	Open            = 1
	Closed          = 2
	ShuffledBallots = 3
	// DecryptedBallots = 4
	ResultAvailable = 5
	Canceled        = 6
)

// Election contains all information about a simple election
type Election struct {
	Title            string
	ElectionID       ID
	AdminId          string
	Status           status // Initial | Open | Closed | Shuffling | Decrypting
	Pubkey           []byte
	EncryptedBallots map[string][]byte
	ShuffledBallots  map[int][][]byte
	Proofs           map[int][]byte
	DecryptedBallots []Ballot
	ShuffleThreshold int
	Members          []CollectiveAuthorityMember
	Format           string
}

// Ballot contains all information about a simple ballot
type Ballot struct {
	Vote string
}
