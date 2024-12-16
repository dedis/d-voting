package integration

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/d-voting/contracts/evoting/types"
	_ "go.dedis.ch/d-voting/services/dkg/pedersen/json"
	_ "go.dedis.ch/d-voting/services/shuffle/neff/json"
	delaPkg "go.dedis.ch/dela"
)

func TestIntegration(t *testing.T) {
	t.Run("4 nodes, 5 votes", getIntegrationTest(4, 5))

}

func IgnoreTestCrash(t *testing.T) {
	t.Run("5 nodes, 5 votes, 1 fail", getIntegrationTestCrash(5, 5, 1))
	//t.Run("5 nodes, 5 votes, 2 fails", getIntegrationTestCrash(5, 5, 2))
}

func BenchmarkIntegration(b *testing.B) {
	b.Run("10 nodes, 100 votes", getIntegrationBenchmark(10, 100))
}

func getIntegrationTest(numNodes, numVotes int) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		adminID := "first admin"

		// ##### SETUP ENV #####

		delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

		dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
		require.NoError(t, err)

		defer os.RemoveAll(dirPath)

		t.Logf("using temp dir %s", dirPath)

		// ##### CREATE NODES #####
		nodes := setupDVotingNodes(t, numNodes, dirPath)

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

		_, err = getForm(formFac, formID, nodes[0].GetOrdering())
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

		fmt.Println("Title of the form : " + form.Configuration.Title.En)
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

func getIntegrationTestCrash(numNodes, numVotes, failingNodes int) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		adminID := "first admin"

		// ##### SETUP ENV #####

		delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

		dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
		require.NoError(t, err)

		defer os.RemoveAll(dirPath)

		t.Logf("using temp dir %s", dirPath)

		// ##### CREATE NODES #####
		nodes := setupDVotingNodes(t, numNodes, dirPath)

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

		// If the number of failing nodes is greater
		// than the threshold, the shuffle will fail
		threshold := getThreshold(numNodes)
		fmt.Println("threshold: ", threshold)
		if failingNodes > threshold {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)

		err = waitForStatus(types.ShuffledBallots, formFac, formID, nodes,
			numNodes, 2*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### SUBMIT PUBLIC SHARES #####
		t.Logf("submitting public shares")

		_, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)
		err = actor.ComputePubshares()
		require.NoError(t, err)

		// Heisenbug: https://github.com/c4dt/d-voting/issues/90
		err = waitForStatus(types.PubSharesSubmitted, formFac, formID, nodes,
			numNodes, 6*time.Second*time.Duration(numNodes))
		require.NoError(t, err)

		// ##### DECRYPT BALLOTS #####
		t.Logf("decrypting")

		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		t.Logf("PubsharesUnit: %v", form.PubsharesUnits)
		require.NoError(t, err)
		// Heisenbug: https://github.com/c4dt/d-voting/issues/90
		err = decryptBallots(m, actor, form)
		require.NoError(t, err)

		err = waitForStatus(types.ResultAvailable, formFac, formID, nodes,
			numNodes, 1500*time.Millisecond*time.Duration(numVotes))
		require.NoError(t, err)

		t.Logf("get vote proof")
		form, err = getForm(formFac, formID, nodes[0].GetOrdering())
		require.NoError(t, err)

		fmt.Println("Title of the form : " + form.Configuration.Title.En)
		fmt.Println("ID of the form : " + string(form.FormID))
		fmt.Println("Status of the form : " + strconv.Itoa(int(form.Status)))
		fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))

		checkBallots(form.DecryptedBallots, castedVotes, t)

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

		delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

		dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
		require.NoError(b, err)

		defer os.RemoveAll(dirPath)

		// ##### CREATE NODES #####
		nodes := setupDVotingNodes(b, numNodes, dirPath)

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

		_, err = getForm(formFac, formID, nodes[0].GetOrdering())
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

		fmt.Println("Title of the form : " + form.Configuration.Title.En)
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
