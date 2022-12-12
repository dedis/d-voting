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
	"os"
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

const defaultNodes = 5

// Check the shuffled votes versus the cast votes on a few nodes
func TestScenario(t *testing.T) {
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

			go startFormProcess(&wg, numNodes, numVotes, proxyList, t, numForm)
			time.Sleep(2 * time.Second)

		}

		t.Log("Waiting for workers to finish")
		wg.Wait()

	}
}

func startFormProcess(wg *sync.WaitGroup, numNodes int, numVotes int, proxyArray []string, t *testing.T, numForm int) {
	defer wg.Done()
	rand.Seed(0)

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

	t.Log("cast ballots")

	oldTime = time.Now()

	//make List of ballots
	b1 := string("select:" + encodeIDBallot("bb") + ":0,0,1,0\n" + "text:" + encodeIDBallot("ee") + ":eWVz\n\n") //encoding of "yes"

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

		err = bMarshal.Unmarshal(ballotList[i], form)
		require.NoError(t, err)

		votesfrontend[i] = bMarshal
	}

	for i := 0; i < numVotes; i++ {

		ballot, err := marshallBallotManual(ballotList[i], pubKey, Chunksperballot)
		require.NoError(t, err)

		castVoteRequest := ptypes.CastVoteRequest{
			UserID: "user" + strconv.Itoa(i+1),
			Ballot: ballot,
		}

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]
		t.Logf("cast ballot to proxy %v", randomproxy)

		signed, err := createSignedRequest(secret, castVoteRequest)
		require.NoError(t, err)

		resp, err := http.Post(randomproxy+"/evoting/forms/"+formID+"/vote", contentType, bytes.NewBuffer(signed))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", resp.Status)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var infos ptypes.TransactionInfoToSend
		err = json.Unmarshal(body, &infos)
		require.NoError(t, err)

		ok, err = pollTxnInclusion(randomproxy, infos.Token, t)
		require.NoError(t, err)
		require.True(t, ok)

		resp.Body.Close()

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
	t.Logf("Creating the form takes %v sec", timeTable[0])
	t.Logf("Setting up DKG takes %v sec", timeTable[1])
	t.Logf("Oppening the form takes %v sec", timeTable[2])
	t.Logf("Casting %v ballots takes %v sec", numVotes, timeTable[3])
	t.Logf("Closing the form takes %v sec", timeTable[4])
	t.Logf("Shuffling ballots takes %v sec", timeTable[5])
	t.Logf("Requesting public shares takes %v sec", timeTable[6])
	t.Logf("Decrypting ballots takes %v sec", timeTable[7])

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
