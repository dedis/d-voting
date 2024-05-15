package json

import (
	"github.com/c4dt/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

type adminListFormat struct{}

func (adminListFormat) Encode(ctx serde.Context, message serde.Message) ([]byte, error) {
	adminList, ok := message.(types.AdminList)
	if !ok {
		return nil, xerrors.Errorf("Unknown format: %T", message)
	}

	adminListJSON := AdminListJSON{
		AdminList: adminList.AdminList,
	}

	buff, err := ctx.Marshal(&adminListJSON)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal form: %v", err)
	}

	return buff, nil
}

func (adminListFormat) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	var adminListJSON AdminListJSON

	err := ctx.Unmarshal(data, &adminListJSON)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal form: %v", err)
	}

	return types.AdminList{
		AdminList: adminListJSON.AdminList,
	}, nil
}

type AdminListJSON struct {
	// List of SCIPER with admin rights
	AdminList []int
}
