package types

import (
	"crypto/sha256"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/dela/serde/registry"
	"golang.org/x/xerrors"
)

var adminFormFormat = registry.NewSimpleRegistry()

func RegisterAdminFormFormat(format serde.Format, engine serde.FormatEngine) {
	adminFormFormat.Register(format, engine)
}

type AdminList struct {
	// List of SCIPER with admin rights
	AdminList []int
}

func (adminList AdminList) Serialize(ctx serde.Context) ([]byte, error) {
	format := adminFormFormat.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, adminList)
	if err != nil {
		return nil, xerrors.Errorf("Failed to encode AdminList: %v", err)
	}

	return data, nil
}

func (adminList AdminList) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := adminFormFormat.Get(ctx.GetFormat())

	message, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("Failed to decode: %v", err)
	}

	return message, nil
}

// AddAdmin add a new admin to the system.
func (adminList *AdminList) AddAdmin(userID string) error {
	sciperInt, err := SciperToInt(userID)
	if err != nil {
		return xerrors.Errorf("Failed SciperToInt: %v", err)
	}

	adminList.AdminList = append(adminList.AdminList, sciperInt)

	return nil
}

// GetAdminIndex return the index of admin if userID is one, else return -1
func (adminList *AdminList) GetAdminIndex(userID string) (int, error) {
	sciperInt, err := SciperToInt(userID)
	if err != nil {
		return -1, xerrors.Errorf("Failed SciperToInt: %v", err)
	}

	for i := 0; i < len(adminList.AdminList); i++ {
		if adminList.AdminList[i] == sciperInt {
			return i, nil
		}
	}

	return -1, nil
}

// RemoveAdmin add a new admin to the system.
func (adminList *AdminList) RemoveAdmin(userID string) error {
	index, err := adminList.GetAdminIndex(userID)
	if err != nil {
		return xerrors.Errorf("Failed GetAdminIndex: %v", err)
	}

	if index < 0 {
		return xerrors.Errorf("Error while retrieving the index of the element.")
	}

	// We don't want to have a form without any Owners.
	if len(adminList.AdminList) <= 1 {
		return xerrors.Errorf("Error, cannot remove this Admin because it is the " +
			"only one remaining.")
	}

	adminList.AdminList = append(adminList.AdminList[:index], adminList.AdminList[index+1:]...)
	return nil
}

func AdminFormFromStore(ctx serde.Context, adminFormFac serde.Factory, store store.Readable, adminListId string) (AdminList, error) {
	adminForm := AdminList{}

	h := sha256.New()
	h.Write([]byte(adminListId))
	adminFormIDBuf := h.Sum(nil)

	adminFormBuf, err := store.Get(adminFormIDBuf)
	if err != nil {
		return adminForm, xerrors.Errorf("While getting data for form: %v", err)
	}
	if len(adminFormBuf) == 0 {
		return adminForm, xerrors.Errorf("No form found")
	}

	message, err := adminFormFac.Deserialize(ctx, adminFormBuf)
	if err != nil {
		return adminForm, xerrors.Errorf("failed to deserialize AdminList: %v", err)
	}

	adminForm, ok := message.(AdminList)
	if !ok {
		return adminForm, xerrors.Errorf("Wrong message type: %T", message)
	}

	return adminForm, nil
}

// AdminFormFactory provides the mean to deserialize a AdminList. It naturally
// uses the formFormat.
//
// - implements serde.Factory
type AdminFormFactory struct{}

// Deserialize implements serde.Factory
func (AdminFormFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := adminFormFormat.Get(ctx.GetFormat())

	message, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode: %v", err)
	}

	return message, nil
}
