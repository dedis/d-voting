package fake

import (
	"context"
	"encoding/hex"

	electionTypes "github.com/dedis/d-voting/contracts/evoting/types"
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
	Elections map[string]electionTypes.Election
	Pool      *Pool
	Status    bool
	Channel   chan ordering.Event
	Context   serde.Context
}

// GetProof implements ordering.Service. It returns the proof associated to the
// election.
func (f Service) GetProof(key []byte) (ordering.Proof, error) {
	keyString := hex.EncodeToString(key)

	election, exists := f.Elections[keyString]
	if !exists {
		proof := Proof{
			key:   key,
			value: []byte(""),
		}
		return proof, f.Err
	}

	electionBuf, err := election.Serialize(f.Context)
	if err != nil {
		return nil, xerrors.Errorf("failed to serialize election: %v", err)
	}

	proof := Proof{
		key:   key,
		value: electionBuf,
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
func NewService(electionID string, election electionTypes.Election, ctx serde.Context) Service {
	elections := make(map[string]electionTypes.Election)
	elections[electionID] = election

	return Service{
		Err:       nil,
		Elections: elections,
		Context:   ctx,
	}
}
