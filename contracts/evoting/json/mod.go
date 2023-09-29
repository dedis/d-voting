package json

import (
	"github.com/c4dt/d-voting/contracts/evoting/types"
	"github.com/c4dt/dela/serde"
)

// Register the JSON formats for the form, ciphervote, and transaction

func init() {
	types.RegisterFormFormat(serde.FormatJSON, formFormat{})
	types.RegisterCiphervoteFormat(serde.FormatJSON, ciphervoteFormat{})
	types.RegisterTransactionFormat(serde.FormatJSON, transactionFormat{})
}
