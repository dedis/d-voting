package json

import (
	"github.com/c4dt/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

type adminFormFormat struct{}

func (adminFormFormat) Encode(ctx serde.Context, message serde.Message) ([]byte, error) {
	switch m := message.(type) {
	case types.AdminForm:
		adminFormJSON := AdminFormJSON{
			FormID:    m.FormID,
			AdminList: m.AdminList,
		}

		buff, err := ctx.Marshal(&adminFormJSON)
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal form: %v", err)
		}

		return buff, nil
	default:
		return nil, xerrors.Errorf("Unknown format: %T", message)
	}
}

func (adminFormFormat) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	var adminFormJSON AdminFormJSON

	err := ctx.Unmarshal(data, &adminFormJSON)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal form: %v", err)
	}

	return types.AdminForm{
		FormID:    adminFormJSON.FormID,
		AdminList: adminFormJSON.AdminList,
	}, nil
}

type AdminFormJSON struct {
	// FormID is the hex-encoded SHA265 of the Tx ID that creates the form
	FormID string

	// List of SCIPER with admin rights
	AdminList []int
}
