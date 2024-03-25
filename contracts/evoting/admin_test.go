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

func TestAdmin_serde(t *testing.T) {
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
