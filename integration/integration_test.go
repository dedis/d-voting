package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg"
	_ "github.com/dedis/d-voting/services/dkg/pedersen/json"
	"github.com/dedis/d-voting/services/shuffle"
	_ "github.com/dedis/d-voting/services/shuffle/neff/json"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	delaPkg "go.dedis.ch/dela"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/dela/serde/json"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

const addAndWaitErr = "failed to addAndWait: %v"

var serdecontext = json.NewContext()

// Check the shuffled votes versus the cast votes on a few nodes
func TestIntegration(t *testing.T) {
	t.Run("3 nodes, 3 votes", getIntegrationTest(3, 3))
	t.Run("10 nodes, 100 votes", getIntegrationTest(10, 100))
}

func TestBadVote(t *testing.T) {
	t.Run("5 nodes, 10 votes including 5 bad votes", getIntegrationTestBadVote(5, 10, 5))
}

func TestRevote(t *testing.T) {
	t.Run("5 nodes, 10 votes ", getIntegrationTestRevote(5, 10))
}

func TestCrash(t *testing.T) {
	t.Run("5 nodes, 5 votes, 1 fail", getIntegrationTestCrash(5, 5, 1))
	t.Run("5 nodes, 5 votes, 2 fails", getIntegrationTestCrash(5, 5, 2))
	t.Run("10 nodes, 5 votes, 3 fails", getIntegrationTestCrash(10, 5, 3))
}

func TestAddNodes(t *testing.T) {
	t.Run("5 nodes, 10 votes add 2 node during the process", getIntegrationTestAddNodes(5, 10, 2))
}

func BenchmarkIntegration(b *testing.B) {
	b.Run("10 nodes, 100 votes", getIntegrationBenchmark(10, 100))
}

func getIntegrationTest(numNodes, numVotes int) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		adminID := "first admin"

		// ##### SETUP ENV #####
		// make tests reproducible
		rand.Seed(1)

		delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

		dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
		require.NoError(t, err)

		defer os.RemoveAll(dirPath)

		t.Logf("using temp dir %s", dirPath)

		// ##### CREATE NODES #####
		nodes := setupDVotingNodes(t, numNodes, dirPath, nil)

		signer := createDVotingAccess(t, nodes, dirPath)

		m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*4)

		err = grantAccess(m, signer)
		require.NoError(t, err)

		for _, n := range nodes {
			err = grantAccess(m, n.GetShuffleSigner())
			require.NoError(t, err)
		}

		// ##### CREATE FORM #####
		formID, err := createForm(m, "Three votes form", adminID)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		// ##### SETUP DKG #####
		actor, err := initDkg(nodes, formID, m.m)
		require.NoError(t, err)

		// ##### OPEN FORM #####
		err = openForm(m, formID)
		require.NoError(t, err)

		formFac := types.NewFormFactory(types.CiphervoteFactory{}, nodes[0].GetRosterFac())

		t.Logf("start casting votes")
		form, err := getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		castedVotes, err := castVotesRandomly(m, actor, form, numVotes)
		require.NoError(t, err)

		fmt.Println("casted votes:", castedVotes)

		// ##### CLOSE FORM #####
		err = closeForm(m, formID, adminID)
		require.NoError(t, err)

		err = waitForStatus(types.Closed, formFac, formID, nodes, numNodes,
			5*time.Second)
		require.NoError(t, err)

		// ##### SHUFFLE BALLOTS #####
		t.Logf("initializing shuffle")
		sActor, err := initShuffle(nodes)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		t.Logf("shuffling")
		err = sActor.Shuffle(formID)
		require.NoError(t, err)

		err = waitForStatus(types.ShuffledBallots, formFac, formID, nodes,
			numNodes, 2*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### SUBMIT PUBLIC SHARES #####
		t.Logf("submitting public shares")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)
		err = actor.ComputePubshares()
		require.NoError(t, err)

		err = waitForStatus(types.PubSharesSubmitted, formFac, formID, nodes,
			numNodes, 6*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### DECRYPT BALLOTS #####
		t.Logf("decrypting")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		t.Logf("PubsharesUnit: %v", form.PubsharesUnits)
		require.NoError(t, err)
		err = decryptBallots(m, actor, form)
		require.NoError(t, err)

		err = waitForStatus(types.ResultAvailable, formFac, formID, nodes,
			numNodes, 1500*time.Millisecond*time.Duration(numVotes))
		require.NoError(t, err)

		t.Logf("get vote proof")
		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		fmt.Println("Title of the form : " + form.Configuration.MainTitle)
		fmt.Println("ID of the form : " + string(form.FormID))
		fmt.Println("Status of the form : " + strconv.Itoa(int(form.Status)))
		fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))

		require.Len(t, form.DecryptedBallots, len(castedVotes))

		for _, b := range form.DecryptedBallots {
			ok := false
			for i, casted := range castedVotes {
				if b.Equal(casted) {
					//remove the casted vote from the list
					castedVotes = append(castedVotes[:i], castedVotes[i+1:]...)
					ok = true
					break
				}
			}
			require.True(t, ok)
		}
		require.Empty(t, castedVotes)

		fmt.Println("closing nodes")

		err = closeNodes(nodes)
		require.NoError(t, err)

		fmt.Println("test done")
	}
}

func getIntegrationTestBadVote(numNodes, numVotes, numBadVotes int) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		adminID := "first admin"

		// ##### SETUP ENV #####
		// make tests reproducible
		rand.Seed(1)

		delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

		dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
		require.NoError(t, err)

		defer os.RemoveAll(dirPath)

		t.Logf("using temp dir %s", dirPath)

		// ##### CREATE NODES #####
		nodes := setupDVotingNodes(t, numNodes, dirPath, nil)

		signer := createDVotingAccess(t, nodes, dirPath)

		m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*4)

		err = grantAccess(m, signer)
		require.NoError(t, err)

		for _, n := range nodes {
			err = grantAccess(m, n.GetShuffleSigner())
			require.NoError(t, err)
		}

		// ##### CREATE FORM #####
		formID, err := createForm(m, "Three votes form", adminID)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		// ##### SETUP DKG #####
		actor, err := initDkg(nodes, formID, m.m)
		require.NoError(t, err)

		// ##### OPEN FORM #####
		err = openForm(m, formID)
		require.NoError(t, err)

		formFac := types.NewFormFactory(types.CiphervoteFactory{}, nodes[0].GetRosterFac())

		t.Logf("start casting votes")
		form, err := getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		//cast a vote with wrong answers: Should not be taken into account

		_, err = castVotesRandomly(m, actor, form, numVotes-numBadVotes)
		require.NoError(t, err)

		err = castBadVote(m, actor, form, numBadVotes)
		require.NoError(t, err)

		//castedVotes = append(castedVotes, badVotes...)

		//fmt.Println("casted votes:", castedVotes)

		// ##### CLOSE FORM #####
		err = closeForm(m, formID, adminID)
		require.NoError(t, err)

		err = waitForStatus(types.Closed, formFac, formID, nodes, numNodes,
			5*time.Second)
		require.NoError(t, err)

		// ##### SHUFFLE BALLOTS #####
		t.Logf("initializing shuffle")
		sActor, err := initShuffle(nodes)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		t.Logf("shuffling")
		err = sActor.Shuffle(formID)
		require.NoError(t, err)

		err = waitForStatus(types.ShuffledBallots, formFac, formID, nodes,
			numNodes, 2*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### SUBMIT PUBLIC SHARES #####
		t.Logf("submitting public shares")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)
		err = actor.ComputePubshares()
		require.NoError(t, err)

		err = waitForStatus(types.PubSharesSubmitted, formFac, formID, nodes,
			numNodes, 6*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### DECRYPT BALLOTS #####
		t.Logf("decrypting")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		t.Logf("PubsharesUnit: %v", form.PubsharesUnits)
		require.NoError(t, err)
		err = decryptBallots(m, actor, form)
		require.NoError(t, err)

		err = waitForStatus(types.ResultAvailable, formFac, formID, nodes,
			numNodes, 1500*time.Millisecond*time.Duration(numVotes))
		require.NoError(t, err)

		t.Logf("get vote proof")
		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		fmt.Println("Title of the form : " + form.Configuration.MainTitle)
		fmt.Println("ID of the form : " + string(form.FormID))
		fmt.Println("Status of the form : " + strconv.Itoa(int(form.Status)))
		fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))

		//require.Len(t, form.DecryptedBallots, numVotes-numBadVotes)
		//should contains numBadVotes empty ballots
		count := 0
		for _, ballot := range form.DecryptedBallots {
			if ballotIsNull(ballot) {
				count++
			}
			//fmt.Println(fmt.Sprintf("%#v", ballot))
		}
		fmt.Println(form.DecryptedBallots)

		require.Equal(t, numBadVotes, count)

		fmt.Println("closing nodes")

		err = closeNodes(nodes)
		require.NoError(t, err)

		fmt.Println("test done")
	}
}

func getIntegrationTestRevote(numNodes, numVotes int) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		adminID := "first admin"

		// ##### SETUP ENV #####
		// make tests reproducible
		rand.Seed(1)

		delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

		dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
		require.NoError(t, err)

		defer os.RemoveAll(dirPath)

		t.Logf("using temp dir %s", dirPath)

		// ##### CREATE NODES #####
		nodes := setupDVotingNodes(t, numNodes, dirPath, nil)

		signer := createDVotingAccess(t, nodes, dirPath)

		m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*4)

		err = grantAccess(m, signer)
		require.NoError(t, err)

		for _, n := range nodes {
			err = grantAccess(m, n.GetShuffleSigner())
			require.NoError(t, err)
		}

		// ##### CREATE FORM #####
		formID, err := createForm(m, "Three votes form", adminID)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		// ##### SETUP DKG #####
		actor, err := initDkg(nodes, formID, m.m)
		require.NoError(t, err)

		// ##### OPEN FORM #####
		err = openForm(m, formID)
		require.NoError(t, err)

		formFac := types.NewFormFactory(types.CiphervoteFactory{}, nodes[0].GetRosterFac())

		t.Logf("start casting votes")
		form, err := getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		_, err = castVotesRandomly(m, actor, form, numVotes)
		require.NoError(t, err)

		castedVotes, err := castVotesRandomly(m, actor, form, numVotes)
		require.NoError(t, err)

		fmt.Println("casted votes:", castedVotes)

		// ##### CLOSE FORM #####
		err = closeForm(m, formID, adminID)
		require.NoError(t, err)

		err = waitForStatus(types.Closed, formFac, formID, nodes, numNodes,
			5*time.Second)
		require.NoError(t, err)

		// ##### SHUFFLE BALLOTS #####
		t.Logf("initializing shuffle")
		sActor, err := initShuffle(nodes)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		t.Logf("shuffling")
		err = sActor.Shuffle(formID)
		require.NoError(t, err)

		err = waitForStatus(types.ShuffledBallots, formFac, formID, nodes,
			numNodes, 2*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### SUBMIT PUBLIC SHARES #####
		t.Logf("submitting public shares")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)
		err = actor.ComputePubshares()
		require.NoError(t, err)

		err = waitForStatus(types.PubSharesSubmitted, formFac, formID, nodes,
			numNodes, 6*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### DECRYPT BALLOTS #####
		t.Logf("decrypting")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		t.Logf("PubsharesUnit: %v", form.PubsharesUnits)
		require.NoError(t, err)
		err = decryptBallots(m, actor, form)
		require.NoError(t, err)

		err = waitForStatus(types.ResultAvailable, formFac, formID, nodes,
			numNodes, 1500*time.Millisecond*time.Duration(numVotes))
		require.NoError(t, err)

		t.Logf("get vote proof")
		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		fmt.Println("Title of the form : " + form.Configuration.MainTitle)
		fmt.Println("ID of the form : " + string(form.FormID))
		fmt.Println("Status of the form : " + strconv.Itoa(int(form.Status)))
		fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))

		require.Len(t, form.DecryptedBallots, len(castedVotes))

		for _, b := range form.DecryptedBallots {
			ok := false
			for i, casted := range castedVotes {
				if b.Equal(casted) {
					ok = true
					//remove the casted vote from the list
					castedVotes = append(castedVotes[:i], castedVotes[i+1:]...)
					break
				}
			}
			require.True(t, ok)
		}

		fmt.Println("closing nodes")

		err = closeNodes(nodes)
		require.NoError(t, err)

		fmt.Println("test done")
	}
}

func getIntegrationTestCrash(numNodes, numVotes, failingNodes int) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		adminID := "first admin"

		// ##### SETUP ENV #####
		// make tests reproducible
		rand.Seed(1)

		delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

		dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
		require.NoError(t, err)

		defer os.RemoveAll(dirPath)

		t.Logf("using temp dir %s", dirPath)

		// ##### CREATE NODES #####
		nodes := setupDVotingNodes(t, numNodes, dirPath, nil)

		signer := createDVotingAccess(t, nodes, dirPath)

		m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*4)

		err = grantAccess(m, signer)
		require.NoError(t, err)

		for _, n := range nodes {
			err = grantAccess(m, n.GetShuffleSigner())
			require.NoError(t, err)
		}

		// ##### CREATE FORM #####
		formID, err := createForm(m, "Three votes form", adminID)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		// ##### SETUP DKG #####
		actor, err := initDkg(nodes, formID, m.m)
		require.NoError(t, err)

		// ##### OPEN FORM #####
		err = openForm(m, formID)
		require.NoError(t, err)

		formFac := types.NewFormFactory(types.CiphervoteFactory{}, nodes[0].GetRosterFac())

		t.Logf("start casting votes")
		form, err := getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		// crashNodeList nodes crashes during the process

		var crashNodeList []dVotingCosiDela
		for i := 0; i < failingNodes; i++ {
			crashID := rand.Intn(numNodes - i)
			crashNode := nodes[crashID]
			nodes = append(nodes[:crashID], nodes[crashID+1:]...)
			crashNodeList = append(crashNodeList, crashNode)
		}
		err = closeNodes(crashNodeList)
		require.NoError(t, err)

		castedVotes, err := castVotesRandomly(m, actor, form, numVotes)
		require.NoError(t, err)

		fmt.Println("casted votes:", castedVotes)

		// ##### CLOSE FORM #####
		err = closeForm(m, formID, adminID)
		require.NoError(t, err)

		err = waitForStatus(types.Closed, formFac, formID, nodes, numNodes,
			5*time.Second)
		require.NoError(t, err)

		// ##### SHUFFLE BALLOTS #####
		t.Logf("initializing shuffle")
		sActor, err := initShuffle(nodes)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		t.Logf("shuffling")
		err = sActor.Shuffle(formID)

		// If the number of failing nodes is greater than the threshold, the shuffle will fail
		fmt.Println("threshold: ", numNodes/3)
		if failingNodes > numNodes/3 {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)

		err = waitForStatus(types.ShuffledBallots, formFac, formID, nodes,
			numNodes, 2*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### SUBMIT PUBLIC SHARES #####
		t.Logf("submitting public shares")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)
		err = actor.ComputePubshares()
		require.NoError(t, err)

		err = waitForStatus(types.PubSharesSubmitted, formFac, formID, nodes,
			numNodes, 6*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### DECRYPT BALLOTS #####
		t.Logf("decrypting")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		t.Logf("PubsharesUnit: %v", form.PubsharesUnits)
		require.NoError(t, err)
		err = decryptBallots(m, actor, form)
		require.NoError(t, err)

		err = waitForStatus(types.ResultAvailable, formFac, formID, nodes,
			numNodes, 1500*time.Millisecond*time.Duration(numVotes))
		require.NoError(t, err)

		t.Logf("get vote proof")
		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		fmt.Println("Title of the form : " + form.Configuration.MainTitle)
		fmt.Println("ID of the form : " + string(form.FormID))
		fmt.Println("Status of the form : " + strconv.Itoa(int(form.Status)))
		fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))

		require.Len(t, form.DecryptedBallots, len(castedVotes))

		for _, b := range form.DecryptedBallots {
			ok := false
			for i, casted := range castedVotes {
				if b.Equal(casted) {
					//remove the casted vote from the list
					castedVotes = append(castedVotes[:i], castedVotes[i+1:]...)
					ok = true
					break
				}
			}
			require.True(t, ok)
		}
		require.Empty(t, castedVotes)

		fmt.Println("closing nodes")

		err = closeNodes(nodes)
		require.NoError(t, err)

		fmt.Println("test done")
	}
}

func getIntegrationTestAddNodes(numNodes, numVotes, numNewNodes int) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		adminID := "first admin"

		// ##### SETUP ENV #####
		// make tests reproducible
		rand.Seed(1)

		delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

		dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
		require.NoError(t, err)

		defer os.RemoveAll(dirPath)

		t.Logf("using temp dir %s", dirPath)

		// ##### CREATE NODES #####
		nodes := setupDVotingNodes(t, numNodes, dirPath, nil)

		signer := createDVotingAccess(t, nodes, dirPath)

		m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*4)

		err = grantAccess(m, signer)
		require.NoError(t, err)

		for _, n := range nodes {
			err = grantAccess(m, n.GetShuffleSigner())
			require.NoError(t, err)
		}

		// ##### CREATE FORM #####
		formID, err := createForm(m, "Three votes form", adminID)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		// ##### SETUP DKG #####
		actor, err := initDkg(nodes, formID, m.m)
		require.NoError(t, err)

		// ##### OPEN FORM #####
		err = openForm(m, formID)
		require.NoError(t, err)

		formFac := types.NewFormFactory(types.CiphervoteFactory{}, nodes[0].GetRosterFac())

		t.Logf("start casting votes")
		form, err := getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		newNodes := setupDVotingNodes(t, numNewNodes, dirPath, nodes[0])
		nodes = append(nodes, newNodes...)

		castedVotes, err := castVotesRandomly(m, actor, form, numVotes)
		require.NoError(t, err)

		fmt.Println("casted votes:", castedVotes)

		// ##### CLOSE FORM #####
		err = closeForm(m, formID, adminID)
		require.NoError(t, err)

		err = waitForStatus(types.Closed, formFac, formID, nodes, numNodes,
			5*time.Second)
		require.NoError(t, err)

		// ##### SHUFFLE BALLOTS #####
		t.Logf("initializing shuffle")
		sActor, err := initShuffle(nodes)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		t.Logf("shuffling")
		err = sActor.Shuffle(formID)
		require.NoError(t, err)

		err = waitForStatus(types.ShuffledBallots, formFac, formID, nodes,
			numNodes, 2*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### SUBMIT PUBLIC SHARES #####
		t.Logf("submitting public shares")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)
		err = actor.ComputePubshares()
		require.NoError(t, err)

		err = waitForStatus(types.PubSharesSubmitted, formFac, formID, nodes,
			numNodes, 6*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### DECRYPT BALLOTS #####
		t.Logf("decrypting")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		t.Logf("PubsharesUnit: %v", form.PubsharesUnits)
		require.NoError(t, err)
		err = decryptBallots(m, actor, form)
		require.NoError(t, err)

		err = waitForStatus(types.ResultAvailable, formFac, formID, nodes,
			numNodes, 1500*time.Millisecond*time.Duration(numVotes))
		require.NoError(t, err)

		t.Logf("get vote proof")
		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		fmt.Println("Title of the form : " + form.Configuration.MainTitle)
		fmt.Println("ID of the form : " + string(form.FormID))
		fmt.Println("Status of the form : " + strconv.Itoa(int(form.Status)))
		fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))

		require.Len(t, form.DecryptedBallots, len(castedVotes))

		for _, b := range form.DecryptedBallots {
			ok := false
			for i, casted := range castedVotes {
				if b.Equal(casted) {
					//remove the casted vote from the list
					castedVotes = append(castedVotes[:i], castedVotes[i+1:]...)
					ok = true
					break
				}
			}
			require.True(t, ok)
		}
		require.Empty(t, castedVotes)

		fmt.Println("closing nodes")

		err = closeNodes(nodes)
		require.NoError(t, err)

		fmt.Println("test done")
	}
}

func getIntegrationBenchmark(numNodes, numVotes int) func(*testing.B) {
	return func(b *testing.B) {

		adminID := "first admin"

		// ##### SETUP ENV #####
		// make tests reproducible
		rand.Seed(1)

		delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

		dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
		require.NoError(b, err)

		defer os.RemoveAll(dirPath)

		// ##### CREATE NODES #####
		nodes := setupDVotingNodes(b, numNodes, dirPath,nil)

		signer := createDVotingAccess(b, nodes, dirPath)

		m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*4)

		err = grantAccess(m, signer)
		require.NoError(b, err)

		for _, n := range nodes {
			err = grantAccess(m, n.GetShuffleSigner())
			require.NoError(b, err)
		}

		// ##### CREATE FORM #####
		formID, err := createForm(m, "Three votes form", adminID)
		require.NoError(b, err)

		time.Sleep(time.Second * 1)

		// ##### SETUP DKG #####
		actor, err := initDkg(nodes, formID, m.m)
		require.NoError(b, err)

		// ##### OPEN FORM #####
		err = openForm(m, formID)
		require.NoError(b, err)

		formFac := types.NewFormFactory(types.CiphervoteFactory{}, nodes[0].GetRosterFac())

		b.Logf("start casting votes")
		form, err := getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(b, err)

		castedVotes, err := castVotesRandomly(m, actor, form, numVotes)
		require.NoError(b, err)

		fmt.Println("casted votes:", castedVotes)

		// ##### CLOSE FORM #####
		err = closeForm(m, formID, adminID)
		require.NoError(b, err)

		err = waitForStatus(types.Closed, formFac, formID, nodes, numNodes,
			5*time.Second)
		require.NoError(b, err)

		// ##### SHUFFLE BALLOTS #####
		sActor, err := initShuffle(nodes)
		require.NoError(b, err)

		time.Sleep(time.Second * 1)

		err = sActor.Shuffle(formID)
		require.NoError(b, err)

		err = waitForStatus(types.ShuffledBallots, formFac, formID, nodes,
			numNodes, 2*time.Second*time.Duration(numNodes))
		require.NoError(b, err)

		// ##### SUBMIT PUBLIC SHARES #####

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(b, err)
		err = actor.ComputePubshares()
		require.NoError(b, err)

		err = waitForStatus(types.PubSharesSubmitted, formFac, formID, nodes,
			numNodes, 6*time.Second*time.Duration(numNodes))
		require.NoError(b, err)

		// ##### DECRYPT BALLOTS #####
		b.Logf("decrypting")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		b.Logf("PubsharesUnit: %v", form.PubsharesUnits)
		require.NoError(b, err)
		err = decryptBallots(m, actor, form)
		require.NoError(b, err)

		err = waitForStatus(types.ResultAvailable, formFac, formID, nodes,
			numNodes, 1500*time.Millisecond*time.Duration(numVotes))
		require.NoError(b, err)

		b.Logf("get vote proof")
		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(b, err)

		fmt.Println("Title of the form : " + form.Configuration.MainTitle)
		fmt.Println("ID of the form : " + string(form.FormID))
		fmt.Println("Status of the form : " + strconv.Itoa(int(form.Status)))
		fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))

		require.Len(b, form.DecryptedBallots, len(castedVotes))

		for _, ballot := range form.DecryptedBallots {
			ok := false
			for _, casted := range castedVotes {
				if ballot.Equal(casted) {
					ok = true
					break
				}
			}
			require.True(b, ok)
		}

		fmt.Println("closing nodes")

		err = closeNodes(nodes)
		require.NoError(b, err)

		fmt.Println("test done")
	}
}

// -----------------------------------------------------------------------------
// Utility functions

func ballotIsNull(ballot types.Ballot) bool {
	return ballot.SelectResultIDs == nil && ballot.SelectResult == nil &&
		ballot.RankResultIDs == nil && ballot.RankResult == nil &&
		ballot.TextResultIDs == nil && ballot.TextResult == nil
}

func newTxManager(signer crypto.Signer, firstNode dVotingCosiDela,
	timeout time.Duration, retry int) txManager {

	client := client{
		srvc: firstNode.GetOrdering(),
		mgr:  firstNode.GetValidationSrv(),
	}

	return txManager{
		m:     signed.NewManager(signer, client),
		n:     firstNode,
		t:     timeout,
		retry: retry,
	}
}

type txManager struct {
	m     txn.Manager
	n     dVotingCosiDela
	t     time.Duration
	retry int
}

func (m txManager) addAndWait(args ...txn.Arg) ([]byte, error) {
	for i := 0; i < m.retry; i++ {
		sentTxn, err := m.m.Make(args...)
		if err != nil {
			return nil, xerrors.Errorf("failed to Make: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), m.t)
		defer cancel()

		events := m.n.GetOrdering().Watch(ctx)

		err = m.n.GetPool().Add(sentTxn)
		if err != nil {
			return nil, xerrors.Errorf("failed to Add: %v", err)
		}

		sentTxnID := sentTxn.GetID()

		accepted := isAccepted(events, sentTxnID)
		if accepted {
			return sentTxnID, nil
		}

		err = m.m.Sync()
		if err != nil {
			return nil, xerrors.Errorf("failed to sync: %v", err)
		}

		cancel()
	}

	return nil, xerrors.Errorf("transaction not included after timeout: %v", args)
}

// isAccepted returns true if the transaction was included then accepted
func isAccepted(events <-chan ordering.Event, txID []byte) bool {
	for event := range events {
		for _, result := range event.Transactions {
			fetchedTxnID := result.GetTransaction().GetID()

			if bytes.Equal(txID, fetchedTxnID) {
				accepted, _ := event.Transactions[0].GetStatus()

				return accepted
			}
		}
	}

	return false
}

func grantAccess(m txManager, signer crypto.Signer) error {
	pubKeyBuf, err := signer.GetPublicKey().MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to GetPublicKey: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte("go.dedis.ch/dela.Access")},
		{Key: "access:grant_id", Value: []byte(hex.EncodeToString(evotingAccessKey[:]))},
		{Key: "access:grant_contract", Value: []byte("go.dedis.ch/dela.Evoting")},
		{Key: "access:grant_command", Value: []byte("all")},
		{Key: "access:identity", Value: []byte(base64.StdEncoding.EncodeToString(pubKeyBuf))},
		{Key: "access:command", Value: []byte("GRANT")},
	}
	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf("failed to grantAccess: %v", err)
	}

	return nil
}

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

func initShuffle(nodes []dVotingCosiDela) (shuffle.Actor, error) {
	var sActor shuffle.Actor

	for _, node := range nodes {
		client := client{
			srvc: node.GetOrdering(),
			mgr:  node.GetValidationSrv(),
		}

		var err error
		shuffler := node.GetShuffle()

		sActor, err = shuffler.Listen(signed.NewManager(node.GetShuffleSigner(), client))
		if err != nil {
			return nil, xerrors.Errorf("failed to init Shuffle: %v", err)
		}
	}

	return sActor, nil
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

func closeNodes(nodes []dVotingCosiDela) error {
	wait := sync.WaitGroup{}
	wait.Add(len(nodes))

	for _, n := range nodes {
		go func(node dVotingNode) {
			defer wait.Done()
			node.GetOrdering().Close()
		}(n.(dVotingNode))
	}

	done := make(chan struct{})

	go func() {
		wait.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(time.Second * 30):
		return xerrors.New("failed to close: timeout")
	}
}

func encodeID(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
}

// waitForStatus polls the nodes until they all updated to the expected status
// for the given form. An error is raised if the timeout expires.
func waitForStatus(status types.Status, formFac types.FormFactory,
	formID []byte, nodes []dVotingCosiDela, numNodes int, timeOut time.Duration) error {

	expiration := time.Now().Add(timeOut)

	isOK := func() (bool, error) {
		for _, node := range nodes {
			form, err := getForm(formFac, formID, node.GetOrdering())
			if err != nil {
				return false, xerrors.Errorf("failed to get form: %v", err)
			}

			if form.Status != status {
				return false, nil
			}
		}

		return true, nil
	}

	for {
		if time.Now().After(expiration) {
			return xerrors.New("status check expired")
		}

		ok, err := isOK()
		if err != nil {
			return xerrors.Errorf("failed to check status: %v", err)
		}

		if ok {
			return nil
		}

		time.Sleep(time.Millisecond * 100)
	}
}
