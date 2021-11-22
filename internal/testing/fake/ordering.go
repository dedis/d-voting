package fake

import (
		"bytes"
		"context"
		"fmt"
		"encoding/json"
		"encoding/hex"

		"github.com/dedis/d-voting/contracts/evoting/types"
		"go.dedis.ch/dela/core/store"
		"go.dedis.ch/dela/core/ordering"
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
		err        error
		election   interface{}
		electionID types.ID
		pool       *Pool
		status     bool
}

// GetProof implements ordering.Service. It returns the proof associated to the election.
func (f Service) GetProof(key []byte) (ordering.Proof, error) {
		electionIDBuf, _ := hex.DecodeString(string(f.electionID))

		if bytes.Equal(key, electionIDBuf) {
				js, _ := json.Marshal(f.election)
				proof := Proof{
						key:   key,
						value: js,
				}
				return proof, f.err
		}

		return nil, f.err
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
				tx:      Transaction{nonce: 10, id: []byte("dummyId1")},
		}

		results[1] = TransactionResult{
				status:  true,
				message: "",
				tx:      Transaction{nonce: 11, id: []byte("dummyId2")},
		}

		results[2] = TransactionResult{
				status:  f.status,
				message: "",
				tx:      f.pool.transaction,
		}

		f.status = true

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
	return f.err
}
