package evoting

import (
	"testing"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/serde"
	sjson "go.dedis.ch/dela/serde/json"
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
// Serialization/Deserialization of an AdminList should not change its values.
func TestAdmin_Serde(t *testing.T) {
	initialAdminList := []int{111111, 222222, 333333, 123456}

	adminList := types.AdminList{AdminList: initialAdminList}

	value, err := adminList.Serialize(ctxAdminTest)

	require.NoError(t, err)

	// deserialization
	deserializedAdminList := types.AdminList{}

	msgs, err := deserializedAdminList.Deserialize(ctxAdminTest, value)

	require.NoError(t, err)

	updatedAdminList := msgs.(types.AdminList)

	require.Equal(t, initialAdminList, updatedAdminList.AdminList)
}

func TestAdmin_AddAdminAndRemoveAdmin(t *testing.T) {
	initialAdminList := []int{}

	myTestID := "123456"

	adminList := types.AdminList{AdminList: initialAdminList}

	res, err := adminList.GetAdminIndex(myTestID)
	require.Equal(t, -1, res)
	require.NoError(t, err)

	err = adminList.AddAdmin(myTestID)
	require.NoError(t, err)
	res, err = adminList.GetAdminIndex(myTestID)
	require.Equal(t, 0, res)
	require.NoError(t, err)

	err = adminList.RemoveAdmin(myTestID)
	require.ErrorContains(t, err, "cannot remove this Admin because it is the only one remaining")

	err = adminList.AddAdmin("654321")
	require.NoError(t, err)

	err = adminList.RemoveAdmin(myTestID)
	require.NoError(t, err)
	res, err = adminList.GetAdminIndex(myTestID)
	require.Equal(t, -1, res)
	require.NoError(t, err)
}
