package integration

import (
	"fmt"
	"os"
	"strconv"

	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
	_ "github.com/dedis/d-voting/services/dkg/pedersen/json"
	_ "github.com/dedis/d-voting/services/shuffle/neff/json"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	delaPkg "go.dedis.ch/dela"
)

func TestBadVote(t *testing.T) {
	t.Skip("Bad votes don't work for the moment")
	t.Run("5 nodes, 10 votes including 5 bad votes", getIntegrationTestBadVote(5, 10, 5))
}

func TestRevote(t *testing.T) {
	t.Skip("Doesn't work in dedis/d-voting, neither")
	t.Run("5 nodes, 10 votes ", getIntegrationTestRevote(5, 10, 10))
}

func getIntegrationTestBadVote(numNodes, numVotes, numBadVotes int) func(*testing.T) {
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

		// cast a vote with wrong answers: Should not be taken into account

		_, err = castVotesRandomly(m, actor, form, numVotes-numBadVotes)
		require.NoError(t, err)

		err = castBadVote(m, actor, form, numBadVotes)
		require.NoError(t, err)

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

		// should contains numBadVotes empty ballots
		count := 0
		for _, ballot := range form.DecryptedBallots {
			if ballotIsNull(ballot) {
				count++
			}
		}
		fmt.Println(form.DecryptedBallots)

		require.Equal(t, numBadVotes, count)

		fmt.Println("closing nodes")

		err = closeNodes(nodes)
		require.NoError(t, err)

		fmt.Println("test done")
	}
}

func getIntegrationTestRevote(numNodes, numVotes, numRevotes int) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		require.LessOrEqual(t, numRevotes, numVotes)

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

		_, err = castVotesRandomly(m, actor, form, numVotes)
		require.NoError(t, err)

		castedVotes, err := castVotesRandomly(m, actor, form, numRevotes)
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

		checkBallots(form.DecryptedBallots, castedVotes, t)

		fmt.Println("closing nodes")

		err = closeNodes(nodes)
		require.NoError(t, err)

		fmt.Println("test done")
	}
}
