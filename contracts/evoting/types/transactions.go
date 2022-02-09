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

// ElectionsMetadata ...
type ElectionsMetadata struct {
	ElectionsIDs ElectionIDs
}

// ElectionIDs is a slice of hex-encoded election IDs
type ElectionIDs []string

// Contains checks if el is present
func (e ElectionIDs) Contains(el string) bool {
	for _, e1 := range e {
		if e1 == el {
			return true
		}
	}

	return false
}

// Add adds an election ID or returns an error if already present
func (e *ElectionIDs) Add(id string) error {
	if e.Contains(id) {
		return xerrors.Errorf("id %q already exist", id)
	}

	*e = append(*e, id)

	return nil
}

// TransactionFactory provides the mean to deserialize a transaction.
//
// - implements serde.Factory
type TransactionFactory struct{}

// Deserialize implements serde.Factory
func (TransactionFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	message, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode: %v", err)
	}

	return message, nil
}

// CreateElection ...
//
// - implements serde.Message
type CreateElection struct {
	Configuration Configuration
	AdminID       string
}

// Serialize implements serde.Message
func (ce CreateElection) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, ce)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode create election: %v", err)
	}

	return data, nil
}

// OpenElection ...
//
// - implements serde.Message
type OpenElection struct {
	// ElectionID is hex-encoded
	ElectionID string
}

// Serialize implements serde.Message
func (oe OpenElection) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, oe)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode open election: %v", err)
	}

	return data, nil
}

// CastVote ...
//
// - implements serde.Message
type CastVote struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
	Ballot     Ciphervote
}

// Serialize implements serde.Message
func (cv CastVote) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, cv)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode cast vote: %v", err)
	}

	return data, nil
}

// CloseElection ...
//
// - implements serde.Message
type CloseElection struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
}

// Serialize implements serde.Message
func (ce CloseElection) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, ce)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode close election: %v", err)
	}

	return data, nil
}

// ShuffleBallots ...
//
// - implements serde.Message
// - implements serde.Fingerprinter
type ShuffleBallots struct {
	ElectionID      string
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
func (sb ShuffleBallots) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, sb)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode shuffle ballots: %v", err)
	}

	return data, nil
}

type RegisterPubShares struct {
	ElectionID string
	// Index is the index of the node making the submission
	Index int
	// PubShares are the public shares of the node submitting the transaction
	// so that they can be used for decryption.
	PubShares PubSharesSubmission
	// Signature is the signature of the result of HashPubShares() with the
	// private key corresponding to PublicKey
	Signature []byte
}

// Serialize implements serde.Message
func (rp RegisterPubShares) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, rp)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode register pubShares: %v", err)
	}

	return data, nil
}

// DecryptBallots ...
//
// - implements serde.Message
type DecryptBallots struct {
	// ElectionID is hex-encoded
	ElectionID       string
	UserID           string
	DecryptedBallots []Ballot
}

// Serialize implements serde.Message
func (db DecryptBallots) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, db)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode decrypt ballot: %v", err)
	}

	return data, nil
}

// CancelElection ...
//
// - implements serde.Message
type CancelElection struct {
	// ElectionID is hex-encoded
	ElectionID string
	UserID     string
}

// Serialize implements serde.Message
func (ce CancelElection) Serialize(ctx serde.Context) ([]byte, error) {
	format := transactionFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, ce)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode cancel election: %v", err)
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

// Fingerprint implements serde.Fingerprinter
func (sb ShuffleBallots) Fingerprint(writer io.Writer) error {
	_, err := writer.Write([]byte(sb.ElectionID))
	if err != nil {
		return xerrors.Errorf("failed to write election ID to fingerprint: %v", err)
	}

	for _, ballot := range sb.ShuffledBallots {
		err := ballot.FingerPrint(writer)
		if err != nil {
			return xerrors.Errorf("failed to fingerprint shuffled ballot: %v", err)
		}
	}

	return nil
}

// Fingerprint implements serde.Fingerprinter
func (ps RegisterPubShares) Fingerprint(writer io.Writer) error {
	_, err := writer.Write([]byte(ps.ElectionID))
	if err != nil {
		return xerrors.Errorf("failed to write election ID to fingerprint: %v", err)
	}

	_, err = writer.Write([]byte(strconv.Itoa(ps.Index)))
	if err != nil {
		return xerrors.Errorf("failed to write pubShare index to fingerprint: %v", err)
	}

	err = ps.PubShares.FingerPrint(writer)
	if err != nil {
		return xerrors.Errorf("failed to fingerprint pubShares: %V", err)
	}

	return nil
}
