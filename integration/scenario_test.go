package integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/xerrors"
)

var suite = suites.MustFind("Ed25519")

// Check the shuffled votes versus the cast votes on a few nodes
func TestScenario(t *testing.T) {
	t.Run("Basic configuration", getScenarioTest())

}

func getScenarioTest() func(*testing.T) {
	return func(t *testing.T) {
		const (
			loginEndpoint               = "/evoting/login"
			createElectionEndpoint      = "/evoting/create"
			openElectionEndpoint        = "/evoting/open"
			castVoteEndpoint            = "/evoting/cast"
			getAllElectionsIdsEndpoint  = "/evoting/allids"
			getElectionInfoEndpoint     = "/evoting/info"
			getAllElectionsInfoEndpoint = "/evoting/all"
			closeElectionEndpoint       = "/evoting/close"
			shuffleBallotsEndpoint      = "/evoting/shuffle"
			beginDecryptionEndpoint     = "/evoting/beginDecryption"
			combineSharesEndpoint       = "/evoting/combineShares"
			getElectionResultEndpoint   = "/evoting/result"
			cancelElectionEndpoint      = "/evoting/cancel"
			initEndpoint                = "/evoting/dkg/init"
		)

		const contentType = "application/json"

		t.Parallel()
		proxyAddr1 := "http://localhost:8081"
		proxyAddr2 := "http://localhost:8082"
		proxyAddr3 := "http://localhost:8083"

		proxyArray := [3]string{proxyAddr1, proxyAddr2, proxyAddr3}

		// ###################################### CREATE SIMPLE ELECTION ######
		create_election_js := `{"Configuration":{"MainTitle":"electionTitle","Scaffold":[{"ID":"YWE=","Title":"subject1","Order":null,"Subjects":null,"Selects":[{"ID":"YmI=","Title":"Select your favorite snacks","MaxN":3,"MinN":0,"Choices":["snickers","mars","vodka","babibel"]}],"Ranks":[],"Texts":null},{"ID":"ZGQ=","Title":"subject2","Order":null,"Subjects":null,"Selects":null,"Ranks":null,"Texts":[{"ID":"ZWU=","Title":"dissertation","MaxN":1,"MinN":1,"MaxLength":3,"Regex":"","Choices":["write yes in your language"]}]}]},"AdminID":"adminId"}`
		t.Logf("Create election")
		t.Logf("create election js: %v", create_election_js)

		resp, err := http.Post(proxyAddr1+createElectionEndpoint, contentType, bytes.NewBuffer([]byte(create_election_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()

		// var payload interface{}
		var objmap map[string]interface{}

		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parse the body of the response from js: %v", err)
		electionID := objmap["ElectionID"].(string)
		t.Logf("ID of the election : " + electionID)

		// ##################################### SETUP DKG #########################

		t.Log("Init DKG")

		t.Log("Node 1")

		resp, err = http.Post(proxyAddr1+initEndpoint, contentType, bytes.NewBuffer([]byte(electionID)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		t.Log("Node 2")
		resp, err = http.Post(proxyAddr2+initEndpoint, contentType, bytes.NewBuffer([]byte(electionID)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		t.Log("Node 3")
		resp, err = http.Post(proxyAddr3+initEndpoint, contentType, bytes.NewBuffer([]byte(electionID)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		t.Log("Setup DKG")
		resp, err = http.Post(proxyAddr1+"/evoting/dkg/setup", contentType, bytes.NewBuffer([]byte(electionID)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		pubkeyBuf, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read body: %v", err)
		t.Logf("DKG public key: %x", pubkeyBuf)

		pubKey := suite.Point()
		err = pubKey.UnmarshalBinary(pubkeyBuf)
		require.NoError(t, err, "failed to unmarshal pubkey: %v", err)
		t.Logf("Pubkey: %v\n", pubKey)

		// ##################################### OPEN ELECTION #####################

		randomproxy := proxyArray[rand.Intn(len(proxyArray))]
		t.Logf("Open election send to proxy %v", randomproxy)

		resp, err = http.Post(randomproxy+"/evoting/open", contentType, bytes.NewBuffer([]byte(electionID)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		// ##################################### GET ELECTION INFO #################
		// Get election public key

		t.Log("Get election info")
		create_info_js := fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		resp, err = http.Post(proxyAddr1+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parse body of the response from js: %v", err)
		electionpubkey := objmap["Pubkey"].(string)
		electionStatus := int(objmap["Status"].(float64))
		BallotSize := int(objmap["BallotSize"].(float64))
		Chunksperballot := ChunksPerBallot_manuel(BallotSize)
		t.Logf("Publickey of the election : " + electionpubkey)
		t.Logf("Status of the election : %v", electionStatus)
		t.Logf("BallotSize of the election : %v", BallotSize)
		t.Logf("Chunksperballot of the election : %v", Chunksperballot)

		// ##################################### CAST BALLOTS ######################

		t.Log("cast ballots")

		// Create the ballots
		b1 := string("select:" + encodeID("bb") + ":0,0,1,0\n" +
			"text:" + encodeID("ee") + ":eWVz\n\n") //encoding of "yes"

		b2 := string("select:" + encodeID("bb") + ":1,1,0,0\n" +
			"text:" + encodeID("ee") + ":amE=\n\n") //encoding of "ja

		b3 := string("select:" + encodeID("bb") + ":0,0,0,1\n" +
			"text:" + encodeID("ee") + "b3Vp\n\n") //encoding of "oui"

		var votesfrontend [3]map[string]interface{}

		fake_configuration := fake.BasicConfiguration
		t.Logf("configuration is: %v", fake_configuration)

		b1_marshal, _ := Unmarshal_Ballot_Manual(b1, fake_configuration)
		ballot_byte, _ := json.Marshal(b1_marshal)
		// var temp_obj map[string]interface{}
		_ = json.Unmarshal(ballot_byte, &votesfrontend[0])
		// t.Logf("b1_marshal is: %v", temp_obj)

		b2_marshal, _ := Unmarshal_Ballot_Manual(b2, fake_configuration)
		ballot_byte, _ = json.Marshal(b2_marshal)
		_ = json.Unmarshal(ballot_byte, &votesfrontend[1])
		// t.Logf("b2_marshal is: %v", temp_obj)
		// votesfrontend[1] = temp_obj

		b3_marshal, _ := Unmarshal_Ballot_Manual(b3, fake_configuration)
		ballot_byte, _ = json.Marshal(b3_marshal)
		_ = json.Unmarshal(ballot_byte, &votesfrontend[2])
		// t.Logf("b1_marshal is: %v", temp_obj)
		// votesfrontend[2] = temp_obj
		t.Logf("b123_marshal is: %v", votesfrontend)

		// Ballot 1
		t.Logf("1st ballot in str is: %v", b1)

		ballot1, err := marshallBallot_manual(b1, pubKey, Chunksperballot)
		t.Logf("1st ballot is: %v", ballot1)

		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		data1, err := Encode_ciphervote(ballot1)
		require.NoError(t, err, "failed to marshall ballot : %v", err)
		t.Logf("1st marshalled ballot is: %v", data1)

		castVoteRequest := types.CastVoteRequest{
			ElectionID: electionID,
			UserID:     "user1",
			Ballot:     data1,
			Token:      "token",
		}

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]
		t.Logf("cast first ballot to proxy %v", randomproxy)
		js_vote, err := json.Marshal(castVoteRequest)
		require.NoError(t, err, "failed to marshal castVoteRequest: %v", err)

		t.Logf("vote is: %v", castVoteRequest)
		resp, err = http.Post(randomproxy+castVoteEndpoint, contentType, bytes.NewBuffer(js_vote))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", resp.Status)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the response of castVoteRequest: %v", err)

		resp.Body.Close()
		t.Log("Response body: " + string(body))

		// Ballot 2
		ballot2, err := marshallBallot_manual(b2, pubKey, Chunksperballot)
		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		data2, err := Encode_ciphervote(ballot2)
		require.NoError(t, err, "failed to marshall ballot : %v", err)

		castVoteRequest = types.CastVoteRequest{
			ElectionID: electionID,
			UserID:     "user2",
			Ballot:     data2,
			Token:      "token",
		}

		t.Logf("cast second ballot")
		js_vote, err = json.Marshal(castVoteRequest)
		require.NoError(t, err, "failed to marshal castVoteRequest: %v", err)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]
		resp, err = http.Post(randomproxy+castVoteEndpoint, contentType, bytes.NewBuffer(js_vote))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the response of castVoteRequest: %v", err)

		resp.Body.Close()
		t.Log("Response body: " + string(body))

		// Ballot 3
		ballot3, err := marshallBallot_manual(b3, pubKey, Chunksperballot)
		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		data3, err := Encode_ciphervote(ballot3)
		require.NoError(t, err, "failed to marshall ballot : %v", err)

		castVoteRequest = types.CastVoteRequest{
			ElectionID: electionID,
			UserID:     "user3",
			Ballot:     data3,
			Token:      "token",
		}

		t.Logf("cast third ballot")
		js_vote, err = json.Marshal(castVoteRequest)
		require.NoError(t, err, "failed to marshal castVoteRequest: %v", err)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		resp, err = http.Post(randomproxy+castVoteEndpoint, contentType, bytes.NewBuffer(js_vote))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the response of castVoteRequest: %v", err)

		resp.Body.Close()
		t.Log("Response body: " + string(body))

		// ############################# CLOSE ELECTION FOR REAL ###################

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		t.Logf("Close election (for real) send to proxy %v", randomproxy)

		closeElectionRequest := types.CloseElectionRequest{
			ElectionID: electionID,
			UserID:     "adminId",
			Token:      "token",
		}

		js, err := json.Marshal(closeElectionRequest)
		require.NoError(t, err, "failed to set marshall types.CloseElectionRequest : %v", err)

		resp, err = http.Post(randomproxy+closeElectionEndpoint, contentType, bytes.NewBuffer(js))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		t.Log("Get election info")
		create_info_js = fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		resp, err = http.Post(randomproxy+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = int(objmap["Status"].(float64))
		t.Logf("Status of the election : %v", electionStatus)

		// ###################################### SHUFFLE BALLOTS ##################

		t.Log("shuffle ballots")

		shuffleBallotsRequest := types.ShuffleBallotsRequest{
			ElectionID: electionID,
			UserID:     "adminId",
			Token:      "token",
		}

		js, err = json.Marshal(shuffleBallotsRequest)
		require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		resp, err = http.Post(randomproxy+shuffleBallotsEndpoint, contentType, bytes.NewBuffer(js))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		time.Sleep(10 * time.Second)

		t.Log("Get election info")
		create_info_js = fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		resp, err = http.Post(randomproxy+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = int(objmap["Status"].(float64))
		t.Logf("Status of the election : %v", electionStatus)

		// ###################################### REQUEST PUBLIC SHARES ############

		t.Log("request public shares")

		beginDecryptionRequest := types.BeginDecryptionRequest{
			ElectionID: electionID,
			UserID:     "adminId",
			Token:      "token",
		}

		js, err = json.Marshal(beginDecryptionRequest)
		require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		resp, err = http.Post(randomproxy+beginDecryptionEndpoint, contentType, bytes.NewBuffer(js))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		time.Sleep(10 * time.Second)

		t.Log("Get election info")
		create_info_js = fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		resp, err = http.Post(randomproxy+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = int(objmap["Status"].(float64))
		t.Logf("Status of the election : %v", electionStatus)

		// ###################################### DECRYPT BALLOTS ##################

		t.Log("decrypt ballots")

		decryptBallotsRequest := types.CombineSharesRequest{
			ElectionID: electionID,
			UserID:     "adminId",
			Token:      "token",
		}

		js, err = json.Marshal(decryptBallotsRequest)
		require.NoError(t, err, "failed to set marshall types.CombineSharesRequest : %v", err)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		resp, err = http.Post(randomproxy+combineSharesEndpoint, contentType, bytes.NewBuffer(js))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		time.Sleep(10 * time.Second)

		t.Log("Get election info")
		create_info_js = fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		resp, err = http.Post(randomproxy+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = int(objmap["Status"].(float64))
		t.Logf("Status of the election : %v", electionStatus)

		// ###################################### GET ELECTION RESULT ##############

		t.Log("Get election result")

		getElectionResultRequest := types.GetElectionResultRequest{
			ElectionID: electionID,
			Token:      "token",
		}

		js, err = json.Marshal(getElectionResultRequest)
		require.NoError(t, err, "failed to set marshall types.GetElectionResultRequest : %v", err)

		randomproxy = proxyArray[rand.Intn(len(proxyArray))]

		resp, err = http.Post(randomproxy+getElectionResultEndpoint, contentType, bytes.NewBuffer(js))

		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		// ###################################### VALIDATE ELECTION RESULT ##############

		// body1 := `{"Result":[{"SelectResultIDs":["YmI="],"SelectResult":[[false,false,true,false]],"RankResultIDs":[],"RankResult":[],"TextResultIDs":["ZWU="],"TextResult":[["yes"]]},{"SelectResultIDs":["YmI="],"SelectResult":[[true,true,false,false]],"RankResultIDs":[],"RankResult":[],"TextResultIDs":["ZWU="],"TextResult":[["ja"]]},{"SelectResultIDs":null,"SelectResult":null,"RankResultIDs":null,"RankResult":null,"TextResultIDs":null,"TextResult":null}]}`

		// var objmap1 map[string]interface{}
		// _ = json.Unmarshal([]byte(body1), &objmap1)
		// tmp_ballots := (objmap1["Result"]).([]interface{})
		// tmp_map := tmp_ballots[0].(map[string]interface{})

		// require.Equal(t, temp_obj, tmp_map)
		// return

		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)

		tmp_ballots := (objmap["Result"]).([]interface{})
		var tmp_comp map[string]interface{}
		var tmp_count bool
		for _, ballot_intem := range tmp_ballots {
			tmp_comp = ballot_intem.(map[string]interface{})
			tmp_count = false
			for _, vote_front := range votesfrontend {
				t.Logf("vote_front: %v", vote_front)
				t.Logf("tmp_comp: %v", tmp_comp)

				tmp_count = reflect.DeepEqual(tmp_comp, vote_front)
				t.Logf("tmp_count: %v", tmp_count)

				if tmp_count {
					break
				}
			}
		}
		require.True(t, tmp_count, "front end votes are different from decrypted votes")

		// // tmp_ballots := (objmap["Result"]).([]types.Ballot)

		// t.Logf("Response body tmp_ballots is %v", tmp_ballots[0])
		// b_test, _ := tmp_ballots[0]["RankResult"]
		// t.Logf("Response body tmp_ballots RankResult is %v", b_test)

		// // var sendback_ballots types.Ballot
		// // _ = json.Unmarshal(b_test, sendback_ballots)
		// // t.Logf("Response body unmarshalled is %v", sendback_ballots)
		return

	}
}

// -----------------------------------------------------------------------------
// Utility functions
func marshallBallot_manual(voteStr string, pubkey kyber.Point, chunks int) (types.Ciphervote, error) {

	var ballot = make(types.Ciphervote, chunks)
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

		K, C, _, err = Encrypt_manual(buf[:n], pubkey)

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

func Encrypt_manual(message []byte, pubkey kyber.Point) (K, C kyber.Point, remainder []byte,
	err error) {

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

func ChunksPerBallot_manuel(BallotSize int) int {
	if BallotSize%29 == 0 {
		return BallotSize / 29
	}

	return BallotSize/29 + 1
}

// Encode implements serde.FormatEngine
func Encode_ciphervote(ciphervote types.Ciphervote) ([]byte, error) {

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
func Unmarshal_Ballot_Manual(marshalledBallot string, configuration types.Configuration) (types.Ballot, error) {
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
