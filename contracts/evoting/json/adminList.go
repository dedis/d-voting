package json

import (
	"github.com/c4dt/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

type adminFormFormat struct{}

func (adminFormFormat) Encode(ctx serde.Context, message serde.Message) ([]byte, error) {
	adminForm, ok := message.(types.AdminList)
	if !ok {
		return nil, xerrors.Errorf("Unknown format: %T", message)
	}

	adminFormJSON := AdminFormJSON{
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

	return types.AdminList{
		AdminList: adminFormJSON.AdminList,
	}, nil
}

type AdminFormJSON struct {
	// List of SCIPER with admin rights
	AdminList []int
}
