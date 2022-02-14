package json

import (
	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
)

// Register the JSON formats for the election, ciphervote, and transaction

func init() {
	types.RegisterElectionFormat(serde.FormatJSON, electionFormat{})
	types.RegisterCiphervoteFormat(serde.FormatJSON, ciphervoteFormat{})
	types.RegisterTransactionFormat(serde.FormatJSON, transactionFormat{})
}
