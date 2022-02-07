package json

import (
	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
)

func init() {
	types.RegisterElectionFormat(serde.FormatJSON, electionFormat{})
	types.RegisterCiphervoteFormat(serde.FormatJSON, ciphervoteFormat{})
	types.RegisterTransactionFormat(serde.FormatJSON, transactionFormat{})
}
