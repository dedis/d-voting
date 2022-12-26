package txnmanager

import (
	"context"
	"net/http"

	"github.com/dedis/d-voting/contracts/evoting"
)

// Manager defines the public HTTP API of the transaction manager
type Manager interface {
	// GET /transactions/{token}
	StatusHandlerGet(http.ResponseWriter, *http.Request)

	// submit the transaction to the blockchain
	// return the transactionID and 
	// the index of the last block when it was submitted
	SubmitTxn(ctx context.Context, cmd evoting.Command, cmdArg string, payload []byte) ([]byte, uint64, error)

	// CreateTransactionResult create the json to send to the client 
	CreateTransactionResult(txnID []byte, lastBlockIdx uint64, status TransactionStatus) (TransactionClientInfo, error)
	SendTransactionInfo(w http.ResponseWriter, txnID []byte, lastBlockIdx uint64, status TransactionStatus) error
}

// TransactionStatus is the status of a transaction
type TransactionStatus byte

const (
	// UnknownTransactionStatus is the basic status of a transaction
	UnknownTransactionStatus TransactionStatus = 0
	// IncludedTransaction is the status of a transaction that has been included
	IncludedTransaction TransactionStatus = 1
	// RejectedTransaction is the status of a transaction will never be included
	RejectedTransaction TransactionStatus = 2
)

// transactionInternalInfo defines the information of a transaction
type transactionInternalInfo struct {
	Status        TransactionStatus // 0 if not yet included, 1 if included, 2 if rejected
	TransactionID []byte
	LastBlockIdx  uint64 // last block of the chain when the transaction was added to the pool
	Time          int64  // time when the transaction was added to the pool
	Hash          []byte // signature of the transaction
	Signature     []byte // signature of the transaction
}



// TransactionClientInfo defines the HTTP response when sending
// transaction infos to the client so that he can use the status
// of the transaction to know if it has been included or not
// and if it has not been included, he can just use the token
// and ask again later
type TransactionClientInfo struct {
	Status TransactionStatus // 0 if not yet included, 1 if included, 2 if rejected
	Token  string
}
