package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"sync"
	"time"

	//"math/rand"
	"net/http"
	//"reflect"
	"strconv"
	"strings"

	//"sync"
	"testing"
	//"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	ptypes "github.com/dedis/d-voting/proxy/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/xerrors"
)

var suite = suites.MustFind("Ed25519")

// Check the shuffled votes versus the cast votes on a few nodes
func TestScenario(t *testing.T) {
	t.Run("Basic configuration", getScenarioTest())
	t.Run("Differents combination ", testVariableNode(3, 3, 1))
}

func getScenarioTest() func(*testing.T) {
	return func(t *testing.T) {

		const contentType = "application/json"
		secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
		require.NoError(t, err, "failed to decode key: %v", err)

		secret := suite.Scalar()
		err = secret.UnmarshalBinary(secretkeyBuf)
		require.NoError(t, err, "failed to Unmarshal key: %v", err)

		t.Parallel()

		proxyAddr1 := "http://localhost:9080"
		proxyAddr2 := "http://localhost:9081"
		proxyAddr3 := "http://localhost:9082"

		proxyArray := [3]string{proxyAddr1, proxyAddr2, proxyAddr3}

		// ###################################### CREATE SIMPLE ELECTION ######

		t.Logf("Create election")

		configuration := fake.BasicConfiguration

		createSimpleElectionRequest := ptypes.CreateElectionRequest{
			Configuration: configuration,
			AdminID:       "adminId",
		}
		fmt.Println(secret)
		signed, err := createSignedRequest(secret, createSimpleElectionRequest)
		require.NoError(t, err, "fail to create singature")
		resp, err := http.Post(proxyAddr1+"/evoting/elections", contentType, bytes.NewBuffer(signed))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()

		var objmap map[string]interface{}

		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parse the body of the response from js: %v", err)
		electionID := objmap["ElectionID"].(string)
		t.Logf("ID of the election : " + electionID)

		// ##################################### SETUP DKG #########################

		t.Log("Init DKG")

		t.Log("Node 1")

		err = initDKG(secret, proxyAddr1, electionID, t)
		require.NoError(t, err, "failed to init dkg 1: %v", err)

		t.Log("Node 2")
		err = initDKG(secret, proxyAddr2, electionID, t)
		require.NoError(t, err, "failed to init dkg 2: %v", err)

		t.Log("Node 3")
		err = initDKG(secret, proxyAddr3, electionID, t)
		require.NoError(t, err, "failed to init dkg 3: %v", err)

		t.Log("Setup DKG")

		msg := ptypes.UpdateDKG{
			Action: "setup",
		}
		signed, err = createSignedRequest(secret, msg)
		require.NoError(t, err, "failed to sign: %v", err)
		req, err := http.NewRequest(http.MethodPut, proxyAddr1+"/evoting/services/dkg/actors/"+electionID, bytes.NewBuffer(signed))
		require.NoError(t, err, "failed to create request: %v", err)
		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err, "failed to setup dkg on node 1: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		// ##################################### OPEN ELECTION #####################

		randomproxy := proxyArray[rand.Intn(len(proxyArray))]
		t.Logf("Open election send to proxy %v", randomproxy)

		_, err = updateElection(secret, proxyAddr1, electionID, "open", t)
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		// ##################################### GET ELECTION INFO #################

		getElectionInfo(objmap, proxyAddr1, electionID, t)

		electionpubkey := objmap["Pubkey"].(string)
		electionStatus := int(objmap["Status"].(float64))
		BallotSize := int(objmap["BallotSize"].(float64))
		Chunksperballot := ChunksPerBallotManuel(BallotSize)
		t.Logf("Publickey of the election : " + electionpubkey)
		t.Logf("Status of the election : %v", electionStatus)

		require.NoError(t, err, "failed to unmarshal pubkey: %v", err)
		t.Logf("BallotSize of the election : %v", BallotSize)
		t.Logf("Chunksperballot of the election : %v", Chunksperballot)

		// Get election public key

		pubkeyBuf, err := hex.DecodeString(electionpubkey)
		require.NoError(t, err, "failed to decode key: %v", err)

		pubKey := suite.Point()
		err = pubKey.UnmarshalBinary(pubkeyBuf)
		require.NoError(t, err, "failed to Unmarshal key: %v", err)

		// ##################################### CAST BALLOTS ######################

		t.Log("cast ballots")

		// Create the ballots
		b1 := string("select:" + encodeIDBallot("bb") + ":0,0,1,0\n" +
			"text:" + encodeIDBallot("ee") + ":eWVz\n\n") //encoding of "yes"

		b2 := string("select:" + encodeIDBallot("bb") + ":1,1,0,0\n" +
			"text:" + encodeIDBallot("ee") + ":amE=\n\n") //encoding of "ja

		b3 := string("select:" + encodeIDBallot("bb") + ":0,0,0,1\n" +
			"text:" + encodeIDBallot("ee") + ":b3Vp\n\n") //encoding of "oui"

		var votesfrontend [3]map[string]interface{}

		fakeConfiguration := fake.BasicConfiguration
		t.Logf("configuration is: %v", fakeConfiguration)

		b1Marshal, _ := UnmarshalBallotManual(b1, fakeConfiguration)
		ballotByte, _ := json.Marshal(b1Marshal)
		_ = json.Unmarshal(ballotByte, &votesfrontend[0])

		b2Marshal, _ := UnmarshalBallotManual(b2, fakeConfiguration)
		ballotByte, _ = json.Marshal(b2Marshal)
		_ = json.Unmarshal(ballotByte, &votesfrontend[1])

		b3Marshal, _ := UnmarshalBallotManual(b3, fakeConfiguration)
		ballotByte, _ = json.Marshal(b3Marshal)
		_ = json.Unmarshal(ballotByte, &votesfrontend[2])

		t.Logf("b123_marshal is: %v", votesfrontend)

		// Ballot 1
		t.Logf("1st ballot in str is: %v", b1)

		ballot1, err := marshallBallotManual(b1, pubKey, Chunksperballot)
		t.Logf("1st ballot is: %v", ballot1)

		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		castVoteRequest := ptypes.CastVoteRequest{
			UserID: "user1",
			Ballot: ballot1,
		}

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]
		t.Logf("cast first ballot to proxy %v", randomproxy)

		t.Logf("vote is: %v", castVoteRequest)
		signed, err = createSignedRequest(secret, castVoteRequest)
		require.NoError(t, err, "failed to sign: %v", err)
		resp, err = http.Post(randomproxy+"/evoting/elections/"+electionID+"/vote", contentType, bytes.NewBuffer(signed))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", resp.Status)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the response of castVoteRequest: %v", err)

		resp.Body.Close()
		t.Log("Response body: " + string(body))

		// Ballot 2
		ballot2, err := marshallBallotManual(b2, pubKey, Chunksperballot)
		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		castVoteRequest = ptypes.CastVoteRequest{
			UserID: "user2",
			Ballot: ballot2,
		}

		t.Logf("cast second ballot")

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		signed, err = createSignedRequest(secret, castVoteRequest)
		require.NoError(t, err, "failed to sign: %v", err)
		resp, err = http.Post(randomproxy+"/evoting/elections/"+electionID+"/vote", contentType, bytes.NewBuffer(signed))
		require.NoError(t, err, "failed to cast ballot: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the response of castVoteRequest: %v", err)

		resp.Body.Close()
		t.Log("Response body: " + string(body))

		// Ballot 3
		ballot3, err := marshallBallotManual(b3, pubKey, Chunksperballot)
		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		castVoteRequest = ptypes.CastVoteRequest{
			UserID: "user3",
			Ballot: ballot3,
		}

		t.Logf("cast third ballot")

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		signed, err = createSignedRequest(secret, castVoteRequest)
		require.NoError(t, err, "failed to sign: %v", err)
		resp, err = http.Post(randomproxy+"/evoting/elections/"+electionID+"/vote", contentType, bytes.NewBuffer(signed))
		require.NoError(t, err, "failed to cast ballot: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the response of castVoteRequest: %v", err)

		resp.Body.Close()
		t.Log("Response body: " + string(body))

		// ############################# CLOSE ELECTION FOR REAL ###################

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		t.Logf("Close election (for real) send to proxy %v", randomproxy)

		_, err = updateElection(secret, randomproxy, electionID, "close", t)
		require.NoError(t, err, "failed to set marshall types.CloseElectionRequest : %v", err)

		getElectionInfo(objmap, proxyAddr1, electionID, t)

		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = int(objmap["Status"].(float64))
		t.Logf("Status of the election : %v", electionStatus)
		require.Equal(t, 2, electionStatus)

		// ###################################### SHUFFLE BALLOTS ##################

		t.Log("shuffle ballots")

		shuffleBallotsRequest := ptypes.UpdateShuffle{
			Action: "shuffle",
		}

		//js, err = json.Marshal(shuffleBallotsRequest)
		signed, err = createSignedRequest(secret, shuffleBallotsRequest)
		require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		oldTime := time.Now()

		req, err = http.NewRequest(http.MethodPut, randomproxy+"/evoting/services/shuffle/"+electionID, bytes.NewBuffer(signed))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)
		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err, "failed to execute the shuffle query: %v", err)

		currentTime := time.Now()
		diff := currentTime.Sub(oldTime)
		t.Logf("Shuffle takes: %v sec", diff.Seconds())

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		time.Sleep(10 * time.Second)

		getElectionInfo(objmap, proxyAddr1, electionID, t)

		electionStatus = int(objmap["Status"].(float64))
		t.Logf("Status of the election : %v", electionStatus)
		require.Equal(t, 3, electionStatus)

		// ###################################### REQUEST PUBLIC SHARES ############

		t.Log("request public shares")

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]
		oldTime = time.Now()

		_, err = updateDKG(secret, randomproxy, electionID, "computePubshares", t)
		require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

		currentTime = time.Now()
		diff = currentTime.Sub(oldTime)
		t.Logf("Shuffle takes: %v sec", diff.Seconds())

		time.Sleep(10 * time.Second)

		getElectionInfo(objmap, proxyAddr1, electionID, t)

		electionStatus = int(objmap["Status"].(float64))
		t.Logf("Status of the election : %v", electionStatus)
		require.Equal(t, 4, electionStatus)

		// ###################################### DECRYPT BALLOTS ##################

		t.Log("decrypt ballots")

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		_, err = updateElection(secret, randomproxy, electionID, "combineShares", t)
		require.NoError(t, err, "failed to combine shares: %v", err)

		time.Sleep(10 * time.Second)

		getElectionInfo(objmap, proxyAddr1, electionID, t)

		electionStatus = int(objmap["Status"].(float64))
		t.Logf("Status of the election : %v", electionStatus)
		require.Equal(t, 5, electionStatus)

		// ###################################### VALIDATE ELECTION RESULT ##############

		tmpBallots := (objmap["Result"]).([]interface{})
		var tmpComp map[string]interface{}
		var tmpCount bool
		for _, ballotIntem := range tmpBallots {
			tmpComp = ballotIntem.(map[string]interface{})
			tmpCount = false
			for _, voteFront := range votesfrontend {
				t.Logf("voteFront: %v", voteFront)
				t.Logf("tmpComp: %v", tmpComp)

				tmpCount = reflect.DeepEqual(tmpComp, voteFront)
				t.Logf("tmpCount: %v", tmpCount)

				if tmpCount {
					break
				}
			}
		}
		require.True(t, tmpCount, "front end votes are different from decrypted votes")

		return

	}
}

func testVariableNode(numNodes int, numVotes int, numElection int) func(*testing.T) {
	return func(t *testing.T) {

		proxyList := make([]string, numNodes)

		for i := 0; i < numNodes; i++ {
			if i < 10 {
				proxyList[i] = "http://localhost:908" + strconv.Itoa(i)
			} else {
				proxyList[i] = "http://localhost:909" + strconv.Itoa(i-10)
			}
			fmt.Println(proxyList[i])
		}

		var wg sync.WaitGroup

		for i := 0; i < numElection; i++ {
			fmt.Println("Starting worker", i)
			wg.Add(1)

			go startElectionProcess(&wg, numNodes, numVotes, proxyList, t, numElection)
			time.Sleep(2 * time.Second)

		}

		fmt.Println("Waiting for workers to finish")
		wg.Wait()

		return

	}
}

func startElectionProcess(wg *sync.WaitGroup, numNodes int, numVotes int, proxyArray []string, t *testing.T, numElection int) {
	defer wg.Done()

	const contentType = "application/json"
	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err, "failed to decode key: %v", err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err, "failed to Unmarshal key: %v", err)

	// ###################################### CREATE SIMPLE ELECTION ######

	t.Logf("Create election")

	configuration := fake.BasicConfiguration

	createSimpleElectionRequest := ptypes.CreateElectionRequest{
		Configuration: configuration,
		AdminID:       "adminId",
	}
	fmt.Println(secret)
	signed, err := createSignedRequest(secret, createSimpleElectionRequest)
	require.NoError(t, err, "fail to create singature")
	resp, err := http.Post(proxyArray[0]+"/evoting/elections", contentType, bytes.NewBuffer(signed))
	require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read the body of the response: %v", err)

	t.Log("response body:", string(body))
	resp.Body.Close()

	var objmap map[string]interface{}

	err = json.Unmarshal(body, &objmap)
	require.NoError(t, err, "failed to parse the body of the response from js: %v", err)
	electionID := objmap["ElectionID"].(string)
	t.Logf("ID of the election : " + electionID)

	// ##################################### SETUP DKG #########################

	t.Log("Init DKG")

	for i := 0; i < len(proxyArray); i++ {
		t.Log("Node" + strconv.Itoa(i))
		fmt.Println(proxyArray[i])
		err = initDKG(secret, proxyArray[i], electionID, t)
		require.NoError(t, err, "failed to init dkg 1: %v", err)
	}

	t.Log("Setup DKG")

	msg := ptypes.UpdateDKG{
		Action: "setup",
	}
	signed, err = createSignedRequest(secret, msg)
	require.NoError(t, err, "failed to sign: %v", err)
	req, err := http.NewRequest(http.MethodPut, proxyArray[0]+"/evoting/services/dkg/actors/"+electionID, bytes.NewBuffer(signed))
	require.NoError(t, err, "failed to create request: %v", err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err, "failed to setup dkg on node 1: %v", err)
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	// ##################################### OPEN ELECTION #####################

	randomproxy := proxyArray[rand.Intn(len(proxyArray))]
	t.Logf("Open election send to proxy %v", randomproxy)

	_, err = updateElection(secret, randomproxy, electionID, "open", t)
	require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
	// ##################################### GET ELECTION INFO #################

	proxyAddr1 := proxyArray[0]

	getElectionInfo(objmap, proxyAddr1, electionID, t)

	electionpubkey := objmap["Pubkey"].(string)
	electionStatus := int(objmap["Status"].(float64))
	BallotSize := int(objmap["BallotSize"].(float64))
	Chunksperballot := ChunksPerBallotManuel(BallotSize)
	t.Logf("Publickey of the election : " + electionpubkey)
	t.Logf("Status of the election : %v", electionStatus)

	require.NoError(t, err, "failed to unmarshal pubkey: %v", err)
	t.Logf("BallotSize of the election : %v", BallotSize)
	t.Logf("Chunksperballot of the election : %v", Chunksperballot)

	// Get election public key

	pubkeyBuf, err := hex.DecodeString(electionpubkey)
	require.NoError(t, err, "failed to decode key: %v", err)

	pubKey := suite.Point()
	err = pubKey.UnmarshalBinary(pubkeyBuf)
	require.NoError(t, err, "failed to Unmarshal key: %v", err)

	// ##################################### CAST BALLOTS ######################

	t.Log("cast ballots")

	//make List of ballots
	b1 := string("select:" + encodeIDBallot("bb") + ":0,0,1,0\n" + "text:" + encodeIDBallot("ee") + ":eWVz\n\n") //encoding of "yes"

	ballotList := make([]string, numVotes)
	for i := 1; i <= numVotes; i++ {
		ballotList[i-1] = b1
	}

	votesfrontend := make([]map[string]interface{}, numVotes)

	fakeConfiguration := fake.BasicConfiguration
	t.Logf("configuration is: %v", fakeConfiguration)

	for i := 0; i < numVotes; i++ {
		bMarshal, _ := UnmarshalBallotManual(ballotList[i], fakeConfiguration)
		ballotByte, _ := json.Marshal(bMarshal)
		_ = json.Unmarshal(ballotByte, &votesfrontend[i])
	}

	t.Logf("b123_marshal is: %v", votesfrontend)

	for i := 0; i < numVotes; i++ {
		t.Logf("1st ballot in str is: %v", ballotList[i])

		ballot, err := marshallBallotManual(ballotList[i], pubKey, Chunksperballot)
		t.Logf("1st ballot is: %v", ballot)

		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		castVoteRequest := ptypes.CastVoteRequest{
			UserID: "user" + strconv.Itoa(i+1),
			Ballot: ballot,
		}

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]
		t.Logf("cast ballot to proxy %v", randomproxy)

		t.Logf("vote is: %v", castVoteRequest)
		signed, err = createSignedRequest(secret, castVoteRequest)
		require.NoError(t, err, "failed to sign: %v", err)
		resp, err = http.Post(randomproxy+"/evoting/elections/"+electionID+"/vote", contentType, bytes.NewBuffer(signed))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", resp.Status)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the response of castVoteRequest: %v", err)

		resp.Body.Close()
		t.Log("Response body: " + string(body))

	}

	// ############################# CLOSE ELECTION FOR REAL ###################
	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	t.Logf("Close election (for real) send to proxy %v", randomproxy)

	_, err = updateElection(secret, randomproxy, electionID, "close", t)
	require.NoError(t, err, "failed to set marshall types.CloseElectionRequest : %v", err)

	getElectionInfo(objmap, proxyAddr1, electionID, t)

	require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
	electionStatus = int(objmap["Status"].(float64))
	t.Logf("Status of the election : %v", electionStatus)
	require.Equal(t, 2, electionStatus)

	// ###################################### SHUFFLE BALLOTS ##################

	t.Log("shuffle ballots")

	shuffleBallotsRequest := ptypes.UpdateShuffle{
		Action: "shuffle",
	}

	//js, err = json.Marshal(shuffleBallotsRequest)
	signed, err = createSignedRequest(secret, shuffleBallotsRequest)
	require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	oldTime := time.Now()

	req, err = http.NewRequest(http.MethodPut, randomproxy+"/evoting/services/shuffle/"+electionID, bytes.NewBuffer(signed))
	require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err, "failed to execute the shuffle query: %v", err)

	currentTime := time.Now()
	diff := currentTime.Sub(oldTime)
	t.Logf("Shuffle takes: %v sec", diff.Seconds())

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read the body of the response: %v", err)

	t.Log("Response body: " + string(body))
	resp.Body.Close()

	time.Sleep(10 * time.Second)

	getElectionInfo(objmap, proxyAddr1, electionID, t)

	electionStatus = int(objmap["Status"].(float64))
	t.Logf("Status of the election : %v", electionStatus)
	require.Equal(t, 3, electionStatus)

	// ###################################### REQUEST PUBLIC SHARES ############

	t.Log("request public shares")

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]
	oldTime = time.Now()

	_, err = updateDKG(secret, randomproxy, electionID, "computePubshares", t)
	require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

	currentTime = time.Now()
	diff = currentTime.Sub(oldTime)
	t.Logf("Shuffle takes: %v sec", diff.Seconds())

	time.Sleep(10 * time.Second)

	getElectionInfo(objmap, proxyAddr1, electionID, t)

	electionStatus = int(objmap["Status"].(float64))
	t.Logf("Status of the election : %v", electionStatus)
	require.Equal(t, 4, electionStatus)

	// ###################################### DECRYPT BALLOTS ##################

	t.Log("decrypt ballots")

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	_, err = updateElection(secret, randomproxy, electionID, "combineShares", t)
	require.NoError(t, err, "failed to combine shares: %v", err)

	time.Sleep(10 * time.Second)

	getElectionInfo(objmap, proxyAddr1, electionID, t)

	electionStatus = int(objmap["Status"].(float64))
	t.Logf("Status of the election : %v", electionStatus)
	require.Equal(t, 5, electionStatus)

	// ###################################### VALIDATE ELECTION RESULT ##############

	tmpBallots := (objmap["Result"]).([]interface{})
	var tmpComp map[string]interface{}
	var tmpCount bool
	for _, ballotIntem := range tmpBallots {
		tmpComp = ballotIntem.(map[string]interface{})
		tmpCount = false
		for _, voteFront := range votesfrontend {
			t.Logf("voteFront: %v", voteFront)
			t.Logf("tmpComp: %v", tmpComp)

			tmpCount = reflect.DeepEqual(tmpComp, voteFront)
			t.Logf("tmpCount: %v", tmpCount)

			if tmpCount {
				break
			}
		}
	}
	require.True(t, tmpCount, "front end votes are different from decrypted votes")

}

// -----------------------------------------------------------------------------
// Utility functions
func marshallBallotManual(voteStr string, pubkey kyber.Point, chunks int) (ptypes.CiphervoteJSON, error) {

	var ballot = make(ptypes.CiphervoteJSON, chunks)
	vote := strings.NewReader(voteStr)
	fmt.Printf("votestr is: %v", voteStr)

	buf := make([]byte, 29)

	for i := 0; i < chunks; i++ {
		var K, C kyber.Point
		var err error

		n, err := vote.Read(buf)
		if err != nil {
			return nil, xerrors.Errorf("failed to read: %v", err)
		}

		K, C, _, err = EncryptManual(buf[:n], pubkey)

		if err != nil {
			return ptypes.CiphervoteJSON{}, xerrors.Errorf("failed to encrypt the plaintext: %v", err)
		}

		kbuff, err := K.MarshalBinary()
		if err != nil {
			return ptypes.CiphervoteJSON{}, xerrors.Errorf("failed to marshal K: %v", err)
		}

		cbuff, err := C.MarshalBinary()
		if err != nil {
			return ptypes.CiphervoteJSON{}, xerrors.Errorf("failed to marshal C: %v", err)
		}

		ballot[i] = ptypes.EGPairJSON{
			K: kbuff,
			C: cbuff,
		}
	}

	return ballot, nil
}

func EncryptManual(message []byte, pubkey kyber.Point) (K, C kyber.Point, remainder []byte, err error) {

	// Embed the message (or as much of it as will fit) into a curve point.
	M := suite.Point().Embed(message, random.New())
	max := suite.Point().EmbedLen()
	if max > len(message) {
		max = len(message)
	}
	remainder = message[max:]
	// ElGamal-encrypt the point to produce ciphertext (K,C).
	k := suite.Scalar().Pick(random.New()) // ephemeral private key
	K = suite.Point().Mul(k, nil)          // ephemeral DH public key
	S := suite.Point().Mul(k, pubkey)      // ephemeral DH shared secret
	C = S.Add(S, M)                        // message blinded with secret

	return K, C, remainder, nil
}

func ChunksPerBallotManuel(BallotSize int) int {
	if BallotSize%29 == 0 {
		return BallotSize / 29
	}

	return BallotSize/29 + 1
}

// Encode implements serde.FormatEngine
func EncodeCiphervote(ciphervote types.Ciphervote) ([]byte, error) {

	m := make(CiphervoteJSON, len(ciphervote))

	for i, egpair := range ciphervote {
		k, err := egpair.K.MarshalBinary()
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal k: %v", err)
		}

		c, err := egpair.C.MarshalBinary()
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal c: %v", err)
		}

		m[i] = EGPairJSON{
			K: k,
			C: c,
		}
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal cipher vote json: %v", err)
	}

	return data, nil
}

// CiphervoteJSON is the JSON representation of a ciphervote
type CiphervoteJSON []EGPairJSON

// EGPairJSON is the JSON representation of an ElGamal pair
type EGPairJSON struct {
	K []byte
	C []byte
}

// Unmarshal decodes the given string according to the format described in
// "state of smart contract.md"
func UnmarshalBallotManual(marshalledBallot string, configuration types.Configuration) (types.Ballot, error) {
	invalidate := func(b types.Ballot) {
		b.RankResultIDs = nil
		b.RankResult = nil
		b.TextResultIDs = nil
		b.TextResult = nil
		b.SelectResultIDs = nil
		b.SelectResult = nil
	}

	var b types.Ballot
	BallotSize := configuration.MaxBallotSize()
	if len(marshalledBallot) > BallotSize {
		invalidate(b)
		return b, fmt.Errorf("ballot has an unexpected size %d, expected <= %d",
			len(marshalledBallot), BallotSize)
	}

	lines := strings.Split(marshalledBallot, "\n")

	b.SelectResultIDs = make([]types.ID, 0)
	b.SelectResult = make([][]bool, 0)

	b.RankResultIDs = make([]types.ID, 0)
	b.RankResult = make([][]int8, 0)

	b.TextResultIDs = make([]types.ID, 0)
	b.TextResult = make([][]string, 0)

	//TODO: Loads of code duplication, can be re-thought
	for _, line := range lines {
		if line == "" {
			// empty line, the valid part of the ballot is over
			break
		}

		question := strings.Split(line, ":")

		if len(question) != 3 {
			invalidate(b)
			return b, xerrors.Errorf("a line in the ballot has length != 3: %s", line)
		}

		_, err := base64.StdEncoding.DecodeString(question[1])
		if err != nil {
			return b, xerrors.Errorf("could not decode question ID: %v", err)
		}
		questionID := question[1]

		q := configuration.GetQuestion(types.ID(questionID))

		if q == nil {
			invalidate(b)
			return b, fmt.Errorf("wrong question ID: the question doesn't exist")
		}

		switch question[0] {

		case "select":
			selections := strings.Split(question[2], ",")

			if len(selections) != q.GetChoicesLength() {
				invalidate(b)
				return b, fmt.Errorf("question %s has a wrong number of answers: expected %d got %d"+
					"", questionID, q.GetChoicesLength(), len(selections))
			}

			b.SelectResultIDs = append(b.SelectResultIDs, types.ID(questionID))
			b.SelectResult = append(b.SelectResult, make([]bool, 0))

			index := len(b.SelectResult) - 1
			var selected uint = 0

			for _, selection := range selections {
				s, err := strconv.ParseBool(selection)

				if err != nil {
					invalidate(b)
					return b, fmt.Errorf("could not parse selection value for Q.%s: %v",
						questionID, err)
				}

				if s {
					selected++
				}

				b.SelectResult[index] = append(b.SelectResult[index], s)
			}

			if selected > q.GetMaxN() {
				invalidate(b)
				return b, fmt.Errorf("question %s has too many selected answers", questionID)
			} else if selected < q.GetMinN() {
				invalidate(b)
				return b, fmt.Errorf("question %s has not enough selected answers", questionID)
			}

		case "rank":
			ranks := strings.Split(question[2], ",")

			if len(ranks) != q.GetChoicesLength() {
				invalidate(b)
				return b, fmt.Errorf("question %s has a wrong number of answers: expected %d got %d"+
					"", questionID, q.GetChoicesLength(), len(ranks))
			}

			b.RankResultIDs = append(b.RankResultIDs, types.ID(questionID))
			b.RankResult = append(b.RankResult, make([]int8, 0))

			index := len(b.RankResult) - 1
			var selected uint = 0
			for _, rank := range ranks {
				if len(rank) > 0 {
					selected++

					r, err := strconv.ParseInt(rank, 10, 8)
					if err != nil {
						invalidate(b)
						return b, fmt.Errorf("could not parse rank value for Q.%s : %v",
							questionID, err)
					}

					if r < 0 || uint(r) >= q.GetMaxN() {
						invalidate(b)
						return b, fmt.Errorf("invalid rank not in range [0, MaxN[")
					}

					b.RankResult[index] = append(b.RankResult[index], int8(r))
				} else {
					b.RankResult[index] = append(b.RankResult[index], int8(-1))
				}
			}

			if selected > q.GetMaxN() {
				invalidate(b)
				return b, fmt.Errorf("question %s has too many selected answers", questionID)
			} else if selected < q.GetMinN() {
				invalidate(b)
				return b, fmt.Errorf("question %s has not enough selected answers", questionID)
			}

		case "text":
			texts := strings.Split(question[2], ",")

			if len(texts) != q.GetChoicesLength() {
				invalidate(b)
				return b, fmt.Errorf("question %s has a wrong number of answers: expected %d got %d"+
					"", questionID, q.GetChoicesLength(), len(texts))
			}

			b.TextResultIDs = append(b.TextResultIDs, types.ID(questionID))
			b.TextResult = append(b.TextResult, make([]string, 0))

			index := len(b.TextResult) - 1
			var selected uint = 0

			for _, text := range texts {
				if len(text) > 0 {
					selected++
				}

				t, err := base64.StdEncoding.DecodeString(text)
				if err != nil {
					return b, fmt.Errorf("could not decode text for Q. %s: %v", questionID, err)
				}

				b.TextResult[index] = append(b.TextResult[index], string(t))
			}

			if selected > q.GetMaxN() {
				invalidate(b)
				return b, fmt.Errorf("question %s has too many selected answers", questionID)
			} else if selected < q.GetMinN() {
				invalidate(b)
				return b, fmt.Errorf("question %s has not enough selected answers", questionID)
			}

		default:
			invalidate(b)
			return b, fmt.Errorf("question type is unknown")
		}

	}

	return b, nil
}

func encodeIDBallot(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
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
func initDKG(secret kyber.Scalar, proxyAddr, electionIDHex string, t *testing.T) error {
	setupDKG := ptypes.NewDKGRequest{
		ElectionID: electionIDHex,
	}

	signed, err := createSignedRequest(secret, setupDKG)
	require.NoError(t, err, "fail to create singature")

	resp, err := http.Post(proxyAddr+"/evoting/services/dkg/actors", "application/json", bytes.NewBuffer(signed))
	if err != nil {
		return xerrors.Errorf("failed to post request: %v", err)
	}
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	return nil
}

func updateElection(secret kyber.Scalar, proxyAddr, electionIDHex, action string, t *testing.T) (int, error) {
	msg := ptypes.UpdateElectionRequest{
		Action: action,
	}

	signed, err := createSignedRequest(secret, msg)
	require.NoError(t, err, "fail to create singature")

	req, err := http.NewRequest(http.MethodPut, proxyAddr+"/evoting/elections/"+electionIDHex, bytes.NewBuffer(signed))
	if err != nil {
		return 0, xerrors.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	return 0, nil
}
func updateDKG(secret kyber.Scalar, proxyAddr, electionIDHex, action string, t *testing.T) (int, error) {
	msg := ptypes.UpdateDKG{
		Action: action,
	}

	signed, err := createSignedRequest(secret, msg)
	require.NoError(t, err, "fail to create singature")

	req, err := http.NewRequest(http.MethodPut, proxyAddr+"/evoting/services/dkg/actors/"+electionIDHex, bytes.NewBuffer(signed))
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

func getElectionInfo(objmap map[string]interface{}, proxyAddr, electionID string, t *testing.T) {
	t.Log("Get election info")
	resp, err := http.Get(proxyAddr + "/evoting/elections" + "/" + electionID)

	require.NoError(t, err, "failed to get the election: %v", err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read the body of the response: %v", err)

	t.Log("response body:", string(body))
	resp.Body.Close()
	err = json.Unmarshal(body, &objmap)
	require.NoError(t, err, "failed to parse body of the response from js: %v", err)

}
