package types

import (
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

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
	Title string

	// ElectionID is the hex-encoded SHA256 of the transaction ID that creates
	// the election
	ElectionID string

	AdminID string
	Status  status // Initial | Open | Closed | Shuffling | Decrypting
	Pubkey  []byte

	// EncryptedBallots is a map from User ID to their ballot ciphertext
	// EncryptedBallots map[string][]byte
	EncryptedBallots EncryptedBallots

	//ShuffleInstances is all the shuffles, along with their proof and identity of shuffler.
	ShuffleInstances []ShuffleInstance

	// ShuffleThreshold is set based on the roster. We save it so we don't have
	// to compute it based on the roster each time we need it.
	ShuffleThreshold int

	DecryptedBallots []Ballot
	Format           string

	// roster is once set when the election is created based on the current
	// roster of the node stored in the global state. The roster won't change
	// during an election and will be used for DKG and Neff. Its type is
	// authority.Authority.
	RosterBuf []byte
}

// ShuffleInstance is an instance of a shuffle, it contains the shuffled ballots, the proofs and the
// identity of the shuffler.
type ShuffleInstance struct {
	// ShuffledBallots contains the list of shuffled ciphertext for this round
	ShuffledBallots Ciphertexts

	// ShuffleProofs is the proof of the shuffle for this round
	ShuffleProofs []byte

	//Shuffler is the identity of the node who made the given shuffle.
	ShufflerPublicKey []byte
}

// Ballot contains all information about a simple ballot
type Ballot struct {
	Vote string
}

// EncryptedBallots maintains a list of encrypted ballots with the associated
// user ID.
type EncryptedBallots struct {
	UserIDs []string
	Ballots Ciphertexts
}

// CastVote updates a user's vote or add a new vote and its associated user.
func (e *EncryptedBallots) CastVote(userID string, encryptedVote Ciphertext) {
	for i, u := range e.UserIDs {
		if u == userID {
			e.Ballots[i] = encryptedVote
			return
		}
	}

	e.UserIDs = append(e.UserIDs, userID)
	e.Ballots = append(e.Ballots, encryptedVote.Copy())
}

// GetBallotFromUser returns the ballot associated to a user. Returns nil if
// user is not found.
func (e *EncryptedBallots) GetBallotFromUser(userID string) (Ciphertext, bool) {
	for i, u := range e.UserIDs {
		if u == userID {
			return e.Ballots[i].Copy(), true
		}
	}

	return Ciphertext{}, false
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

// Ciphertexts represents a list of Ciphertext
type Ciphertexts []Ciphertext

// GetKsCs returns corresponding kyber.Points from the ciphertexts
func (c Ciphertexts) GetKsCs() (ks []kyber.Point, cs []kyber.Point, err error) {
	ks = make([]kyber.Point, len(c))
	cs = make([]kyber.Point, len(c))

	for i, ct := range c {
		k, c, err := ct.GetPoints()
		if err != nil {
			return nil, nil, xerrors.Errorf("failed to get points: %v", err)
		}

		ks[i] = k
		cs[i] = c
	}

	return ks, cs, nil
}

// InitFromKsCs sets the ciphertext based on ks, cs
func (c *Ciphertexts) InitFromKsCs(ks []kyber.Point, cs []kyber.Point) error {
	if len(ks) != len(cs) {
		return xerrors.Errorf("ks and cs must have same length: %d != %d",
			len(ks), len(cs))
	}

	*c = make([]Ciphertext, len(ks))

	for i := range ks {
		var ct Ciphertext

		err := ct.FromPoints(ks[i], cs[i])
		if err != nil {
			return xerrors.Errorf("failed to init ciphertext: %v", err)
		}

		(*c)[i] = ct
	}

	return nil
}
