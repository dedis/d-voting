package integration

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/c4dt/d-voting/contracts/evoting/types"
	ptypes "github.com/c4dt/d-voting/proxy/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/util/encoding"
)

const defaultNodes = 5

type testType int

const (
	SCENARIO testType = iota
	LOAD
)

// Check the shuffled votes versus the cast votes on a few nodes
func TestScenario(t *testing.T) {
	t.Skip("Doesn't work in dedis/d-voting, neither")
	var err error
	numNodes := defaultNodes

	n, ok := os.LookupEnv("NNODES")
	if ok {
		numNodes, err = strconv.Atoi(n)
		require.NoError(t, err)
	}
	t.Run("Basic configuration", getScenarioTest(numNodes, 20, 1))
}

func getScenarioTest(numNodes int, numVotes int, numForm int) func(*testing.T) {
	return func(t *testing.T) {

		proxyList := make([]string, numNodes)

		for i := 0; i < numNodes; i++ {
			proxyList[i] = fmt.Sprintf("http://localhost:%d", 9080+i)
			t.Log(proxyList[i])
		}

		var wg sync.WaitGroup

		for i := 0; i < numForm; i++ {
			t.Log("Starting worker", i)
			wg.Add(1)

			go startFormProcess(&wg, numNodes, numVotes, 0, proxyList, t, numForm, SCENARIO)
			time.Sleep(2 * time.Second)

		}

		t.Log("Waiting for workers to finish")
		wg.Wait()

	}
}

func startFormProcess(wg *sync.WaitGroup, numNodes, numVotes, numSec int, proxyArray []string, t *testing.T, numForm int, testType testType) {
	defer wg.Done()

	const contentType = "application/json"
	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err)
	timeTable := make([]float64, 10)

	step := 0

	// ###################################### CREATE SIMPLE FORM ######
	oldTime := time.Now()

	formID := createFormScenario(contentType, proxyArray[0], secret, t)

	timeTable[step] = time.Since(oldTime).Seconds()
	t.Logf("Creating the election takes: %v sec", timeTable[step])

	step++

	// ##################################### SETUP DKG #########################
	oldTime = time.Now()

	startDKGScenario(numNodes, timeTable, formID, secret, proxyArray, t)

	timeTable[step] = time.Since(oldTime).Seconds()
	t.Logf("DKG setup takes: %v sec", timeTable[step])

	step++

	// ##################################### OPEN FORM #####################
	t.Log("Open form")

	randomproxy := "http://localhost:9081"
	t.Logf("Open form send to proxy %v", randomproxy)

	oldTime = time.Now()

	ok, err := updateForm(secret, randomproxy, formID, "open", t)
	require.NoError(t, err)
	require.True(t, ok)

	timeTable[step] = time.Since(oldTime).Seconds()
	t.Logf("Opening the election takes: %v sec", timeTable[step])

	step++

	// ##################################### GET FORM INFO #################

	proxyAddr1 := proxyArray[0]
	time.Sleep(time.Second * 5)

	getFormResponse := getFormInfo(proxyAddr1, formID, t)
	formpubkey := getFormResponse.Pubkey
	formStatus := getFormResponse.Status
	BallotSize := getFormResponse.BallotSize

	// Get form public key
	pubKey, err := encoding.StringHexToPoint(suite, formpubkey)
	require.NoError(t, err)

	form := types.Form{
		Pubkey:     pubKey,
		Status:     types.Status(formStatus),
		BallotSize: getFormResponse.BallotSize,
	}

	chunksPerBallot := form.ChunksPerBallot()

	t.Logf("Publickey of the form : " + formpubkey)
	t.Logf("Status of the form : %v", formStatus)

	require.NoError(t, err)
	t.Logf("BallotSize of the form : %v", BallotSize)
	t.Logf("chunksPerBallot of the form : %v", chunksPerBallot)

	// ##################################### CAST BALLOTS ######################

	t.Log("cast ballots")

	oldTime = time.Now()

	var votesfrontend []types.Ballot

	switch testType {
	case SCENARIO:
		votesfrontend = castVotesScenario(numVotes, BallotSize, chunksPerBallot, formID, contentType, proxyArray, pubKey, secret, t)
		t.Log("Cast votes scenario")
	case LOAD:
		votesfrontend = castVotesLoad(numVotes/numSec, numSec, BallotSize, chunksPerBallot, formID, contentType, proxyArray, pubKey, secret, t)
		t.Log("Cast votes load")
	}

	timeTable[step] = time.Since(oldTime).Seconds()
	t.Logf("Casting %v ballots takes: %v sec", numVotes, timeTable[step])

	step++

	time.Sleep(time.Second * 5)

	// Kill and restart the node, change false to true when we want to use
	KILLNODE := false
	tmp, ok := os.LookupEnv("KILLNODE")
	if ok {
		KILLNODE, err = strconv.ParseBool(tmp)
		require.NoError(t, err)
	}
	if KILLNODE {
		proxyArray = killNode(proxyArray, 3, t)
		time.Sleep(time.Second * 3)

		t.Log("Restart node")
		restartNode(3, t)
		t.Log("Finished to restart node")
	}

	// ############################# CLOSE FORM FOR REAL ###################

	oldTime = time.Now()

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	t.Logf("Close form (for real) send to proxy %v", randomproxy)

	ok, err = updateForm(secret, randomproxy, formID, "close", t)
	require.NoError(t, err)
	require.True(t, ok)

	timeTable[step] = time.Since(oldTime).Seconds()
	t.Logf("Closing form takes: %v sec", timeTable[step])

	step++

	time.Sleep(time.Second * 3)

	getFormResponse = getFormInfo(proxyAddr1, formID, t)
	formStatus = getFormResponse.Status

	t.Logf("Status of the form : %v", formStatus)
	require.Equal(t, uint16(2), formStatus)

	// ###################################### SHUFFLE BALLOTS ##################
	time.Sleep(time.Second * 5)

	t.Log("shuffle ballots")

	oldTime = time.Now()

	shuffleBallotsRequest := ptypes.UpdateShuffle{
		Action: "shuffle",
	}

	signed, err := createSignedRequest(secret, shuffleBallotsRequest)
	require.NoError(t, err)

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	req, err := http.NewRequest(http.MethodPut, randomproxy+"/evoting/services/shuffle/"+formID, bytes.NewBuffer(signed))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Log("Response body: " + string(body))
	resp.Body.Close()

	timeTable[step] = time.Since(oldTime).Seconds()
	t.Logf("Shuffling ballots takes: %v sec", timeTable[step])

	step++

	getFormResponse = getFormInfo(proxyAddr1, formID, t)
	formStatus = getFormResponse.Status

	err = waitForFormStatus(proxyAddr1, formID, uint16(3), time.Second*100, t)
	require.NoError(t, err)

	t.Logf("Status of the form : %v", formStatus)
	require.Equal(t, uint16(3), formStatus)

	// ###################################### REQUEST PUBLIC SHARES ############
	time.Sleep(time.Second * 5)

	t.Log("request public shares")

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	oldTime = time.Now()

	_, err = updateDKG(secret, randomproxy, formID, "computePubshares", t)
	require.NoError(t, err)

	timeTable[step] = time.Since(oldTime).Seconds()
	t.Logf("Requesting public shares takes: %v sec", timeTable[step])

	step++

	time.Sleep(7 * time.Second)

	getFormResponse = getFormInfo(proxyAddr1, formID, t)
	formStatus = getFormResponse.Status

	err = waitForFormStatus(proxyAddr1, formID, uint16(4), time.Second*300, t)
	require.NoError(t, err)

	t.Logf("Status of the form : %v", formStatus)
	require.Equal(t, uint16(4), formStatus)

	// ###################################### DECRYPT BALLOTS ##################
	time.Sleep(time.Second * 5)

	t.Log("decrypt ballots")

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	oldTime = time.Now()

	ok, err = updateForm(secret, randomproxy, formID, "combineShares", t)
	require.NoError(t, err)
	require.True(t, ok)

	timeTable[step] = time.Since(oldTime).Seconds()
	t.Logf("Decrypting ballots takes: %v sec", timeTable[step])

	step++

	time.Sleep(time.Second * 3)

	getFormResponse = getFormInfo(proxyAddr1, formID, t)
	formStatus = getFormResponse.Status

	err = waitForFormStatus(proxyAddr1, formID, uint16(5), time.Second*100, t)
	require.NoError(t, err)

	t.Logf("Status of the form : %v", formStatus)
	require.Equal(t, uint16(5), formStatus)

	//#################################### VALIDATE FORM RESULT ##############

	tmpBallots := getFormResponse.Result
	var tmpCount bool

	for _, ballotIntem := range tmpBallots {
		tmpComp := ballotIntem
		tmpCount = false
		for _, voteFront := range votesfrontend {
			tmpCount = reflect.DeepEqual(tmpComp, voteFront)
			if tmpCount {
				break
			}
		}
	}

	require.True(t, tmpCount, "front end votes are different from decrypted votes")
	t.Logf("Creating the form takes %v sec", timeTable[0])
	t.Logf("Setting up DKG takes %v sec", timeTable[1])
	t.Logf("Oppening the form takes %v sec", timeTable[2])
	t.Logf("Casting %v ballots takes %v sec", numVotes, timeTable[3])
	t.Logf("Closing the form takes %v sec", timeTable[4])
	t.Logf("Shuffling ballots takes %v sec", timeTable[5])
	t.Logf("Requesting public shares takes %v sec", timeTable[6])
	t.Logf("Decrypting ballots takes %v sec", timeTable[7])

}
