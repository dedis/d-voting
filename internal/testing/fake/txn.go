package fake

import (
	"context"
	"io"

	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
)

// Transaction is a fake implementation of Transaction.
//
// - implements txn.Transaction
type Transaction struct {
	nonce uint64
	id    []byte
}

func (f Transaction) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

func (f Transaction) Fingerprint(writer io.Writer) error {
	return nil
}

func (f Transaction) GetID() []byte {
	return f.id
}

func (f Transaction) GetNonce() uint64 {
	return f.nonce
}

func (f Transaction) GetIdentity() access.Identity {
	return nil
}

func (f Transaction) GetArg(key string) []byte {
	return nil
}

// Pool is a fake implementation of Pool.
//
// - implements txn.pool.Pool
type Pool struct {
	err         error
	transaction Transaction
}

func (f Pool) SetPlayers(players mino.Players) error {
	return nil
}

func (f Pool) AddFilter(filter pool.Filter) {
}

func (f Pool) Len() int {
	return 0
}

func (f *Pool) Add(transaction txn.Transaction) error {
	newTx := Transaction{
		nonce: transaction.GetNonce(),
		id:    transaction.GetID(),
	}

	f.transaction = newTx
	return f.err
}

func (f Pool) Remove(transaction txn.Transaction) error {
	return nil
}

func (f Pool) Gather(ctx context.Context, config pool.Config) []txn.Transaction {
	return nil
}

func (f Pool) Close() error {
	return nil
}
