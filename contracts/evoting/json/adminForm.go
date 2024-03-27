package json

import (
	"github.com/c4dt/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

type adminFormFormat struct{}

func (adminFormFormat) Encode(ctx serde.Context, message serde.Message) ([]byte, error) {
	adminForm, ok := message.(types.AdminForm)
	if !ok {
		return nil, xerrors.Errorf("Unknown format: %T", message)
	}

	adminFormJSON := AdminFormJSON{
		FormID:    adminForm.FormID,
		AdminList: adminForm.AdminList,
	}

	buff, err := ctx.Marshal(&adminFormJSON)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal form: %v", err)
	}

	return buff, nil
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
