package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	ptypes "github.com/dedis/d-voting/proxy/types"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/serde"
	jsonDela "go.dedis.ch/dela/serde/json"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

var serdecontext = jsonDela.NewContext()

func encodeID(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
}

func ballotIsNull(ballot types.Ballot) bool {
	return ballot.SelectResultIDs == nil && ballot.SelectResult == nil &&
		ballot.RankResultIDs == nil && ballot.RankResult == nil &&
		ballot.TextResultIDs == nil && ballot.TextResult == nil
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

	formID := createFormResponse.FormID

	ok, err := pollTxnInclusion(proxy, createFormResponse.Token, t)
	require.NoError(t, err)
	require.True(t, ok)

	t.Logf("ID of the form : " + formID)

	return formID
}

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

func castVotesRandomly(m txManager, actor dkg.Actor, form types.Form,
	numberOfVotes int) ([]types.Ballot, error) {

	possibleBallots := []string{
		string("select:" + encodeID("bb") + ":0,0,1,0\n" +
			"text:" + encodeID("ee") + ":eWVz\n\n"), //encoding of "yes"
		string("select:" + encodeID("bb") + ":1,1,0,0\n" +
			"text:" + encodeID("ee") + ":amE=\n\n"), //encoding of "ja
		string("select:" + encodeID("bb") + ":0,0,0,1\n" +
			"text:" + encodeID("ee") + ":b3Vp\n\n"), //encoding of "oui"
	}

	votes := make([]types.Ballot, numberOfVotes)

	for i := 0; i < numberOfVotes; i++ {
		randomIndex := rand.Intn(len(possibleBallots))
		vote := possibleBallots[randomIndex]

		ciphervote, err := marshallBallot(strings.NewReader(vote), actor, form.ChunksPerBallot())
		if err != nil {
			return nil, xerrors.Errorf("failed to marshallBallot: %v", err)
		}

		userID := "user " + strconv.Itoa(i)

		castVote := types.CastVote{
			FormID: form.FormID,
			UserID: userID,
			Ballot: ciphervote,
		}

		data, err := castVote.Serialize(serdecontext)
		if err != nil {
			return nil, xerrors.Errorf("failed to serialize cast vote: %v", err)
		}

		args := []txn.Arg{
			{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
			{Key: evoting.FormArg, Value: data},
			{Key: evoting.CmdArg, Value: []byte(evoting.CmdCastVote)},
		}

		_, err = m.addAndWait(args...)
		if err != nil {
			return nil, xerrors.Errorf(addAndWaitErr, err)
		}

		var ballot types.Ballot
		err = ballot.Unmarshal(vote, form)
		if err != nil {
			return nil, xerrors.Errorf("failed to unmarshal ballot: %v", err)
		}

		votes[i] = ballot
	}

	return votes, nil
}

func castBadVote(m txManager, actor dkg.Actor, form types.Form, numberOfBadVotes int) error {

	possibleBallots := []string{
		string("select:" + encodeID("bb") + ":1,0,1,1\n" +
			"text:" + encodeID("ee") + ":bm9ub25vbm8=\n\n"), //encoding of "nononono"
		string("select:" + encodeID("bb") + ":1,1,1,1\n" +
			"text:" + encodeID("ee") + ":bm8=\n\n"), //encoding of "no"

	}

	for i := 0; i < numberOfBadVotes; i++ {
		randomIndex := rand.Intn(len(possibleBallots))
		vote := possibleBallots[randomIndex]

		ciphervote, err := marshallBallot(strings.NewReader(vote), actor, form.ChunksPerBallot())
		if err != nil {
			return xerrors.Errorf("failed to marshallBallot: %v", err)
		}

		userID := "badUser " + strconv.Itoa(i)

		castVote := types.CastVote{
			FormID: form.FormID,
			UserID: userID,
			Ballot: ciphervote,
		}

		data, err := castVote.Serialize(serdecontext)
		if err != nil {
			return xerrors.Errorf("failed to serialize cast vote: %v", err)
		}

		args := []txn.Arg{
			{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
			{Key: evoting.FormArg, Value: data},
			{Key: evoting.CmdArg, Value: []byte(evoting.CmdCastVote)},
		}

		_, err = m.addAndWait(args...)
		if err != nil {
			return xerrors.Errorf(addAndWaitErr, err)
		}

		//votes[i] = ballot
	}

	return nil
}

func marshallBallot(vote io.Reader, actor dkg.Actor, chunks int) (types.Ciphervote, error) {

	var ballot = make([]types.EGPair, chunks)

	buf := make([]byte, 29)

	for i := 0; i < chunks; i++ {
		var K, C kyber.Point
		var err error

		n, err := vote.Read(buf)
		if err != nil {
			return nil, xerrors.Errorf("failed to read: %v", err)
		}

		K, C, _, err = actor.Encrypt(buf[:n])
		if err != nil {
			return types.Ciphervote{}, xerrors.Errorf("failed to encrypt the plaintext: %v", err)
		}

		ballot[i] = types.EGPair{
			K: K,
			C: C,
		}
	}

	return ballot, nil
}

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

func decryptBallots(m txManager, actor dkg.Actor, form types.Form) error {
	if form.Status != types.PubSharesSubmitted {
		return xerrors.Errorf("cannot decrypt: not all pubShares submitted")
	}

	decryptBallots := types.CombineShares{
		FormID: form.FormID,
	}

	data, err := decryptBallots.Serialize(serdecontext)
	if err != nil {
		return xerrors.Errorf("failed to serialize ballots: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.FormArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCombineShares)},
	}

	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf(addAndWaitErr, err)
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

// for Scenario
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

	//use the pollTxnInclusion func
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, xerrors.Errorf("failed to unmarshal response body: %v", err)
	}

	return pollTxnInclusion(proxyAddr, result["Token"].(string), t)

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
