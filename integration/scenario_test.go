package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
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
		return
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

		t.Log("Open election")
		resp, err = http.Post(proxyAddr1+"/evoting/open", contentType, bytes.NewBuffer([]byte(electionID)))
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
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionpubkey := objmap["Pubkey"].(string)
		electionStatus := objmap["Status"].(string)
		t.Logf("Publickey of the election : " + electionpubkey)
		t.Logf("Status of the election : " + electionStatus)
		// ##################################### CAST BALLOTS ######################

		t.Log("cast ballots")

		// Create the ballots
		b1 := string("select:" + encodeID("bb") + ":0,0,1,0\n" +
			"text:" + encodeID("ee") + ":eWVz\n\n") //encoding of "yes"

		b2 := string("select:" + encodeID("bb") + ":1,1,0,0\n" +
			"text:" + encodeID("ee") + ":amE=\n\n") //encoding of "ja

		b3 := string("select:" + encodeID("bb") + ":0,0,0,1\n" +
			"text:" + encodeID("ee") + "b3Vp\n\n") //encoding of "oui"

		// Ballot 1
		// chunk 255 by default
		ballot1, err := marshallBallot_manual(b1, pubKey, 255)
		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		data1, err := json.Marshal(ballot1)
		require.NoError(t, err, "failed to marshall ballot : %v", err)

		castVoteRequest := types.CastVoteRequest{
			ElectionID: electionID,
			UserID:     "user1",
			Ballot:     data1,
			Token:      "token",
		}

		t.Logf("cast first ballot")
		js_vote, err := json.Marshal(castVoteRequest)
		resp, err = http.Post(proxyAddr1+castVoteEndpoint, contentType, bytes.NewBuffer(js_vote))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Log("Response body: " + string(body))

		// Ballot 2
		// chunk 255 by default
		ballot2, err := marshallBallot_manual(b2, pubKey, 255)
		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		data2, err := json.Marshal(ballot2)
		require.NoError(t, err, "failed to marshall ballot : %v", err)

		castVoteRequest = types.CastVoteRequest{
			ElectionID: electionID,
			UserID:     "user2",
			Ballot:     data2,
			Token:      "token",
		}

		t.Logf("cast second ballot")
		js_vote, _ = json.Marshal(castVoteRequest)
		resp, err = http.Post(proxyAddr1+castVoteEndpoint, contentType, bytes.NewBuffer(js_vote))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)
		body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Log("Response body: " + string(body))

		// Ballot 3
		// chunk 255 by default
		ballot3, err := marshallBallot_manual(b3, pubKey, 255)
		require.NoError(t, err, "failed to encrypt ballot : %v", err)

		data3, err := json.Marshal(ballot3)
		require.NoError(t, err, "failed to marshall ballot : %v", err)

		castVoteRequest = types.CastVoteRequest{
			ElectionID: electionID,
			UserID:     "user3",
			Ballot:     data3,
			Token:      "token",
		}

		t.Logf("cast third ballot")
		js_vote, err = json.Marshal(castVoteRequest)
		resp, err = http.Post(proxyAddr1+castVoteEndpoint, contentType, bytes.NewBuffer(js_vote))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Log("Response body: " + string(body))

		// ############################# CLOSE ELECTION FOR REAL ###################

		t.Log("Close election (for real)")

		closeElectionRequest := types.CloseElectionRequest{
			ElectionID: electionID,
			UserID:     "adminId",
			Token:      "token",
		}

		js, err := json.Marshal(closeElectionRequest)
		require.NoError(t, err, "failed to set marshall types.CloseElectionRequest : %v", err)
		resp, err = http.Post(proxyAddr1+closeElectionEndpoint, contentType, bytes.NewBuffer(js))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		t.Log("Get election info")
		create_info_js = fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		resp, err = http.Post(proxyAddr1+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = objmap["Status"].(string)
		t.Logf("Status of the election : " + electionStatus)

		// ###################################### SHUFFLE BALLOTS ##################

		t.Log("shuffle ballots")

		shuffleBallotsRequest := types.ShuffleBallotsRequest{
			ElectionID: electionID,
			UserID:     "adminId",
			Token:      "token",
		}

		js, err = json.Marshal(shuffleBallotsRequest)
		require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

		resp, err = http.Post(proxyAddr1+shuffleBallotsEndpoint, contentType, bytes.NewBuffer(js))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		time.Sleep(10 * time.Second)

		t.Log("Get election info")
		create_info_js = fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		resp, err = http.Post(proxyAddr1+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = objmap["Status"].(string)
		t.Logf("Status of the election : " + electionStatus)

		// ###################################### REQUEST PUBLIC SHARES ############

		t.Log("request public shares")

		beginDecryptionRequest := types.BeginDecryptionRequest{
			ElectionID: electionID,
			UserID:     "adminId",
			Token:      "token",
		}

		js, err = json.Marshal(beginDecryptionRequest)
		require.NoError(t, err, "failed to set marshall types.SimpleElection : %v", err)

		resp, err = http.Post(proxyAddr1+beginDecryptionEndpoint, contentType, bytes.NewBuffer(js))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		time.Sleep(10 * time.Second)

		t.Log("Get election info")
		create_info_js = fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		resp, err = http.Post(proxyAddr1+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = objmap["Status"].(string)
		t.Logf("Status of the election : " + electionStatus)

		// ###################################### DECRYPT BALLOTS ##################

		t.Log("decrypt ballots")

		decryptBallotsRequest := types.CombineSharesRequest{
			ElectionID: electionID,
			UserID:     "adminId",
			Token:      "token",
		}

		js, err = json.Marshal(decryptBallotsRequest)
		require.NoError(t, err, "failed to set marshall types.CombineSharesRequest : %v", err)

		resp, err = http.Post(proxyAddr1+combineSharesEndpoint, contentType, bytes.NewBuffer(js))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		time.Sleep(10 * time.Second)

		t.Log("Get election info")
		create_info_js = fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		resp, err = http.Post(proxyAddr1+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = objmap["Status"].(string)
		t.Logf("Status of the election : " + electionStatus)

		// ###################################### GET ELECTION RESULT ##############

		t.Log("Get election result")

		getElectionResultRequest := types.GetElectionResultRequest{
			ElectionID: electionID,
			Token:      "token",
		}

		js, err = json.Marshal(getElectionResultRequest)
		require.NoError(t, err, "failed to set marshall types.GetElectionResultRequest : %v", err)

		resp, err = http.Post(proxyAddr1+getElectionResultEndpoint, contentType, bytes.NewBuffer(js))

		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)
		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("Response body: " + string(body))
		resp.Body.Close()

		time.Sleep(10 * time.Second)

		t.Log("Get election info")
		create_info_js = fmt.Sprintf(`{"ElectionID":"%v","Token":"token"}`, electionID)

		resp, err = http.Post(proxyAddr1+getElectionInfoEndpoint, contentType, bytes.NewBuffer([]byte(create_info_js)))
		require.NoError(t, err, "failed retrieve the decryption from the server: %v", err)

		require.Equal(t, resp.StatusCode, http.StatusOK, "unexpected status: %s", resp.Status)

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read the body of the response: %v", err)

		t.Log("response body:", string(body))
		resp.Body.Close()
		err = json.Unmarshal(body, &objmap)
		require.NoError(t, err, "failed to parsethe body of the response from js: %v", err)
		electionStatus = objmap["Status"].(string)
		t.Logf("Status of the election : " + electionStatus)
	}
}

// -----------------------------------------------------------------------------
// Utility functions
func marshallBallot_manual(voteStr string, pubkey kyber.Point, chunks int) (types.Ciphervote, error) {
	// chunk by default 1
	var ballot = make(types.Ciphervote, chunks)
	vote := strings.NewReader(voteStr)

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
