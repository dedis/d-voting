package types

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

// TransactionInfo defines the information of a transaction
type TransactionInfo struct {
	Status        TransactionStatus // 0 if not yet included, 1 if included, 2 if rejected
	TransactionID []byte
	LastBlockIdx  uint64 // last block of the chain when the transaction was added to the pool
	Time          int64  // time when the transaction was added to the pool
	Hash          []byte // signature of the transaction
	Signature     []byte // signature of the transaction
}

// TransactionInfoToSend defines the HTTP response when sending
// transaction infos to the client so that he can use the status
// of the transaction to know if it has been included or not
// and if it has not been included, he can just use the token
// and ask again later
type TransactionInfoToSend struct {
	Status TransactionStatus // 0 if not yet included, 1 if included, 2 if rejected
	Token  string
}
