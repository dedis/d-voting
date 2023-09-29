package types

import (
	"fmt"
	"io"

	"github.com/c4dt/dela/serde"
	"github.com/c4dt/dela/serde/registry"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

var (
	ciphervoteFormats = registry.NewSimpleRegistry()
)

// RegisterCiphervoteFormat registers the engine for the provided format
func RegisterCiphervoteFormat(f serde.Format, e serde.FormatEngine) {
	ciphervoteFormats.Register(f, e)
}

// Ciphervote represents an encrypted vote. It consists of a list of ElGamal
// pairs.
//
// - implements serde.Message
// - implements serde.Fingerprinter
type Ciphervote []EGPair

// Serialize implements serde.Message
func (c Ciphervote) Serialize(ctx serde.Context) ([]byte, error) {
	format := ciphervoteFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, c)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode encrypted ballot: %v", err)
	}

	return data, nil
}

// Copy returns a copy of a ciphervote
func (c Ciphervote) Copy() Ciphervote {
	ciphervote := make(Ciphervote, len(c))

	for i, egpair := range c {
		ciphervote[i] = egpair.Copy()
	}

	return ciphervote
}

// FingerPrint implements serde.Fingerprinter
func (c Ciphervote) FingerPrint(writer io.Writer) error {
	for _, egpair := range c {
		_, err := egpair.K.MarshalTo(writer)
		if err != nil {
			return xerrors.Errorf("failed to marshal K: %v", err)
		}

		_, err = egpair.C.MarshalTo(writer)
		if err != nil {
			return xerrors.Errorf("failed to marshal C: %v", err)
		}
	}

	return nil
}

// GetElGPairs returns corresponding kyber.Points from the ciphertexts
func (c Ciphervote) GetElGPairs() (ks []kyber.Point, cs []kyber.Point) {
	ks = make([]kyber.Point, len(c))
	cs = make([]kyber.Point, len(c))

	for i, egpair := range c {
		ks[i] = egpair.K
		cs[i] = egpair.C
	}

	return ks, cs
}

// Equal returns if the other ciphervote is equal
func (c Ciphervote) Equal(other Ciphervote) bool {
	if len(c) != len(other) {
		return false
	}

	for i, e := range c {
		if !e.K.Equal(other[i].K) || !e.C.Equal(other[i].C) {
			return false
		}
	}

	return true
}

// EGPair defines an ElGamal pair.
type EGPair struct {
	K kyber.Point
	C kyber.Point
}

// Copy returns a copy of EGPair
func (ct EGPair) Copy() EGPair {
	return EGPair{
		K: ct.K.Clone(),
		C: ct.C.Clone(),
	}
}

// String returns a string representation of an ElGamal pair
func (ct EGPair) String() string {
	return fmt.Sprintf("{K: %s, C: %s}", ct.K.String(), ct.C.String())
}

// CiphervoteKey is the factory key for Ciphervote
type CiphervoteKey struct{}

// CiphervoteFactory provides the mean to deserialize a ciphervote. It naturally
// uses the formFormat.
//
// - implements serde.Factory
type CiphervoteFactory struct{}

// Deserialize implements serde.Factory
func (CiphervoteFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := ciphervoteFormats.Get(ctx.GetFormat())

	message, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode: %v", err)
	}

	return message, nil
}

// CiphervotesToPairs converts a slice of ciphervote to X, Y
func CiphervotesToPairs(ciphervotes []Ciphervote) (X [][]kyber.Point, Y [][]kyber.Point) {
	seqSize := len(ciphervotes[0])

	X = make([][]kyber.Point, seqSize)
	Y = make([][]kyber.Point, seqSize)

	for _, ciphervote := range ciphervotes {

		x, y := ciphervote.GetElGPairs()

		for i := 0; i < seqSize; i++ {
			X[i] = append(X[i], x[i])
			Y[i] = append(Y[i], y[i])
		}
	}

	return X, Y
}
