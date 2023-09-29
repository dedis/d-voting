package fake

import (
	"context"
	"io"

	"github.com/c4dt/dela/core/access"
	"github.com/c4dt/dela/core/txn"
	"github.com/c4dt/dela/core/txn/pool"
	"github.com/c4dt/dela/mino"
	"github.com/c4dt/dela/serde"
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

// Manager is a fake manager
//
// - implements txn.Manager
type Manager struct {
	txn.Manager
}

func (m Manager) Sync() error {
	return nil
}

func (m Manager) Make(args ...txn.Arg) (txn.Transaction, error) {
	return nil, GetError()
}
