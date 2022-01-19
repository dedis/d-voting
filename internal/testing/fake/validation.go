package fake

import (
	"go.dedis.ch/dela/core/txn"
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

// GetTransaction implements TransactionResult.
func (f TransactionResult) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

// GetTransaction implements TransactionResult. It returns the transaction associated to the result.
func (f TransactionResult) GetTransaction() txn.Transaction {
	return f.tx
}

// GetStatus implements TransactionResult. It returns the status of the execution.
func (f TransactionResult) GetStatus() (bool, string) {
	return f.status, f.message
}
