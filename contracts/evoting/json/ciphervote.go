package json

import (
	"github.com/c4dt/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

// ciphervoteFormat is the JSON format to encode and decode a ciphervote.
//
// - implements serde.FormatEngine
type ciphervoteFormat struct{}

// Encode implements serde.FormatEngine
func (ciphervoteFormat) Encode(ctx serde.Context, msg serde.Message) ([]byte, error) {
	ciphervote, ok := msg.(types.Ciphervote)
	if !ok {
		return nil, xerrors.Errorf("unexpected type: %T", msg)
	}

	m := make(CiphervoteJSON, len(ciphervote))

	for i, egpair := range ciphervote {
		k, err := egpair.K.MarshalBinary()
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal k: %v", err)
		}

		c, err := egpair.C.MarshalBinary()
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal c: %v", err)
		}

		m[i] = EGPairJSON{
			K: k,
			C: c,
		}
	}

	data, err := ctx.Marshal(m)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal cipher vote json: %v", err)
	}

	return data, nil
}

// Decode implements serde.FormatEngine
func (ciphervoteFormat) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	var ciphervoteJSON CiphervoteJSON

	err := ctx.Unmarshal(data, &ciphervoteJSON)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal ciphervote json: %v", err)
	}

	ciphervote := make(types.Ciphervote, len(ciphervoteJSON))

	for i, egpair := range ciphervoteJSON {
		k := suite.Point()

		err = k.UnmarshalBinary(egpair.K)
		if err != nil {
			return nil, xerrors.Errorf("failed to unmarshal K: %v", err)
		}

		c := suite.Point()

		err = c.UnmarshalBinary(egpair.C)
		if err != nil {
			return nil, xerrors.Errorf("failed to unmarshal C: %v", err)
		}

		ciphervote[i] = types.EGPair{
			K: k,
			C: c,
		}
	}

	return ciphervote, nil
}

// CiphervoteJSON is the JSON representation of a ciphervote
type CiphervoteJSON []EGPairJSON

// EGPairJSON is the JSON representation of an ElGamal pair
type EGPairJSON struct {
	K []byte
	C []byte
}
