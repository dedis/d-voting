package types

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"strconv"

	"go.dedis.ch/dela/serde"
	"go.dedis.ch/dela/serde/registry"
	"golang.org/x/xerrors"
)

var transactionFormats = registry.NewSimpleRegistry()

// TransactionKey is the key for the transaction factory
type TransactionKey struct{}

// RegisterTransactionFormat registers the engine for the provided format
func RegisterTransactionFormat(f serde.Format, e serde.FormatEngine) {
	transactionFormats.Register(f, e)
}

// FormsMetadata ...
type FormsMetadata struct {
	FormsIDs FormIDs
}

// FormIDs is a slice of hex-encoded form IDs
type FormIDs []string

// Contains checks if el is present. Return < 0 if not.
func (formID FormIDs) Contains(el string) int {
	for i, e1 := range formID {
		if e1 == el {
			return i
		}
	}

	return -1
}

// Add adds a form ID or returns an error if already present
func (formID *FormIDs) Add(id string) error {
	if formID.Contains(id) >= 0 {
		return xerrors.Errorf("id %q already exist", id)
	}

	*formID = append(*formID, id)

	return nil
}

// Remove removes a form ID from the list, if it exists
func (formID *FormIDs) Remove(id string) {
	i := formID.Contains(id)
	if i >= 0 {
		*formID = append((*formID)[:i], (*formID)[i+1:]...)
	}
}

// TransactionFactory provides the mean to deserialize a transaction.
//
// - implements serde.Factory
type TransactionFactory struct {
	ciphervoteFac serde.Factory
}

// NewTransactionFactory creates a new transaction factory
func NewTransactionFactory(cf serde.Factory) TransactionFactory {
	return TransactionFactory{
		ciphervoteFac: cf,
	}
}

// Deserialize implements serde.Factory
func (transactionFactory TransactionFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	ctx = serde.WithFactory(ctx, CiphervoteKey{}, transactionFactory.ciphervoteFac)

	message, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode: %v", err)
	}

	return message, nil
}

// CreateForm defines the transaction to create a form
//
// - implements serde.Message
type CreateForm struct {
	Configuration Configuration
	UserID        string
}

// Serialize implements serde.Message
func (createForm CreateForm) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, createForm)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode create form: %v", err)
	}

	return data, nil
}

// OpenForm defines the transaction to open a form
//
// - implements serde.Message
type OpenForm struct {
	// FormID is hex-encoded
	FormID string
}

// Serialize implements serde.Message
func (openForm OpenForm) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, openForm)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode open form: %v", err)
	}

	return data, nil
}

// CastVote defines the transaction to cast a vote
//
// - implements serde.Message
type CastVote struct {
	// FormID is hex-encoded
	FormID  string
	VoterID string
	Ballot  Ciphervote
}

// Serialize implements serde.Message
func (castVote CastVote) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, castVote)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode cast vote: %v", err)
	}

	return data, nil
}

// CloseForm defines the transaction to close a form
//
// - implements serde.Message
type CloseForm struct {
	// FormID is hex-encoded
	FormID string
	UserID string
}

// Serialize implements serde.Message
func (closeForm CloseForm) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, closeForm)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode close form: %v", err)
	}

	return data, nil
}

// ShuffleBallots defines the transaction to shuffle the ballots
//
// - implements serde.Message
// - implements serde.Fingerprinter
type ShuffleBallots struct {
	FormID          string
	Round           int
	ShuffledBallots []Ciphervote
	// RandomVector is the vector to be used to generate the proof of the next
	// shuffle
	RandomVector RandomVector
	// Proof is the proof corresponding to the shuffle of this transaction
	Proof []byte
	// Signature is the signature of the result of HashShuffle() with the private
	// key corresponding to PublicKey
	Signature []byte
	//PublicKey is the public key of the signer.
	PublicKey []byte
}

// Serialize implements serde.Message
func (shuffleBallots ShuffleBallots) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, shuffleBallots)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode shuffle ballots: %v", err)
	}

	return data, nil
}

// RegisterPubShares defines the transaction used by a node to send its
// pubshares on the chain.
//
// - implements serde.Message
type RegisterPubShares struct {
	FormID string
	// Index is the index of the node making the submission
	Index int
	// Pubshares are the public shares of the node submitting the transaction
	// so that they can be used for decryption.
	Pubshares PubsharesUnit
	// Signature is the signature of the result of HashPubShares() with the
	// private key corresponding to PublicKey
	Signature []byte
	// PublicKey is the public key of the signer
	PublicKey []byte
}

// Serialize implements serde.Message
func (registerPubShares RegisterPubShares) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, registerPubShares)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode register pubShares: %v", err)
	}

	return data, nil
}

// CombineShares defines the transaction to decrypt the ballots by combining all
// the public shares.
//
// - implements serde.Message
type CombineShares struct {
	// FormID is hex-encoded
	FormID string
	UserID string
}

// Serialize implements serde.Message
func (combineShares CombineShares) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, combineShares)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode decrypt ballot: %v", err)
	}

	return data, nil
}

// CancelForm defines the transaction to cancel the form
//
// - implements serde.Message
type CancelForm struct {
	// FormID is hex-encoded
	FormID string
	UserID string
}

// Serialize implements serde.Message
func (cancelForm CancelForm) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, cancelForm)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode cancel form: %v", err)
	}

	return data, nil
}

// DeleteForm defines the transaction to delete the form
//
// - implements serde.Message
type DeleteForm struct {
	// FormID is hex-encoded
	FormID string
}

// Serialize implements serde.Message
func (deleteForm DeleteForm) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, deleteForm)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode cancel form: %v", err)
	}

	return data, nil
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

// Fingerprint implements serde.Fingerprinter. If creates a fingerprint only
// based on the formID and the shuffled ballots.
func (shuffleBallots ShuffleBallots) Fingerprint(writer io.Writer) error {
	_, err := writer.Write([]byte(shuffleBallots.FormID))
	if err != nil {
		return xerrors.Errorf("failed to write the form ID: %v", err)
	}

	for _, ballot := range shuffleBallots.ShuffledBallots {
		err := ballot.FingerPrint(writer)
		if err != nil {
			return xerrors.Errorf("failed to fingerprint shuffled ballot: %v", err)
		}
	}

	return nil
}

// Fingerprint implements serde.Fingerprinter
func (registerPubShares RegisterPubShares) Fingerprint(writer io.Writer) error {
	_, err := writer.Write([]byte(registerPubShares.FormID))
	if err != nil {
		return xerrors.Errorf("failed to write the form ID: %v", err)
	}

	_, err = writer.Write([]byte(strconv.Itoa(registerPubShares.Index)))
	if err != nil {
		return xerrors.Errorf("failed to write the pubShare index: %v", err)
	}

	err = registerPubShares.Pubshares.Fingerprint(writer)
	if err != nil {
		return xerrors.Errorf("failed to fingerprint pubShares: %V", err)
	}

	return nil
}

// AddAdmin defines the transaction to Add an Admin
//
// - implements serde.Message
type AddAdmin struct {
	// FormID is hex-encoded
	FormID string
	UserID string
}

// Serialize implements serde.Message
func (addAdmin AddAdmin) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, addAdmin)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode Add Admin: %v", err)
	}

	return data, nil
}

// RemoveAdmin defines the transaction to Remove an Admin
//
// - implements serde.Message
type RemoveAdmin struct {
	// FormID is hex-encoded
	FormID string
	UserID string
}

// Serialize implements serde.Message
func (removeAdmin RemoveAdmin) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, removeAdmin)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode remove admin: %v", err)
	}

	return data, nil
}
