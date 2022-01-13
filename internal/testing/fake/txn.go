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
	Nonce uint64
	Id    []byte
}

func (f Transaction) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

func (f Transaction) Fingerprint(writer io.Writer) error {
	return nil
}

func (f Transaction) GetID() []byte {
	return f.Id
}

func (f Transaction) GetNonce() uint64 {
	return f.Nonce
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
	Err         error
	Transaction Transaction
	Service     *Service
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
		Nonce: transaction.GetNonce(),
		Id:    transaction.GetID(),
	}

	f.Transaction = newTx
	f.Service.AddTx(newTx)

	return f.Err
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
