package integration

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	ptypes "github.com/dedis/d-voting/proxy/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/encoding"
)

// Check the shuffled votes versus the cast votes on a few nodes
func TestLoad(t *testing.T) {
	var err error
	numNodes := 10

	n, ok := os.LookupEnv("NNODES")
	if ok {
		numNodes, err = strconv.Atoi(n)
		require.NoError(t, err)
	}
	t.Run("Basic configuration", getLoadTest(numNodes, 10, 60, 1))
}

func getLoadTest(numNodes, numVotesPerSec, numSec, numForm int) func(*testing.T) {
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

			go startFormProcessLoad(&wg, numNodes, numVotesPerSec, numSec, proxyList, t, numForm)
			time.Sleep(2 * time.Second)

		}

		t.Log("Waiting for workers to finish")
		wg.Wait()

	}
}

func startFormProcessLoad(wg *sync.WaitGroup, numNodes, numVotesPerSec, numSec int, proxyArray []string, t *testing.T, numForm int) {
	defer wg.Done()
	rand.Seed(0)

	const contentType = "application/json"
	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err)

	// ###################################### CREATE SIMPLE FORM ######

	t.Log("Create form")

	configuration := fake.BasicConfiguration

	createSimpleFormRequest := ptypes.CreateFormRequest{
		Configuration: configuration,
		AdminID:       "adminId",
	}

	signed, err := createSignedRequest(secret, createSimpleFormRequest)
	require.NoError(t, err)

	resp, err := http.Post(proxyArray[0]+"/evoting/forms", contentType, bytes.NewBuffer(signed))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s  %s", resp.Status,body)

	

	t.Log("response body:", string(body))
	resp.Body.Close()

	var createFormResponse ptypes.CreateFormResponse

	err = json.Unmarshal(body, &createFormResponse)
	require.NoError(t, err)

	formID := createFormResponse.FormID

	t.Logf("ID of the form : " + formID)

	// ##################################### SETUP DKG #########################

	t.Log("Init DKG")

	for i := 0; i < numNodes; i++ {
		t.Log("Node" + strconv.Itoa(i+1))
		t.Log(proxyArray[i])
		err = initDKG(secret, proxyArray[i], formID, t)
		require.NoError(t, err)

	}
	t.Log("Setup DKG")

	msg := ptypes.UpdateDKG{
		Action: "setup",
	}
	signed, err = createSignedRequest(secret, msg)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, proxyArray[0]+"/evoting/services/dkg/actors/"+formID, bytes.NewBuffer(signed))
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	// ##################################### OPEN FORM #####################
	// Wait for DKG setup
	timeTable := make([]float64, 5)
	oldTime := time.Now()

	err = waitForDKG(proxyArray[0], formID, time.Second*100, t)
	require.NoError(t, err)

	currentTime := time.Now()
	diff := currentTime.Sub(oldTime)
	timeTable[0] = diff.Seconds()
	t.Logf("DKG setup takes: %v sec", diff.Seconds())

	randomproxy := "http://localhost:9081"
	t.Logf("Open form send to proxy %v", randomproxy)

	_, err = updateForm(secret, randomproxy, formID, "open", t)
	require.NoError(t, err)
	// ##################################### GET FORM INFO #################

	proxyAddr1 := proxyArray[0]
	time.Sleep(time.Second * 5)

	getFormResponse := getFormInfo(proxyAddr1, formID, t)
	formpubkey := getFormResponse.Pubkey
	formStatus := getFormResponse.Status
	BallotSize := getFormResponse.BallotSize
	Chunksperballot := chunksPerBallot(BallotSize)

	t.Logf("Publickey of the form : " + formpubkey)
	t.Logf("Status of the form : %v", formStatus)

	require.NoError(t, err)
	t.Logf("BallotSize of the form : %v", BallotSize)
	t.Logf("Chunksperballot of the form : %v", Chunksperballot)

	// Get form public key
	pubKey, err := encoding.StringHexToPoint(suite, formpubkey)
	require.NoError(t, err)

	// ##################################### CAST BALLOTS ######################

	votesfrontend := castVotesLoad(numVotesPerSec, numSec, BallotSize, Chunksperballot, formID, contentType, proxyArray, pubKey, secret, t)

	// ############################# CLOSE FORM FOR REAL ###################
	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	t.Logf("Close form (for real) send to proxy %v", randomproxy)

	_, err = updateForm(secret, randomproxy, formID, "close", t)
	require.NoError(t, err)

	time.Sleep(time.Second * 3)

	getFormResponse = getFormInfo(proxyAddr1, formID, t)
	formStatus = getFormResponse.Status

	t.Logf("Status of the form : %v", formStatus)
	require.Equal(t, uint16(2), formStatus)

	// ###################################### SHUFFLE BALLOTS ##################
	time.Sleep(time.Second * 5)

	t.Log("shuffle ballots")

	shuffleBallotsRequest := ptypes.UpdateShuffle{
		Action: "shuffle",
	}

	signed, err = createSignedRequest(secret, shuffleBallotsRequest)
	require.NoError(t, err)

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	req, err = http.NewRequest(http.MethodPut, randomproxy+"/evoting/services/shuffle/"+formID, bytes.NewBuffer(signed))
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)

	currentTime = time.Now()
	diff = currentTime.Sub(oldTime)
	timeTable[1] = diff.Seconds()
	t.Logf("Shuffle takes: %v sec", diff.Seconds())

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Log("Response body: " + string(body))
	resp.Body.Close()

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

	currentTime = time.Now()
	diff = currentTime.Sub(oldTime)
	timeTable[2] = diff.Seconds()

	t.Logf("Request public share takes: %v sec", diff.Seconds())

	time.Sleep(10 * time.Second)

	getFormResponse = getFormInfo(proxyAddr1, formID, t)
	formStatus = getFormResponse.Status

	oldTime = time.Now()
	err = waitForFormStatus(proxyAddr1, formID, uint16(4), time.Second*300, t)
	require.NoError(t, err)

	currentTime = time.Now()
	diff = currentTime.Sub(oldTime)
	timeTable[4] = diff.Seconds()
	t.Logf("Status goes to 4 takes: %v sec", diff.Seconds())

	t.Logf("Status of the form : %v", formStatus)
	require.Equal(t, uint16(4), formStatus)

	// ###################################### DECRYPT BALLOTS ##################
	time.Sleep(time.Second * 5)

	t.Log("decrypt ballots")

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]
	oldTime = time.Now()

	_, err = updateForm(secret, randomproxy, formID, "combineShares", t)
	require.NoError(t, err)

	currentTime = time.Now()
	diff = currentTime.Sub(oldTime)
	timeTable[3] = diff.Seconds()

	t.Logf("decryption takes: %v sec", diff.Seconds())

	time.Sleep(time.Second * 7)

	getFormResponse = getFormInfo(proxyAddr1, formID, t)
	formStatus = getFormResponse.Status

	err = waitForFormStatus(proxyAddr1, formID, uint16(5), time.Second*100, t)
	require.NoError(t, err)

	t.Logf("Status of the form : %v", formStatus)
	require.Equal(t, uint16(5), formStatus)

	//#################################### VALIDATE FORM RESULT ##############

	tmpBallots := getFormResponse.Result
	var tmpCount bool
	require.Equal(t, len(tmpBallots), numVotesPerSec*numSec)

	for _, ballotIntem := range tmpBallots {
		tmpComp := ballotIntem
		tmpCount = false
		for _, voteFront := range votesfrontend {
			// t.Logf("voteFront: %v", voteFront)
			// t.Logf("tmpComp: %v", tmpComp)

			tmpCount = reflect.DeepEqual(tmpComp, voteFront)
			// t.Logf("tmpCount: %v", tmpCount)

			if tmpCount {
				break
			}
		}
	}

	require.True(t, tmpCount, "front end votes are different from decrypted votes")
	t.Logf("DKG setup time : %v", timeTable[0])
	t.Logf("shuffle time : %v", timeTable[1])
	t.Logf("Public share time : %v", timeTable[2])
	t.Logf("Status goes to 4 takes: %v sec", diff.Seconds())
	t.Logf("decryption time : %v", timeTable[3])
}

func castVotesLoad(numVotesPerSec, numSec, BallotSize, chunksPerBallot int, formID, contentType string, proxyArray []string, pubKey kyber.Point, secret kyber.Scalar, t *testing.T) []types.Ballot {
	t.Log("cast ballots")

	//make List of ballots
	b1 := string("select:" + encodeIDBallot("bb") + ":0,0,1,0\n" + "text:" + encodeIDBallot("ee") + ":eWVz\n\n") //encoding of "yes"

	numVotes := numVotesPerSec * numSec

	ballotList := make([]string, numVotes)
	for i := 1; i <= numVotes; i++ {
		ballotList[i-1] = b1
	}

	votesfrontend := make([]types.Ballot, numVotes)

	fakeConfiguration := fake.BasicConfiguration

	for i := 0; i < numVotes; i++ {

		var bMarshal types.Ballot
		form := types.Form{
			Configuration: fakeConfiguration,
			FormID:        formID,
			BallotSize:    BallotSize,
		}

		err := bMarshal.Unmarshal(ballotList[i], form)
		require.NoError(t, err)

		votesfrontend[i] = bMarshal
	}
	proxyCount := len(proxyArray)

	// all ballots are identical
	ballot, err := marshallBallotManual(b1, pubKey, chunksPerBallot)
	require.NoError(t, err)

	for i := 0; i < numSec; i++ {
		// send the votes asynchrounously and wait for the response
		
		for j := 0; j < numVotesPerSec; j++ {
			randomproxy := proxyArray[rand.Intn(proxyCount)]
			castVoteRequest := ptypes.CastVoteRequest{
				UserID: "user"+strconv.Itoa(i*numVotesPerSec+j),
				Ballot: ballot,
			}
			go cast(false,castVoteRequest, contentType, randomproxy, formID, secret, t)

		}
		t.Logf("casted votes %d", (i+1)*numVotesPerSec)
		time.Sleep(time.Second)


	}

	time.Sleep(time.Second * 30)

	return votesfrontend
}


func cast(isRetry bool, castVoteRequest ptypes.CastVoteRequest, contentType, randomproxy, formID string, secret kyber.Scalar, t *testing.T)  {
	

	
	t.Logf("cast ballot to proxy %v", randomproxy)

	// t.Logf("vote is: %v", castVoteRequest)
	signed, err := createSignedRequest(secret, castVoteRequest)
	require.NoError(t, err)

	resp, err := http.Post(randomproxy+"/evoting/forms/"+formID+"/vote", contentType, bytes.NewBuffer(signed))
	require.NoError(t, err)
	
	if http.StatusOK != resp.StatusCode && !isRetry {
		t.Logf("unexpected status: %s retry", resp.Status)
		cast(true,castVoteRequest, contentType, randomproxy, formID, secret, t)
		return
	}

	responseBody,err:=ioutil.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s %s", resp.Status,responseBody)

	_, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	resp.Body.Close()
}
