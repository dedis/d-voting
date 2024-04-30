package evoting

import (
	"github.com/c4dt/d-voting/contracts/evoting/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/serde"
	sjson "go.dedis.ch/dela/serde/json"
	"testing"
)

var ctxAdminTest serde.Context

var formFacAdminTest serde.Factory
var transactionFacAdminTest serde.Factory

func init() {
	ciphervoteFac := types.CiphervoteFactory{}
	formFacAdminTest = types.NewFormFactory(ciphervoteFac, fakeAuthorityFactory{})
	transactionFacAdminTest = types.NewTransactionFactory(ciphervoteFac)

	ctxAdminTest = sjson.NewContext()
}

// This test create an Admin Form structure which is then serialized and
// deserialized to check whether these operations work as intended.
// Serialization/Deserialization of an AdminForm should not change its values.
func TestAdmin_Serde(t *testing.T) {
	adminFormID := "myID"
	adminFormList := []int{111111, 222222, 333333, 123456}

	adminForm := types.AdminForm{FormID: adminFormID, AdminList: adminFormList}

	value, err := adminForm.Serialize(ctxAdminTest)

	require.NoError(t, err)

	// deserialization
	newAdminForm := types.AdminForm{}

	msgs, err := newAdminForm.Deserialize(ctxAdminTest, value)

	require.NoError(t, err)

	updatedAdminForm := msgs.(types.AdminForm)

	require.Equal(t, adminFormID, updatedAdminForm.FormID)
	require.Equal(t, adminFormList, updatedAdminForm.AdminList)
}

func TestAdmin_AddAdminAndRemoveAdmin(t *testing.T) {
	adminFormID := "myID"
	adminFormList := []int{}

	myTestID := "123456"

	adminForm := types.AdminForm{FormID: adminFormID, AdminList: adminFormList}

	res, err := adminForm.GetAdminIndex(myTestID)
	require.Equal(t, -1, res)
	require.NoError(t, err)

	err = adminForm.AddAdmin(myTestID)
	require.NoError(t, err)
	res, err = adminForm.GetAdminIndex(myTestID)
	require.Equal(t, 0, res)
	require.NoError(t, err)

	err = adminForm.RemoveAdmin(myTestID)
	require.NoError(t, err)
	res, err = adminForm.GetAdminIndex(myTestID)
	require.Equal(t, -1, res)
	require.NoError(t, err)
}
