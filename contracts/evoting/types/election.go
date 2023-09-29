package types

import (
	"io"

	"github.com/c4dt/dela/core/ordering/cosipbft/authority"
	ctypes "github.com/c4dt/dela/core/ordering/cosipbft/types"
	"github.com/c4dt/dela/serde"
	"github.com/c4dt/dela/serde/registry"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

var suite = suites.MustFind("Ed25519")

// ID defines the ID of a ballot question
type ID string

// Status defines the status of the form
type Status uint16

const (
	// DecryptedBallots = 4
	// Initial is when the form has just been created
	Initial Status = 0
	// Open is when the form is open, i.e. it fetched the public key
	Open Status = 1
	// Closed is when no more users can cast ballots
	Closed Status = 2
	// ShuffledBallots is when the ballots have been shuffled
	ShuffledBallots Status = 3
	// PubSharesSubmitted is when we have enough shares to decrypt the ballots
	PubSharesSubmitted Status = 4
	// ResultAvailable is when the ballots have been decrypted
	ResultAvailable Status = 5
	// Canceled is when the form has been cancel
	Canceled Status = 6
)

// formFormat contains the supported formats for the form. Right now
// only JSON is supported.
var formFormat = registry.NewSimpleRegistry()

// RegisterFormFormat registers the engine for the provided format
func RegisterFormFormat(format serde.Format, engine serde.FormatEngine) {
	formFormat.Register(format, engine)
}

// FormKey is the key of the form factory
type FormKey struct{}

// Form contains all information about a simple form
//
// - implements serde.Message
type Form struct {
	Configuration Configuration

	// FormID is the hex-encoded SHA256 of the transaction ID that creates
	// the form
	FormID string

	Status Status
	Pubkey kyber.Point

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

	// PubsharesUnits is an array containing all the submission of pubShares.
	// Each node submits its share to its personal index from the DKG service.
	PubsharesUnits PubsharesUnits

	DecryptedBallots []Ballot

	// roster is set when the form is created based on the current
	// roster of the node stored in the global state. The roster will not change
	// during a form and will be used for DKG and Neff. Its type is
	// authority.Authority.

	Roster authority.Authority
}

// Serialize implements serde.Message
func (e Form) Serialize(ctx serde.Context) ([]byte, error) {
	format := formFormat.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, e)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode form: %v", err)
	}

	return data, nil
}

// FormFactory provides the mean to deserialize a form. It naturally
// uses the formFormat.
//
// - implements serde.Factory
type FormFactory struct {
	ciphervoteFac serde.Factory
	rosterFac     authority.Factory
}

// NewFormFactory creates a new Form factory
func NewFormFactory(cf serde.Factory, rf authority.Factory) FormFactory {
	return FormFactory{
		ciphervoteFac: cf,
		rosterFac:     rf,
	}
}

// Deserialize implements serde.Factory
func (e FormFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := formFormat.Get(ctx.GetFormat())

	ctx = serde.WithFactory(ctx, CiphervoteKey{}, e.ciphervoteFac)
	ctx = serde.WithFactory(ctx, ctypes.RosterKey{}, e.rosterFac)

	message, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode: %v", err)
	}

	return message, nil
}

// ChunksPerBallot returns the number of chunks of El Gamal pairs needed to
// represent an encrypted ballot, knowing that one chunk is 29 bytes at most.
func (e *Form) ChunksPerBallot() int {
	if e.BallotSize%29 == 0 {
		return e.BallotSize / 29
	}

	return e.BallotSize/29 + 1
}

// RandomVector is a slice of kyber.Scalar (encoded) which is used to prove
// and verify the proof of a shuffle
type RandomVector [][]byte

// Unmarshal returns the native type of a random vector
func (r RandomVector) Unmarshal() ([]kyber.Scalar, error) {
	e := make([]kyber.Scalar, len(r))

	for i, v := range r {
		scalar := suite.Scalar()
		err := scalar.UnmarshalBinary(v)
		if err != nil {
			return nil, xerrors.Errorf("cannot unmarshal form random vector: %v", err)
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
// valid.
func (c *Configuration) IsValid() bool {
	// serves as a set to check each ID is unique
	uniqueIDs := make(map[ID]bool)

	for _, subject := range c.Scaffold {
		if !subject.isValid(uniqueIDs) {
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

// CiphervotesFromPairs transforms two parallel lists of EGPoints to a list of
// Ciphervotes.
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

// ciphervoteFromPairs transforms two parallel lists of EGPoints to a list of
// ElGamal pairs.
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

// Pubshare represents a public share.
type Pubshare kyber.Point

// PubsharesUnit holds all the public shares produced by a given node,
// 1 for each ElGamal pair
type PubsharesUnit [][]Pubshare

// Fingerprint implements serde.Fingerprinter
func (p PubsharesUnit) Fingerprint(writer io.Writer) error {
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

// PubsharesUnits contains the pubshares submitted in parallel with the
// necessary data to identify the nodes who submitted them and their index.
type PubsharesUnits struct {
	// Pubshares holds the nodes' public shares
	Pubshares []PubsharesUnit
	// PubKeys contains the pubKey of the nodes who made each corresponding
	// PubsharesUnit
	PubKeys [][]byte
	// Indexes is the index of the nodes who made each corresponding
	// PubsharesUnit
	Indexes []int
}
