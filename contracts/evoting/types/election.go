package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	ctypes "go.dedis.ch/dela/core/ordering/cosipbft/types"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/dela/serde/registry"
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
	// Canceled is when the form has been canceled
	Canceled Status = 6
)

// BallotsPerBlock to improve performance, so that (de)serializing only touches
// 100 ballots at a time.
var BallotsPerBlock = uint32(100)

// TestCastBallots if true, automatically fills every block with ballots.
var TestCastBallots = false

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

	// SuffragiaIDs holds a slice of IDs to slices of SuffragiaIDs.
	// This is to optimize the time it takes to (De)serialize a Form.
	SuffragiaIDs [][]byte

	// BallotCount is the total number of ballots cast, including double
	// ballots.
	BallotCount uint32

	// SuffragiaHashes holds a slice of hashes to all SuffragiaIDs.
	// LG: not really sure if this is needed. In case a Form has also to be
	// proven to be correct outside the nodes, the hashes are definitely
	// needed.
	SuffragiaHashes [][]byte

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

// FormFromStore returns a form from the store given the formIDHex.
// An error indicates a wrong storage of the form.
func FormFromStore(ctx serde.Context, formFac serde.Factory, formIDHex string,
	store store.Readable) (Form, error) {

	form := Form{}

	formIDBuf, err := hex.DecodeString(formIDHex)
	if err != nil {
		return form, xerrors.Errorf("failed to decode formIDHex: %v", err)
	}

	formBuff, err := store.Get(formIDBuf)
	if err != nil {
		return form, xerrors.Errorf("while getting data for form: %v", err)
	}
	if len(formBuff) == 0 {
		return form, xerrors.Errorf("no form found")
	}

	message, err := formFac.Deserialize(ctx, formBuff)
	if err != nil {
		return form, xerrors.Errorf("failed to deserialize Form: %v", err)
	}

	form, ok := message.(Form)
	if !ok {
		return form, xerrors.Errorf("wrong message type: %T", message)
	}

	return form, nil
}

// ChunksPerBallot returns the number of chunks of El Gamal pairs needed to
// represent an encrypted ballot, knowing that one chunk is 29 bytes at most.
func (e *Form) ChunksPerBallot() int {
	if e.BallotSize%29 == 0 {
		return e.BallotSize / 29
	}

	return e.BallotSize/29 + 1
}

// CastVote stores the new vote in the memory.
func (s *Form) CastVote(ctx serde.Context, st store.Snapshot, userID string, ciphervote Ciphervote) error {
	var suff Suffragia
	var blockID []byte
	if s.BallotCount%BallotsPerBlock == 0 {
		// Need to create a random ID for storing the ballots.
		// H( formID | ballotcount )
		// should be random enough, even if it's previsible.
		id, err := hex.DecodeString(s.FormID)
		if err != nil {
			return xerrors.Errorf("couldn't decode formID: %v", err)
		}
		h := sha256.New()
		h.Write(id)
		binary.LittleEndian.PutUint32(id, s.BallotCount)
		blockID = h.Sum(id[0:4])[:32]
		err = st.Set(blockID, []byte{})
		if err != nil {
			return xerrors.Errorf("couldn't store new ballot block: %v", err)
		}
		s.SuffragiaIDs = append(s.SuffragiaIDs, blockID)
		s.SuffragiaHashes = append(s.SuffragiaHashes, []byte{})
	} else {
		blockID = s.SuffragiaIDs[len(s.SuffragiaIDs)-1]
		buf, err := st.Get(blockID)
		if err != nil {
			return xerrors.Errorf("couldn't get ballots block: %v", err)
		}
		format := suffragiaFormat.Get(ctx.GetFormat())
		ctx = serde.WithFactory(ctx, CiphervoteKey{}, CiphervoteFactory{})
		msg, err := format.Decode(ctx, buf)
		if err != nil {
			return xerrors.Errorf("couldn't unmarshal ballots block in cast: %v", err)
		}
		suff = msg.(Suffragia)
	}

	suff.CastVote(userID, ciphervote)
	if TestCastBallots {
		for i := uint32(1); i < BallotsPerBlock; i++ {
			suff.CastVote(fmt.Sprintf("%s-%d", userID, i), ciphervote)
		}
		s.BallotCount += BallotsPerBlock - 1
	}
	buf, err := suff.Serialize(ctx)
	if err != nil {
		return xerrors.Errorf("couldn't marshal ballots block: %v", err)
	}
	err = st.Set(blockID, buf)
	if err != nil {
		xerrors.Errorf("couldn't set new ballots block: %v", err)
	}
	s.BallotCount += 1
	return nil
}

// Suffragia returns all ballots from the storage. This should only
// be called rarely, as it might take a long time.
// It overwrites ballots cast by the same user and keeps only
// the latest ballot.
func (s *Form) Suffragia(ctx serde.Context, rd store.Readable) (Suffragia, error) {
	var suff Suffragia
	for _, id := range s.SuffragiaIDs {
		buf, err := rd.Get(id)
		if err != nil {
			return suff, xerrors.Errorf("couldn't get ballot block: %v", err)
		}
		format := suffragiaFormat.Get(ctx.GetFormat())
		ctx = serde.WithFactory(ctx, CiphervoteKey{}, CiphervoteFactory{})
		msg, err := format.Decode(ctx, buf)
		if err != nil {
			return suff, xerrors.Errorf("couldn't unmarshal ballots block in cast: %v", err)
		}
		suffTmp := msg.(Suffragia)
		for i, uid := range suffTmp.UserIDs {
			suff.CastVote(uid, suffTmp.Ciphervotes[i])
		}
	}
	return suff, nil
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
	Title          Title
	Scaffold       []Subject
	AdditionalInfo string
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
