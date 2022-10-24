package fake

import (
	"context"
	"encoding/hex"

	formTypes "github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/validation"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

// Proof
//
// - implements ordering.Proof
type Proof struct {
	key   []byte
	value []byte
}

// GetKey implements ordering.Proof. It returns the key associated to the proof.
func (f Proof) GetKey() []byte {
	return f.key
}

// GetValue implements ordering.Proof. It returns the value associated to the
// proof if the key exists, otherwise it returns nil.
func (f Proof) GetValue() []byte {
	return f.value
}

// Service
//
// - implements ordering.Service
type Service struct {
	Err       error
	Forms map[string]formTypes.Form
	Pool      *Pool
	Status    bool
	Channel   chan ordering.Event
	Context   serde.Context
}

// GetProof implements ordering.Service. It returns the proof associated to the
// form.
func (f Service) GetProof(key []byte) (ordering.Proof, error) {
	keyString := hex.EncodeToString(key)

	form, exists := f.Forms[keyString]
	if !exists {
		proof := Proof{
			key:   key,
			value: []byte(""),
		}
		return proof, f.Err
	}

	formBuf, err := form.Serialize(f.Context)
	if err != nil {
		return nil, xerrors.Errorf("failed to serialize form: %v", err)
	}

	proof := Proof{
		key:   key,
		value: formBuf,
	}

	return proof, f.Err
}

// GetStore implements ordering.Service. It returns the store associated to the
// service.
func (f Service) GetStore() store.Readable {
	return nil
}

// Watch implements ordering.Service. It returns the events that occurred within
// the service.
func (f *Service) Watch(ctx context.Context) <-chan ordering.Event {
	f.Channel = make(chan ordering.Event, 100)
	return f.Channel

}

func (f Service) Close() error {
	return f.Err
}

func (f *Service) AddTx(tx Transaction) {
	results := make([]validation.TransactionResult, 3)

	results[0] = TransactionResult{
		status:  true,
		message: "",
		tx:      Transaction{Nonce: 10, Id: []byte("dummyId1")},
	}

	results[1] = TransactionResult{
		status:  true,
		message: "",
		tx:      Transaction{Nonce: 11, Id: []byte("dummyId2")},
	}

	results[2] = TransactionResult{
		status:  f.Status,
		message: "",
		tx:      tx,
	}

	f.Status = true

	f.Channel <- ordering.Event{
		Index:        0,
		Transactions: results,
	}
	close(f.Channel)

}

// NewService returns a new initialized service
func NewService(formID string, form formTypes.Form, ctx serde.Context) Service {
	forms := make(map[string]formTypes.Form)
	forms[formID] = form

	return Service{
		Err:       nil,
		Forms: forms,
		Context:   ctx,
	}
}
