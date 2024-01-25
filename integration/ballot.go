package integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"sync/atomic"

	"github.com/c4dt/d-voting/contracts/evoting"
	"github.com/c4dt/d-voting/contracts/evoting/controller"
	"github.com/c4dt/d-voting/contracts/evoting/types"
	"github.com/c4dt/d-voting/internal/testing/fake"
	"github.com/c4dt/d-voting/proxy/txnmanager"
	ptypes "github.com/c4dt/d-voting/proxy/types"
	"github.com/c4dt/d-voting/services/dkg"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/xerrors"
)

const addAndWaitErr = "failed to addAndWait: %v"

var suite = suites.MustFind("Ed25519")

// ballotIsNull checks if a ballot is empty i.e. if all his fields are empty
func ballotIsNull(ballot types.Ballot) bool {
	return ballot.SelectResultIDs == nil && ballot.SelectResult == nil &&
		ballot.RankResultIDs == nil && ballot.RankResult == nil &&
		ballot.TextResultIDs == nil && ballot.TextResult == nil
}

// castVotesRandomly chooses numberOfVotes predefined ballots randomly
// and cast them
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

// castBadVote casts a vote with the good format but invalid content
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
	}

	return nil
}

// marshallBallot marshals a ballot and encrypts it
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

		pair := types.EGPair{
			K: K,
			C: C,
		}

		ballot[i] = pair

	}

	return ballot, nil
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

// encryptManual encrypts a ballot manually
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

// encodeBallotID encodes the ballotID
func encodeBallotID(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
}

// marshallBallotManual marshall a ballot and encrypt it manually
func marshallBallotManual(voteStr string, pubkey kyber.Point, chunks int) (ptypes.CiphervoteJSON, error) {

	ballot := make(ptypes.CiphervoteJSON, chunks)
	vote := strings.NewReader(voteStr)

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

// checkBallots checks that the decrypted ballots are correct
// and match the casted votes
func checkBallots(decryptedBallots, castedVotes []types.Ballot, t *testing.T) {
	require.Len(t, decryptedBallots, len(castedVotes))

	for _, b := range decryptedBallots {
		ok := false
		for i, casted := range castedVotes {
			if b.Equal(casted) {
				ok = true
				// remove the casted vote from the list
				castedVotes = append(castedVotes[:i], castedVotes[i+1:]...)
				break
			}
		}
		require.True(t, ok)
	}
	require.Empty(t, castedVotes)

}

// castVotesLoad casts vote for the load test
func castVotesLoad(numVotesPerSec, numSec, BallotSize, chunksPerBallot int, formID, contentType string, proxyArray []string, pubKey kyber.Point, secret kyber.Scalar, t *testing.T) []types.Ballot {

	t.Log("cast ballots")

	// make List of identical valid ballots
	b1 := string("select:" + encodeBallotID("bb") + ":0,0,1,0\n" + "text:" + encodeBallotID("ee") + ":eWVz\n\n") //encoding of "yes"

	numVotes := numVotesPerSec * numSec

	// create all the ballots
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

	// atomic counter
	var includedVoteCount uint64

	for i := 0; i < numSec; i++ {
		// send the votes asynchrounously and wait for the response

		for j := 0; j < numVotesPerSec; j++ {
			idx := i*numVotesPerSec + j
			randomproxy := proxyArray[rand.Intn(proxyCount)]
			castVoteRequest := ptypes.CastVoteRequest{
				UserID: "user" + strconv.Itoa(idx),
				Ballot: ballot,
			}
			// cast asynchrounously and increment includedVoteCount
			// if the cast was succesfull
			go func() {
				accepted := cast(idx, castVoteRequest, contentType, randomproxy, formID, secret, false, t)
				if accepted {
					atomic.AddUint64(&includedVoteCount, 1)
				}
			}()
		}
		t.Logf("casted votes %d", (i+1)*numVotesPerSec)
		time.Sleep(time.Second)

	}

	// wait until includedVoteCount == numVotes
	// i.e. that every votes has been included
	for {
		if atomic.LoadUint64(&includedVoteCount) == uint64(numVotes) {
			break
		}

		t.Logf("Waiting... included votes %d", atomic.LoadUint64(&includedVoteCount))
		infos := getFormInfo(proxyArray[0], formID, t)
		// check that our counter is synchronized with the blockchain
		t.Logf("Voters count: %v", len(infos.Voters))
		// check every 10 seconds
		time.Sleep(time.Second * 10)
	}

	return votesfrontend

}

// cast casts a vote for the load test
func cast(idx int, castVoteRequest ptypes.CastVoteRequest, contentType, randomproxy, formID string, secret kyber.Scalar, isRetry bool, t *testing.T) (accepted bool) {

	t.Logf("cast ballot to proxy %v", randomproxy)

	signed, err := createSignedRequest(secret, castVoteRequest)
	require.NoError(t, err)

	resp, err := http.Post(randomproxy+controller.FormPathSlash+formID+"/vote", contentType, bytes.NewBuffer(signed))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var infos txnmanager.TransactionClientInfo
	err = json.Unmarshal(body, &infos)
	require.NoError(t, err)

	ok, err := pollTxnInclusion(30, 2*time.Second, randomproxy, infos.Token, t)
	// if the transaction was not included after 2s, we retry once
	if !ok && !isRetry {
		t.Logf("retrying vote %d", idx)
		return cast(idx, castVoteRequest, contentType, randomproxy, formID, secret, true, t)

	}

	require.NoError(t, err)
	return ok
}

// castVotesScenario casts votes for the scenario test
func castVotesScenario(numVotes, BallotSize, chunksPerBallot int, formID, contentType string, proxyArray []string, pubKey kyber.Point, secret kyber.Scalar, t *testing.T) []types.Ballot {
	// make List of ballots
	b1 := string("select:" + encodeBallotID("bb") + ":0,0,1,0\n" + "text:" + encodeBallotID("ee") + ":eWVz\n\n") //encoding of "yes"

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

	for i := 0; i < numVotes; i++ {

		ballot, err := marshallBallotManual(ballotList[i], pubKey, chunksPerBallot)
		require.NoError(t, err)

		castVoteRequest := ptypes.CastVoteRequest{
			UserID: "user" + strconv.Itoa(i+1),
			Ballot: ballot,
		}

		randomproxy := proxyArray[rand.Intn(len(proxyArray))]
		t.Logf("cast ballot to proxy %v", randomproxy)

		signed, err := createSignedRequest(secret, castVoteRequest)
		require.NoError(t, err)

		resp, err := http.Post(randomproxy+"/evoting/forms/"+formID+"/vote", contentType, bytes.NewBuffer(signed))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", resp.Status)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var infos txnmanager.TransactionClientInfo
		err = json.Unmarshal(body, &infos)
		require.NoError(t, err)

		// send the votes 1 by 1 and wait for it to be included
		// to send the next one
		ok, err := pollTxnInclusion(60, 1*time.Second, randomproxy, infos.Token, t)
		require.NoError(t, err)
		require.True(t, ok)

		resp.Body.Close()

	}

	return votesfrontend
}
