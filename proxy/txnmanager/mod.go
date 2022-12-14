package txnmanager

import (
	"context"
	"net/http"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/proxy/types"
)

// TxnManager defines the public HTTP API of the transaction service
type Manager interface {
	// GET /transactions/{token}
	IsTxnIncluded(http.ResponseWriter, *http.Request)
	submitTxn(ctx context.Context, cmd evoting.Command, cmdArg string, payload []byte) ([]byte, uint64, error)
	CreateTransactionInfoToSend(txnID []byte, lastBlockIdx uint64, status types.TransactionStatus) (types.TransactionInfoToSend, error)
	sendTransactionInfo(w http.ResponseWriter, txnID []byte, lastBlockIdx uint64, status types.TransactionStatus) error

}