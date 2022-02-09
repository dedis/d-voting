package types

import (
	"encoding/base64"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/dela/serde/registry"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
	"io"
)

type ID string
type Status uint16

const (
	Initial            Status = 0
	Open               Status = 1
	Closed             Status = 2
	ShuffledBallots    Status = 3
	PubSharesSubmitted Status = 4
	// DecryptedBallots = 4
	ResultAvailable Status = 5
	Canceled        Status = 6
)

// electionFormat contains the supported formats for the election. Right now
// only JSON is supported.
var electionFormat = registry.NewSimpleRegistry()

// RegisterElectionFormat registers the engine for the provided format
func RegisterElectionFormat(format serde.Format, engine serde.FormatEngine) {
	electionFormat.Register(format, engine)
}

// ElectionKey is the key of the election factory
type ElectionKey struct{}

// Election contains all information about a simple election
//
// - implements serde.Message
type Election struct {
	Configuration Configuration

	// ElectionID is the hex-encoded SHA256 of the transaction ID that creates
	// the election
	ElectionID string

	AdminID string
	Status  Status
	Pubkey  kyber.Point

	// BallotSize represents the total size in bytes of one ballot. It is used
	// to pad smaller ballots such that all  ballots cast have the same size
	BallotSize int

	// Suffragia is a map from User ID to their encrypted ballot
	Suffragia Suffragia

	// ShuffleInstances is all the shuffles, along with their proof and identity
	// of shuffler.
	ShuffleInstances []ShuffleInstance

	// ShuffleThreshold is set based on the roster. We save it so we do not have
	// to compute it based on the roster each time we need it.
	ShuffleThreshold int

	// PubShareSubmissions is an array containing all the submission of pubShares.
	// One entry per node, the index of a node's submission is its index in the
	// roster.
	PubShareSubmissions []PubSharesSubmission

	DecryptedBallots []Ballot

	// roster is set when the election is created based on the current
	// roster of the node stored in the global state. The roster will not change
	// during an election and will be used for DKG and Neff. Its type is
	// authority.Authority.

	RosterBuf []byte
}

// Serialize implements serde.Message
func (e Election) Serialize(ctx serde.Context) ([]byte, error) {
	format := electionFormat.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, e)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode election: %v", err)
	}

	return data, nil
}

// ElectionFactory provides the mean to deserialize an election. It naturally
// uses the electionFormat.
//
// - implements serde.Factory
type ElectionFactory struct{}

// Deserialize implements serde.Factory
func (ElectionFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := electionFormat.Get(ctx.GetFormat())

	message, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode: %v", err)
	}

	return message, nil
}

// ChunksPerBallot returns the number of chunks of El Gamal pairs needed to
// represent an encrypted ballot, knowing that one chunk is 29 bytes at most.
func (e *Election) ChunksPerBallot() int {
	if e.BallotSize%29 == 0 {
		return e.BallotSize / 29
	}

	return e.BallotSize/29 + 1
}

// RandomVector is a slice of kyber.Scalar (encoded) which is used to prove
// and verify the proof of a shuffle
type RandomVector [][]byte

// Unmarshal ...
func (r RandomVector) Unmarshal() ([]kyber.Scalar, error) {
	e := make([]kyber.Scalar, len(r))

	for i, v := range r {
		scalar := suite.Scalar()
		err := scalar.UnmarshalBinary(v)
		if err != nil {
			return nil, xerrors.Errorf("cannot unmarshall election random vector: %v", err)
		}
		e[i] = scalar
	}

	return e, nil
}

// LoadFromScalars marshals a given vector of scalars into the current
// RandomVector
func (r *RandomVector) LoadFromScalars(e []kyber.Scalar) error {
	*r = make([][]byte, len(e))

	for i, scalar := range e {
		v, err := scalar.MarshalBinary()
		if err != nil {
			return xerrors.Errorf("could not marshal random vector: %v", err)
		}
		(*r)[i] = v
	}

	return nil
}

// ShuffleInstance is an instance of a shuffle, it contains the shuffled
// ballots, the proofs and the identity of the shuffler.
type ShuffleInstance struct {
	// ShuffledBallots contains the list of shuffled ciphertext for this round
	ShuffledBallots []Ciphervote

	// ShuffleProofs is the proof of the shuffle for this round
	ShuffleProofs []byte

	// ShufflerPublicKey is the key of the node who made the given shuffle.
	ShufflerPublicKey []byte
}

// Configuration contains the configuration of a new poll.
type Configuration struct {
	MainTitle string
	Scaffold  []Subject
}

// MaxBallotSize returns the maximum number of bytes required to store a ballot
func (c *Configuration) MaxBallotSize() int {
	size := 0
	for _, subject := range c.Scaffold {
		size += subject.MaxEncodedSize()
	}
	return size
}

// GetQuestion finds the question associated to a given ID and returns it.
// Returns nil if no question found.
func (c *Configuration) GetQuestion(ID ID) Question {
	for _, subject := range c.Scaffold {
		question := subject.GetQuestion(ID)

		if question != nil {
			return question
		}
	}

	return nil
}

// IsValid returns true if and only if the whole configuration is coherent and
// valid
func (c *Configuration) IsValid() bool {
	// serves as a set to check each ID is unique
	uniqueIDs := make(map[ID]bool)

	for _, subject := range c.Scaffold {
		if !subject.isValid(uniqueIDs) {
			return false
		}
	}

	// if an id is not encoded in base64
	for id := range uniqueIDs {
		_, err := base64.StdEncoding.DecodeString(string(id))
		if err != nil {
			return false
		}
	}

	return true
}

type Suffragia struct {
	UserIDs     []string
	Ciphervotes []Ciphervote
}

// CastVote adds a new vote and its associated user or updates a user's vote.
func (s *Suffragia) CastVote(userID string, ciphervote Ciphervote) {
	for i, u := range s.UserIDs {
		if u == userID {
			s.Ciphervotes[i] = ciphervote
			return
		}
	}

	s.UserIDs = append(s.UserIDs, userID)
	s.Ciphervotes = append(s.Ciphervotes, ciphervote.Copy())
}

// CiphervotesFromPairs ...
func CiphervotesFromPairs(X, Y [][]kyber.Point) ([]Ciphervote, error) {
	if len(X) != len(Y) {
		return nil, xerrors.Errorf("X and Y must have same length: %d != %d",
			len(X), len(Y))
	}

	if len(X) == 0 {
		return nil, xerrors.Errorf("ElGamal pairs are empty")
	}

	NQ := len(X)   // sequence size
	k := len(X[0]) // number of votes
	res := make([]Ciphervote, k)

	for i := 0; i < k; i++ {
		x := make([]kyber.Point, NQ)
		y := make([]kyber.Point, NQ)

		for j := 0; j < NQ; j++ {
			x[j] = X[j][i]
			y[j] = Y[j][i]
		}

		ciphervote, err := ciphervoteFromPairs(x, y)
		if err != nil {
			return nil, xerrors.Errorf("failed to init from ElGamal pairs: %v", err)
		}

		res[i] = ciphervote
	}

	return res, nil
}

func ciphervoteFromPairs(ks []kyber.Point, cs []kyber.Point) (Ciphervote, error) {
	if len(ks) != len(cs) {
		return Ciphervote{}, xerrors.Errorf("ks and cs must have same length: %d != %d",
			len(ks), len(cs))
	}

	res := make(Ciphervote, len(ks))

	for i := range ks {
		res[i] = EGPair{
			K: ks[i],
			C: cs[i],
		}
	}

	return res, nil
}

// PubShare represents a public share.
type PubShare kyber.Point

// PubSharesSubmission holds all the PubSharesSubmission produced by a given node, []PubSharesSubmission per ballot
type PubSharesSubmission [][]PubShare

func (p PubSharesSubmission) FingerPrint(writer io.Writer) error {
	for _, ballotShares := range p {
		for _, pubShare := range ballotShares {
			_, err := pubShare.MarshalTo(writer)
			if err != nil {
				return xerrors.Errorf("failed to Marshal V: %v", err)
			}
		}
	}

	return nil
}
