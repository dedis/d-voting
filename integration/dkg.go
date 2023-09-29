package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	ptypes "github.com/c4dt/d-voting/proxy/types"
	"github.com/c4dt/d-voting/services/dkg"
	"github.com/c4dt/dela/core/txn"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

// initDkg initializes the DKG for all nodes
func initDkg(nodes []dVotingCosiDela, formID []byte, m txn.Manager) (dkg.Actor, error) {
	var actor dkg.Actor
	var err error

	for _, node := range nodes {
		d := node.(dVotingNode).GetDkg()

		// put Listen in a goroutine to optimize for speed
		actor, err = d.Listen(formID, m)
		if err != nil {
			return nil, xerrors.Errorf("failed to GetDkg: %v", err)
		}
	}

	_, err = actor.Setup()
	if err != nil {
		return nil, xerrors.Errorf("failed to Setup: %v", err)
	}

	return actor, nil
}

// startDKGScenario starts the DKG for all nodes in the scenario tests
func startDKGScenario(numNodes int, timeTable []float64, formID string, secret kyber.Scalar, proxyArray []string, t *testing.T) {
	t.Log("Init DKG")

	for i := 0; i < numNodes; i++ {
		t.Log("Node" + strconv.Itoa(i+1))
		t.Log(proxyArray[i])
		err := initDKG(secret, proxyArray[i], formID, t)
		require.NoError(t, err)

	}
	t.Log("Setup DKG")

	msg := ptypes.UpdateDKG{
		Action: "setup",
	}
	signed, err := createSignedRequest(secret, msg)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, proxyArray[0]+"/evoting/services/dkg/actors/"+formID, bytes.NewBuffer(signed))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	// Wait for DKG setup
	err = waitForDKG(proxyArray[0], formID, time.Second*100, t)
	require.NoError(t, err)

}

// initDkg initializes the DKG for all nodes for Scenario tests
func initDKG(secret kyber.Scalar, proxyAddr, formIDHex string, t *testing.T) error {
	setupDKG := ptypes.NewDKGRequest{
		FormID: formIDHex,
	}

	signed, err := createSignedRequest(secret, setupDKG)
	require.NoError(t, err)

	resp, err := http.Post(proxyAddr+"/evoting/services/dkg/actors", "application/json", bytes.NewBuffer(signed))
	if err != nil {
		return xerrors.Errorf("failed to post request: %v", err)
	}
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	return nil
}

// updateDKG updates the DKG step for Scenario tests
func updateDKG(secret kyber.Scalar, proxyAddr, formIDHex, action string, t *testing.T) (int, error) {
	msg := ptypes.UpdateDKG{
		Action: action,
	}

	signed, err := createSignedRequest(secret, msg)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, proxyAddr+"/evoting/services/dkg/actors/"+formIDHex, bytes.NewBuffer(signed))
	if err != nil {
		return 0, xerrors.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, xerrors.Errorf("failed to execute the query: %v", err)
	}

	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	return 0, nil
}

// getDKGInfo gets the DKG info for the scenario tests
func getDKGInfo(proxyAddr, formID string, t *testing.T) ptypes.GetActorInfo {

	resp, err := http.Get(proxyAddr + "/evoting/services/dkg/actors" + "/" + formID)
	require.NoError(t, err)

	var infoDKG ptypes.GetActorInfo
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&infoDKG)
	require.NoError(t, err)

	resp.Body.Close()

	return infoDKG

}

// waitForDKG waits for the DKG to be setup for the scenario tests
func waitForDKG(proxyAddr, formID string, timeOut time.Duration, t *testing.T) error {
	expired := time.Now().Add(timeOut)

	isOK := func() bool {
		infoDKG := getDKGInfo(proxyAddr, formID, t)
		return infoDKG.Status == 1
	}

	for !isOK() {

		if time.Now().After(expired) {
			return xerrors.New("expired")
		}

		time.Sleep(time.Millisecond * 1000)
	}

	return nil
}
