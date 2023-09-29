package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/c4dt/d-voting/contracts/evoting"
	"github.com/c4dt/d-voting/contracts/evoting/types"
	"github.com/c4dt/d-voting/internal/testing/fake"
	"github.com/c4dt/d-voting/proxy/txnmanager"
	ptypes "github.com/c4dt/d-voting/proxy/types"
	"github.com/c4dt/dela/core/execution/native"
	"github.com/c4dt/dela/core/ordering"
	"github.com/c4dt/dela/core/txn"
	"github.com/c4dt/dela/serde"
	jsonDela "github.com/c4dt/dela/serde/json"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"golang.org/x/xerrors"
)

var serdecontext = jsonDela.NewContext()

func encodeID(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
}

// for integration tests
func createForm(m txManager, title string, admin string) ([]byte, error) {
	// Define the configuration :
	configuration := fake.BasicConfiguration

	createForm := types.CreateForm{
		Configuration: configuration,
		AdminID:       admin,
	}

	data, err := createForm.Serialize(serdecontext)
	if err != nil {
		return nil, xerrors.Errorf("failed to serialize: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.FormArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCreateForm)},
	}

	txID, err := m.addAndWait(args...)
	if err != nil {
		return nil, xerrors.Errorf(addAndWaitErr, err)
	}

	// Calculate formID from
	hash := sha256.New()
	hash.Write(txID)
	formID := hash.Sum(nil)

	return formID, nil
}

// for scenario/load test
func createFormScenario(contentType, proxy string, secret kyber.Scalar, t *testing.T) string {
	t.Log("Create form")

	configuration := fake.BasicConfiguration

	createSimpleFormRequest := ptypes.CreateFormRequest{
		Configuration: configuration,
		AdminID:       "adminId",
	}

	signed, err := createSignedRequest(secret, createSimpleFormRequest)
	require.NoError(t, err)

	resp, err := http.Post(proxy+"/evoting/forms", contentType, bytes.NewBuffer(signed))
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", body)

	t.Log("response body:", string(body))
	resp.Body.Close()

	var createFormResponse ptypes.CreateFormResponse

	err = json.Unmarshal(body, &createFormResponse)
	require.NoError(t, err)

	t.Log("response token:", createFormResponse.Token)
	formID := createFormResponse.FormID

	// wait for the election to be created
	ok, err := pollTxnInclusion(60, time.Second, proxy, createFormResponse.Token, t)
	require.NoError(t, err)
	require.True(t, ok)

	t.Logf("ID of the form : " + formID)

	return formID
}

// for integration tests
func openForm(m txManager, formID []byte) error {
	openForm := &types.OpenForm{
		FormID: hex.EncodeToString(formID),
	}

	data, err := openForm.Serialize(serdecontext)
	if err != nil {
		return xerrors.Errorf("failed to serialize open form: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.FormArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdOpenForm)},
	}

	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf(addAndWaitErr, err)
	}

	return nil
}

func getForm(formFac serde.Factory, formID []byte,
	service ordering.Service) (types.Form, error) {

	form := types.Form{}

	proof, err := service.GetProof(formID)
	if err != nil {
		return form, xerrors.Errorf("failed to GetProof: %v", err)
	}

	if proof == nil {
		return form, xerrors.Errorf("form does not exist: %v", err)
	}

	message, err := formFac.Deserialize(serdecontext, proof.GetValue())
	if err != nil {
		return form, xerrors.Errorf("failed to deserialize Form: %v", err)
	}

	form, ok := message.(types.Form)
	if !ok {
		return form, xerrors.Errorf("wrong message type: %T", message)
	}

	return form, nil
}

// for integration tests
func closeForm(m txManager, formID []byte, admin string) error {
	closeForm := &types.CloseForm{
		FormID: hex.EncodeToString(formID),
		UserID: admin,
	}

	data, err := closeForm.Serialize(serdecontext)
	if err != nil {
		return xerrors.Errorf("failed to serialize open form: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.FormArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCloseForm)},
	}

	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf("failed to Marshall closeForm: %v", err)
	}

	return nil
}

// for Scenario
func waitForFormStatus(proxyAddr, formID string, status uint16, timeOut time.Duration, t *testing.T) error {
	expired := time.Now().Add(timeOut)

	isOK := func() bool {
		infoForm := getFormInfo(proxyAddr, formID, t)
		return infoForm.Status == status
	}

	for !isOK() {
		if time.Now().After(expired) {
			return xerrors.New("expired")
		}

		time.Sleep(time.Millisecond * 1000)
	}

	return nil
}

// updateForm updates the form with the given action for the scenario tests
func updateForm(secret kyber.Scalar, proxyAddr, formIDHex, action string, t *testing.T) (bool, error) {
	msg := ptypes.UpdateFormRequest{
		Action: action,
	}

	signed, err := createSignedRequest(secret, msg)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, proxyAddr+"/evoting/forms/"+formIDHex, bytes.NewBuffer(signed))
	if err != nil {
		return false, xerrors.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, xerrors.Errorf("failed to read response body: %v", err)
	}
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", body)

	// use the pollTxnInclusion func
	var result txnmanager.TransactionClientInfo
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, xerrors.Errorf("failed to unmarshal response body: %v", err)
	}

	// wait until the update is completed
	return pollTxnInclusion(60, time.Second, proxyAddr, result.Token, t)

}

// for Scenario
func getFormInfo(proxyAddr, formID string, t *testing.T) ptypes.GetFormResponse {
	// t.Log("Get form info")

	resp, err := http.Get(proxyAddr + "/evoting/forms" + "/" + formID)
	require.NoError(t, err)

	var infoForm ptypes.GetFormResponse
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&infoForm)
	require.NoError(t, err)

	resp.Body.Close()

	return infoForm

}

func createSignedRequest(secret kyber.Scalar, msg interface{}) ([]byte, error) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal json: %v", err)
	}

	payload := base64.URLEncoding.EncodeToString(jsonMsg)

	hash := sha256.New()

	hash.Write([]byte(payload))
	md := hash.Sum(nil)

	signature, err := schnorr.Sign(suite, secret, md)
	if err != nil {
		return nil, xerrors.Errorf("failed to sign: %v", err)
	}

	signed := ptypes.SignedRequest{
		Payload:   payload,
		Signature: hex.EncodeToString(signature),
	}

	signedJSON, err := json.Marshal(signed)
	if err != nil {
		return nil, xerrors.Errorf("failed to create json signed: %v", err)
	}

	return signedJSON, nil
}
