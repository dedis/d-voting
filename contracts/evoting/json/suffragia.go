package json

import (
	"encoding/json"

	"github.com/c4dt/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

type suffragiaFormat struct{}

func (suffragiaFormat) Encode(ctx serde.Context, msg serde.Message) ([]byte, error) {
	switch m := msg.(type) {
	case types.Suffragia:
		sJson, err := encodeSuffragia(ctx, m)
		if err != nil {
			return nil, xerrors.Errorf("couldn't encode suffragia: %v", err)
		}

		buff, err := ctx.Marshal(&sJson)
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal form: %v", err)
		}

		return buff, nil
	default:
		return nil, xerrors.Errorf("Unknown format: %T", msg)
	}
}

func (suffragiaFormat) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	var sJson SuffragiaJSON

	err := ctx.Unmarshal(data, &sJson)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal form: %v", err)
	}

	return decodeSuffragia(ctx, sJson)
}

// SuffragiaJSON defines the JSON representation of a suffragia.
type SuffragiaJSON struct {
	UserIDs     []string
	Ciphervotes []json.RawMessage
}

func encodeSuffragia(ctx serde.Context, suffragia types.Suffragia) (SuffragiaJSON, error) {
	ciphervotes := make([]json.RawMessage, len(suffragia.Ciphervotes))

	for i, ciphervote := range suffragia.Ciphervotes {
		buff, err := ciphervote.Serialize(ctx)
		if err != nil {
			return SuffragiaJSON{}, xerrors.Errorf("failed to serialize ciphervote: %v", err)
		}

		ciphervotes[i] = buff
	}
	return SuffragiaJSON{
		UserIDs:     suffragia.UserIDs,
		Ciphervotes: ciphervotes,
	}, nil
}

func decodeSuffragia(ctx serde.Context, suffragiaJSON SuffragiaJSON) (types.Suffragia, error) {
	var res types.Suffragia
	fac := ctx.GetFactory(types.CiphervoteKey{})

	factory, ok := fac.(types.CiphervoteFactory)
	if !ok {
		return res, xerrors.Errorf("invalid ciphervote factory: '%T'", fac)
	}

	ciphervotes := make([]types.Ciphervote, len(suffragiaJSON.Ciphervotes))

	for i, ciphervoteJSON := range suffragiaJSON.Ciphervotes {
		msg, err := factory.Deserialize(ctx, ciphervoteJSON)
		if err != nil {
			return res, xerrors.Errorf("failed to deserialize ciphervote json: %v", err)
		}

		ciphervote, ok := msg.(types.Ciphervote)
		if !ok {
			return res, xerrors.Errorf("wrong type: '%T'", msg)
		}

		ciphervotes[i] = ciphervote
	}

	res = types.Suffragia{
		UserIDs:     suffragiaJSON.UserIDs,
		Ciphervotes: ciphervotes,
	}

	return res, nil
}
