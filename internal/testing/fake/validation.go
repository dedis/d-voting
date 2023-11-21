package fake

import (
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/validation"
	"go.dedis.ch/dela/serde"
)

// TransactionResult is a fake implementation of TransactionResult.
//
// - implements validation.TransactionResult
type TransactionResult struct {
	status  bool
	message string
	tx      Transaction
}

// Serialize implements TransactionResult.
func (f TransactionResult) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

// GetTransaction implements TransactionResult. It returns the transaction
// associated to the result.
func (f TransactionResult) GetTransaction() txn.Transaction {
	return f.tx
}

// GetStatus implements TransactionResult. It returns the status of the
// execution.
func (f TransactionResult) GetStatus() (bool, string) {
	return f.status, f.message
}

// ValidationService is a fake implementation of Service
//
// - implements validation.Service
type ValidationService struct {
	Execution execution.Service
	Fac       validation.ResultFactory
}

// GetFactory returns the result factory.
func (v ValidationService) GetFactory() validation.ResultFactory {
	return nil
}

// GetNonce returns the nonce associated with the identity. The value
// returned should be used for the next transaction to be valid.
func (v ValidationService) GetNonce(r store.Readable, i access.Identity) (uint64, error) {
	return 0, nil
}

// Accept returns nil if the transaction will be accepted by the service.
// The leeway parameter allows to reduce some constraints.
func (v ValidationService) Accept(r store.Readable, t txn.Transaction, l validation.Leeway) error {
	return nil
}

// Validate takes a snapshot and a list of transactions and returns a
// result.
func (v ValidationService) Validate(s store.Snapshot, t []txn.Transaction) (validation.Result, error) {
	return nil, nil
}
