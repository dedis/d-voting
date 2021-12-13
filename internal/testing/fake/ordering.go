package fake

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	electionTypes "github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/validation"
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
	Elections map[string]interface{}
	Pool      *Pool
	Status    bool
	Channel   chan ordering.Event
}

// GetProof implements ordering.Service. It returns the proof associated to the election.
func (f Service) GetProof(key []byte) (ordering.Proof, error) {
	// electionIDBuf, _ := hex.DecodeString(string(f.ElectionID))

	keyString := hex.EncodeToString(key)

	election, exists := f.Elections[keyString]
	if !exists {
		proof := Proof{
			key:   key,
			value: []byte(""),
		}
		return proof, f.Err
	}

	js, _ := json.Marshal(election)
	proof := Proof{
		key:   key,
		value: js,
	}

	return proof, f.Err
}

// GetStore implements ordering.Service. It returns the store associated to the service.
func (f Service) GetStore() store.Readable {
	return nil
}

// Watch implements ordering.Service. It returns the events that occurred within the service.
func (f Service) Watch(ctx context.Context) <-chan ordering.Event {

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
		tx:      f.Pool.Transaction,
	}

	f.Status = true

	channel := make(chan ordering.Event, 1)
	fmt.Println("watch", results[0])
	channel <- ordering.Event{
		Index:        0,
		Transactions: results,
	}
	close(channel)

	return channel

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
		tx:      f.Pool.Transaction,
	}

	f.Status = true

	fmt.Println("watch", results[0])
	f.Channel <- ordering.Event{
		Index:        0,
		Transactions: results,
	}
	close(f.Channel)

}

func NewService(election electionTypes.Election, electionID string) Service {
	elections := make(map[string]interface{})
	elections[electionID] = election

	return Service{
		Err:       nil,
		Elections: elections,
	}
}
