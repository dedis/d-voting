package json

import (
	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
)

// Register the JSON formats for the form, ciphervote, and transaction

func init() {
	types.RegisterFormFormat(serde.FormatJSON, formFormat{})
	types.RegisterSuffragiaFormat(serde.FormatJSON, suffragiaFormat{})
	types.RegisterCiphervoteFormat(serde.FormatJSON, ciphervoteFormat{})
	types.RegisterTransactionFormat(serde.FormatJSON, transactionFormat{})
}
