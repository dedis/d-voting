package types

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
