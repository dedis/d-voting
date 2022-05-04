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
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	ptypes "github.com/dedis/d-voting/proxy/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/encoding"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/xerrors"
)

var suite = suites.MustFind("Ed25519")

// Check the shuffled votes versus the cast votes on a few nodes
func TestScenario(t *testing.T) {
	t.Run("Basic configuration", getScenarioTest(3, 3, 1))
}

func getScenarioTest(numNodes int, numVotes int, numElection int) func(*testing.T) {
	return func(t *testing.T) {

		proxyList := make([]string, numNodes)

		for i := 0; i < numNodes; i++ {
			proxyList[i] = fmt.Sprintf("http://localhost:%v", 9081+i)
			t.Log(proxyList[i])
		}

		var wg sync.WaitGroup

		for i := 0; i < numElection; i++ {
			t.Log("Starting worker", i)
			wg.Add(1)

			go startElectionProcess(&wg, numNodes, numVotes, proxyList, t, numElection)
			time.Sleep(2 * time.Second)

		}

		t.Log("Waiting for workers to finish")
		wg.Wait()

	}
}

func startElectionProcess(wg *sync.WaitGroup, numNodes int, numVotes int, proxyArray []string, t *testing.T, numElection int) {
	defer wg.Done()
	rand.Seed(0)

	const contentType = "application/json"
	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err, "failed to decode key: %v", err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err, "failed to Unmarshal key: %v", err)

	// ###################################### CREATE SIMPLE ELECTION ######

	t.Log("Create election")

	configuration := fake.BasicConfiguration

	createSimpleElectionRequest := ptypes.CreateElectionRequest{
		Configuration: configuration,
		AdminID:       "adminId",
	}

	signed, err := createSignedRequest(secret, createSimpleElectionRequest)
	require.NoError(t, err, "failed to create signature")

	resp, err := http.Post(proxyArray[0]+"/evoting/elections", contentType, bytes.NewBuffer(signed))
	require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read the response body: %v", err)

	t.Log("response body:", string(body))
	resp.Body.Close()

	var createElectionResponse ptypes.CreateElectionResponse

	err = json.Unmarshal(body, &createElectionResponse)
	require.NoError(t, err, "failed to parse the response body from js: %v", err)

	electionID := createElectionResponse.ElectionID

	t.Logf("ID of the election : " + electionID)

	// ##################################### SETUP DKG #########################

	t.Log("Init DKG")

	for i := 0; i < numNodes; i++ {
		t.Log("Node" + strconv.Itoa(i+1))
		t.Log(proxyArray[i])
		err = initDKG(secret, proxyArray[i], electionID, t)
		require.NoError(t, err, "failed to init dkg: %v", err)
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
	time.Sleep(time.Second * 3)

	getElectionResponse := getElectionInfo(proxyAddr1, electionID, t)
	electionpubkey := getElectionResponse.Pubkey
	electionStatus := getElectionResponse.Status
	BallotSize := getElectionResponse.BallotSize
	Chunksperballot := chunksPerBallot(BallotSize)

	t.Logf("Publickey of the election : " + electionpubkey)
	t.Logf("Status of the election : %v", electionStatus)

	require.NoError(t, err, "failed to unmarshal pubkey: %v", err)
	t.Logf("BallotSize of the election : %v", BallotSize)
	t.Logf("Chunksperballot of the election : %v", Chunksperballot)

	// Get election public key
	pubKey, err := encoding.StringHexToPoint(suite, electionpubkey)
	require.NoError(t, err, "failed to Unmarshal key: %v", err)

	// ##################################### CAST BALLOTS ######################
	t.Log("cast ballots")

	//make List of ballots
	b1 := string("select:" + encodeIDBallot("bb") + ":0,0,1,0\n" + "text:" + encodeIDBallot("ee") + ":eWVz\n\n") //encoding of "yes"

	ballotList := make([]string, numVotes)
	for i := 1; i <= numVotes; i++ {
		ballotList[i-1] = b1
	}

	votesfrontend := make([]types.Ballot, numVotes)

	fakeConfiguration := fake.BasicConfiguration
	t.Logf("configuration is: %v", fakeConfiguration)

	for i := 0; i < numVotes; i++ {

		var bMarshal types.Ballot
		election := types.Election{
			Configuration: fakeConfiguration,
			ElectionID:    electionID,
			BallotSize:    BallotSize,
		}

		err = bMarshal.Unmarshal(ballotList[i], election)
		require.NoError(t, err, "failed to unmarshal ballot : %v", err)

		votesfrontend[i] = bMarshal
	}

	for i := 0; i < numVotes; i++ {
		t.Logf("ballot in str is: %v", ballotList[i])

		ballot, err := marshallBallotManual(ballotList[i], pubKey, Chunksperballot)
		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		t.Logf("ballot is: %v", ballot)

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

	time.Sleep(time.Second * 3)

	getElectionResponse = getElectionInfo(proxyAddr1, electionID, t)
	electionStatus = getElectionResponse.Status

	t.Logf("Status of the election : %v", electionStatus)
	require.Equal(t, uint16(2), electionStatus)

	// ###################################### SHUFFLE BALLOTS ##################

	t.Log("shuffle ballots")

	shuffleBallotsRequest := ptypes.UpdateShuffle{
		Action: "shuffle",
	}

	signed, err = createSignedRequest(secret, shuffleBallotsRequest)
	require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]

	timeTable := make([]float64, 3)
	oldTime := time.Now()

	req, err = http.NewRequest(http.MethodPut, randomproxy+"/evoting/services/shuffle/"+electionID, bytes.NewBuffer(signed))
	require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
	require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err, "failed to execute the shuffle query: %v", err)

	currentTime := time.Now()
	diff := currentTime.Sub(oldTime)
	timeTable[0] = diff.Seconds()
	t.Logf("Shuffle takes: %v sec", diff.Seconds())

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read the response body: %v", err)

	t.Log("Response body: " + string(body))
	resp.Body.Close()

	getElectionResponse = getElectionInfo(proxyAddr1, electionID, t)
	electionStatus = getElectionResponse.Status

	t.Logf("Status of the election : %v", electionStatus)
	require.Equal(t, uint16(3), electionStatus)

	// ###################################### REQUEST PUBLIC SHARES ############

	t.Log("request public shares")

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]
	oldTime = time.Now()

	_, err = updateDKG(secret, randomproxy, electionID, "computePubshares", t)
	require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

	currentTime = time.Now()
	diff = currentTime.Sub(oldTime)
	timeTable[1] = diff.Seconds()

	t.Logf("Request public share takes: %v sec", diff.Seconds())

	time.Sleep(10 * time.Second)

	getElectionResponse = getElectionInfo(proxyAddr1, electionID, t)
	electionStatus = getElectionResponse.Status

	t.Logf("Status of the election : %v", electionStatus)
	require.Equal(t, uint16(4), electionStatus)

	// ###################################### DECRYPT BALLOTS ##################

	t.Log("decrypt ballots")

	randomproxy = proxyArray[rand.Intn(len(proxyArray))]
	oldTime = time.Now()

	_, err = updateElection(secret, randomproxy, electionID, "combineShares", t)
	require.NoError(t, err, "failed to combine shares: %v", err)

	currentTime = time.Now()
	diff = currentTime.Sub(oldTime)
	timeTable[2] = diff.Seconds()

	t.Logf("decryption takes: %v sec", diff.Seconds())

	time.Sleep(time.Second * 3)

	getElectionResponse = getElectionInfo(proxyAddr1, electionID, t)
	electionStatus = getElectionResponse.Status

	t.Logf("Status of the election : %v", electionStatus)
	require.Equal(t, uint16(5), electionStatus)

	//#################################### VALIDATE ELECTION RESULT ##############

	tmpBallots := getElectionResponse.Result
	var tmpCount bool

	for _, ballotIntem := range tmpBallots {
		tmpComp := ballotIntem
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
	t.Logf("shuffle time : %v", timeTable[0])
	t.Logf("Public share time : %v", timeTable[1])
	t.Logf("decryption time : %v", timeTable[2])
}

// -----------------------------------------------------------------------------
// Utility functions
func marshallBallotManual(voteStr string, pubkey kyber.Point, chunks int) (ptypes.CiphervoteJSON, error) {

	ballot := make(ptypes.CiphervoteJSON, chunks)
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

		K, C, _, err = encryptManual(buf[:n], pubkey)

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

func encryptManual(message []byte, pubkey kyber.Point) (K, C kyber.Point, remainder []byte, err error) {

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

func chunksPerBallot(size int) int { return (size-1)/29 + 1 }

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
	require.NoError(t, err, "failed to create signature")

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
	require.NoError(t, err, "failed to create signature")

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
	require.NoError(t, err, "failed to create signature")

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

func getElectionInfo(proxyAddr, electionID string, t *testing.T) ptypes.GetElectionResponse {
	t.Log("Get election info")

	resp, err := http.Get(proxyAddr + "/evoting/elections" + "/" + electionID)
	require.NoError(t, err, "failed to get the election: %v", err)

	var infoElection ptypes.GetElectionResponse
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&infoElection)
	require.NoError(t, err, "failed to decode getInfoElection: %v", err)

	resp.Body.Close()

	return infoElection

}
