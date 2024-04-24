package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

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

// BallotsPerBatch to improve performance, so that (de)serializing only touches
// 100 ballots at a time.
var BallotsPerBatch = uint32(100)

// TestCastBallots if true, automatically fills every batch with ballots.
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

	// SuffragiaStoreKeys holds a slice of storage-keys to 0 or more Suffragia.
	// This is to optimize the time it takes to (De)serialize a Form.
	SuffragiaStoreKeys [][]byte

	// BallotCount is the total number of ballots cast, including double
	// ballots.
	BallotCount uint32

	// SuffragiaHashes holds a slice of hashes to all SuffragiaStoreKeys.
	// In case a Form has also to be proven to be correct outside the nodes,
	// the hashes are needed to prove the Suffragia are correct.
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

	// Store the list of admins SCIPER that are Owners of the form.
	Owners []int

	// Store the list of SCIPER of user that are Voters on the form.
	Voters []int
}

// Serialize implements serde.Message
func (form Form) Serialize(ctx serde.Context) ([]byte, error) {
	format := formFormat.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, form)
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
func (formFactory FormFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := formFormat.Get(ctx.GetFormat())

	ctx = serde.WithFactory(ctx, CiphervoteKey{}, formFactory.ciphervoteFac)
	ctx = serde.WithFactory(ctx, ctypes.RosterKey{}, formFactory.rosterFac)

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
func (form *Form) ChunksPerBallot() int {
	if form.BallotSize%29 == 0 {
		return form.BallotSize / 29
	}

	return form.BallotSize/29 + 1
}

// CastVote stores the new vote in the memory.
func (form *Form) CastVote(ctx serde.Context, st store.Snapshot, userID string, ciphervote Ciphervote) error {
	var suff Suffragia
	var batchID []byte

	if form.BallotCount%BallotsPerBatch == 0 {
		// Need to create a random ID for storing the ballots.
		// H( formID | ballotcount )
		// should be random enough, even if it's previsible.
		id, err := hex.DecodeString(form.FormID)
		if err != nil {
			return xerrors.Errorf("couldn't decode formID: %v", err)
		}
		h := sha256.New()
		h.Write(id)
		binary.LittleEndian.PutUint32(id, form.BallotCount)
		batchID = h.Sum(id[0:4])[:32]

		err = st.Set(batchID, []byte{})
		if err != nil {
			return xerrors.Errorf("couldn't store new ballot batch: %v", err)
		}
		form.SuffragiaStoreKeys = append(form.SuffragiaStoreKeys, batchID)
		form.SuffragiaHashes = append(form.SuffragiaHashes, []byte{})
	} else {
		batchID = form.SuffragiaStoreKeys[len(form.SuffragiaStoreKeys)-1]
		buf, err := st.Get(batchID)
		if err != nil {
			return xerrors.Errorf("couldn't get ballots batch: %v", err)
		}
		format := suffragiaFormat.Get(ctx.GetFormat())
		ctx = serde.WithFactory(ctx, CiphervoteKey{}, CiphervoteFactory{})
		msg, err := format.Decode(ctx, buf)
		if err != nil {
			return xerrors.Errorf("couldn't unmarshal ballots batch in cast: %v", err)
		}
		suff = msg.(Suffragia)
	}

	suff.CastVote(userID, ciphervote)
	if TestCastBallots {
		for i := uint32(1); i < BallotsPerBatch; i++ {
			suff.CastVote(fmt.Sprintf("%s-%d", userID, i), ciphervote)
		}
		form.BallotCount += BallotsPerBatch - 1
	}
	buf, err := suff.Serialize(ctx)
	if err != nil {
		return xerrors.Errorf("couldn't marshal ballots batch: %v", err)
	}
	err = st.Set(batchID, buf)
	if err != nil {
		xerrors.Errorf("couldn't set new ballots batch: %v", err)
	}
	form.BallotCount += 1
	return nil
}

// Suffragia returns all ballots from the storage. This should only
// be called rarely, as it might take a long time.
// It overwrites ballots cast by the same user and keeps only
// the latest ballot.
func (form *Form) Suffragia(ctx serde.Context, rd store.Readable) (Suffragia, error) {
	var suff Suffragia
	for _, id := range form.SuffragiaStoreKeys {
		buf, err := rd.Get(id)
		if err != nil {
			return suff, xerrors.Errorf("couldn't get ballot batch: %v", err)
		}
		format := suffragiaFormat.Get(ctx.GetFormat())
		ctx = serde.WithFactory(ctx, CiphervoteKey{}, CiphervoteFactory{})
		msg, err := format.Decode(ctx, buf)
		if err != nil {
			return suff, xerrors.Errorf("couldn't unmarshal ballots batch in cast: %v", err)
		}
		suffTmp := msg.(Suffragia)
		for i, uid := range suffTmp.VoterIDs {
			suff.CastVote(uid, suffTmp.Ciphervotes[i])
		}
	}
	return suff, nil
}

// RandomVector is a slice of kyber.Scalar (encoded) which is used to prove
// and verify the proof of a shuffle
type RandomVector [][]byte

// Unmarshal returns the native type of a random vector
func (randomVector RandomVector) Unmarshal() ([]kyber.Scalar, error) {
	e := make([]kyber.Scalar, len(randomVector))

	for i, v := range randomVector {
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
func (randomVector *RandomVector) LoadFromScalars(e []kyber.Scalar) error {
	*randomVector = make([][]byte, len(e))

	for i, scalar := range e {
		v, err := scalar.MarshalBinary()
		if err != nil {
			return xerrors.Errorf("could not marshal random vector: %v", err)
		}
		(*randomVector)[i] = v
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
func (configuration *Configuration) MaxBallotSize() int {
	size := 0
	for _, subject := range configuration.Scaffold {
		size += subject.MaxEncodedSize()
	}
	return size
}

// GetQuestion finds the question associated to a given ID and returns it.
// Returns nil if no question found.
func (configuration *Configuration) GetQuestion(ID ID) Question {
	for _, subject := range configuration.Scaffold {
		question := subject.GetQuestion(ID)

		if question != nil {
			return question
		}
	}

	return nil
}

// IsValid returns true if and only if the whole configuration is coherent and
// valid.
func (configuration *Configuration) IsValid() bool {
	// serves as a set to check each ID is unique
	uniqueIDs := make(map[ID]bool)

	for _, subject := range configuration.Scaffold {
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
func (pubshareUnit PubsharesUnit) Fingerprint(writer io.Writer) error {
	for _, ballotShares := range pubshareUnit {
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

// AddVoter add a new admin to the system.
func (form *Form) AddVoter(userID string) error {
	sciperInt, err := strconv.Atoi(userID)
	if err != nil {
		return xerrors.Errorf("Failed to convert SCIPER to an INT: %v", err)
	}

	form.Voters = append(form.Voters, sciperInt)

	return nil
}

// IsVoter return the index of admin if userID is one, else return -1
func (form *Form) IsVoter(userID string) int {
	sciperInt, err := strconv.Atoi(userID)
	if err != nil {
		return -1
	}

	for i := 0; i < len(form.Voters); i++ {
		if form.Voters[i] == sciperInt {
			return i
		}
	}

	return -1
}

// RemoveVoter add a new admin to the system.
func (form *Form) RemoveVoter(userID string) error {
	_, err := strconv.Atoi(userID)
	if err != nil {
		return xerrors.Errorf("Failed to convert SCIPER to an INT: %v", err)
	}

	index := form.IsVoter(userID)

	if index < 0 {
		return xerrors.Errorf("Error while retrieving the index of the element.")
	}

	form.Voters = append(form.Voters[:index], form.Voters[index+1:]...)
	return nil
}

// AddOwner add a new admin to the system.
func (form *Form) AddOwner(userID string) error {
	sciperInt, err := strconv.Atoi(userID)
	if err != nil {
		return xerrors.Errorf("Failed to convert SCIPER to an INT: %v", err)
	}

	// TODO need to check that the new user is admin !
	form.Owners = append(form.Owners, sciperInt)

	return nil
}

// IsOwner return the index of admin if userID is one, else return -1
func (form *Form) IsOwner(userID string) int {
	sciperInt, err := strconv.Atoi(userID)
	if err != nil {
		return -1
	}

	for i := 0; i < len(form.Owners); i++ {
		if form.Owners[i] == sciperInt {
			return i
		}
	}

	return -1
}

// RemoveOwner add a new admin to the system.
func (form *Form) RemoveOwner(userID string) error {
	_, err := strconv.Atoi(userID)
	if err != nil {
		return xerrors.Errorf("Failed to convert SCIPER to an INT: %v", err)
	}

	index := form.IsOwner(userID)

	if index < 0 {
		return xerrors.Errorf("Error while retrieving the index of the element.")
	}

	// We don't want to have a form without any Owners.
	if len(form.Owners) <= 1 {
		return xerrors.Errorf("Error, cannot remove this owner because it is the " +
			"only one remaining for this form")
	}

	form.Owners = append(form.Owners[:index], form.Owners[index+1:]...)
	return nil
}
