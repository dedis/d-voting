package integration

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	delaPkg "go.dedis.ch/dela"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/txn"
	"golang.org/x/xerrors"
)

// Check the shuffled votes versus the casted votes and a few nodes
func BenchmarkIntegration_CustomVotesScenario(b *testing.B) {

	numNodes := 3
	numVotes := 3
	numChunksPerBallot := 3

	adminID := "I am an admin"

	// ##### SETUP ENV #####
	// make tests reproducible
	rand.Seed(1)

	delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

	dirPath, err := os.MkdirTemp(os.TempDir(), "d-voting-three-votes")
	require.NoError(b, err)

	defer os.RemoveAll(dirPath)

	b.Logf("using temp dir %s", dirPath)

	// ##### CREATE NODES #####
	nodes := setupDVotingNodes(b, numNodes, dirPath,nil)

	signer := createDVotingAccess(b, nodes, dirPath)

	m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*2)

	err = grantAccess(m, signer)
	require.NoError(b, err)

	for _, n := range nodes {
		err = grantAccess(m, n.GetShuffleSigner())
		require.NoError(b, err)
	}

	// ##### CREATE FORM #####
	formID, err := createFormNChunks(m, "Three votes form", adminID, numChunksPerBallot)
	require.NoError(b, err)

	time.Sleep(time.Millisecond * 1000)

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

	castedVotes, err := castVotesNChunks(m, actor, form, numVotes)
	require.NoError(b, err)

	// ##### CLOSE FORM #####
	err = closeForm(m, formID, adminID)
	require.NoError(b, err)

	time.Sleep(time.Millisecond * 1000)

	// ##### SHUFFLE BALLOTS ####

	//b.ResetTimer()

	b.Logf("initializing shuffle")
	sActor, err := initShuffle(nodes)
	require.NoError(b, err)

	b.Logf("shuffling")
	err = sActor.Shuffle(formID)
	require.NoError(b, err)

	//b.StopTimer()

	// ##### DECRYPT BALLOTS #####
	time.Sleep(time.Second * 1)

	b.Logf("decrypting")

	//b.ResetTimer()

	form, err = getForm(formFac, formID, nodes[0].GetOrdering())
	require.NoError(b, err)

	err = decryptBallots(m, actor, form)
	require.NoError(b, err)

	//b.StopTimer()

	time.Sleep(time.Second * 1)

	b.Logf("get vote proof")
	form, err = getForm(formFac, formID, nodes[0].GetOrdering())
	require.NoError(b, err)

	fmt.Println("Title of the form : " + form.Configuration.MainTitle)
	fmt.Println("ID of the form : " + string(form.FormID))
	fmt.Println("Status of the form : " + strconv.Itoa(int(form.Status)))
	fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))
	fmt.Println("Chunks per ballot : " + strconv.Itoa(form.ChunksPerBallot()))

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

	closeNodesBench(b, nodes)
}

func createFormNChunks(m txManager, title string, admin string, numChunks int) ([]byte, error) {

	defaultBallotContent := "text:" + encodeID("bb") + ":\n\n"
	textSize := 29*numChunks - len(defaultBallotContent)

	// Define the configuration :
	configuration := types.Configuration{
		MainTitle: title,
		Scaffold: []types.Subject{
			{
				ID:       "aa",
				Title:    "subject1",
				Order:    nil,
				Subjects: nil,
				Selects:  nil,
				Ranks:    []types.Rank{},
				Texts: []types.Text{{
					ID:        "bb",
					Title:     "Enter favorite snack",
					MaxN:      1,
					MinN:      0,
					MaxLength: uint(base64.StdEncoding.DecodedLen(textSize)),
					Regex:     "",
					Choices:   []string{"Your fav snack: "},
				}},
			},
		},
	}

	createForm := types.CreateForm{
		Configuration: configuration,
		AdminID:       admin,
	}

	data, err := createForm.Serialize(serdecontext)
	if err != nil {
		return nil, xerrors.Errorf("failed to serialize create form: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.FormArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCreateForm)},
	}

	txID, err := m.addAndWait(args...)
	if err != nil {
		return nil, xerrors.Errorf("failed to addAndWait: %v", err)
	}

	// Calculate formID from
	hash := sha256.New()
	hash.Write(txID)
	formID := hash.Sum(nil)

	return formID, nil
}

func castVotesNChunks(m txManager, actor dkg.Actor, form types.Form,
	numberOfVotes int) ([]types.Ballot, error) {

	ballotBuilder := strings.Builder{}

	ballotBuilder.Write([]byte("text:"))
	ballotBuilder.Write([]byte(encodeID("bb")))
	ballotBuilder.Write([]byte(":"))

	textSize := 29*form.ChunksPerBallot() - ballotBuilder.Len() - 3

	ballotBuilder.Write([]byte(strings.Repeat("=", textSize)))
	ballotBuilder.Write([]byte("\n\n"))

	vote := ballotBuilder.String()

	ballot, err := marshallBallot(strings.NewReader(vote), actor, form.ChunksPerBallot())
	if err != nil {
		return nil, xerrors.Errorf("failed to marshallBallot: %v", err)
	}

	votes := make([]types.Ballot, numberOfVotes)

	for i := 0; i < numberOfVotes; i++ {

		userID := "user " + strconv.Itoa(i)

		castVote := types.CastVote{
			FormID: form.FormID,
			UserID:     userID,
			Ballot:     ballot,
		}

		data, err := castVote.Serialize(serdecontext)
		if err != nil {
			return nil, xerrors.Errorf("failed to serialize castVote: %v", err)
		}

		args := []txn.Arg{
			{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
			{Key: evoting.FormArg, Value: data},
			{Key: evoting.CmdArg, Value: []byte(evoting.CmdCastVote)},
		}
		_, err = m.addAndWait(args...)
		if err != nil {
			return nil, xerrors.Errorf("failed to addAndWait: %v", err)
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

func closeNodesBench(b *testing.B, nodes []dVotingCosiDela) {
	wait := sync.WaitGroup{}
	wait.Add(len(nodes))

	for _, n := range nodes {
		go func(node dVotingNode) {
			defer wait.Done()
			node.GetOrdering().Close()
		}(n.(dVotingNode))
	}

	b.Log("stopping nodes...")

	done := make(chan struct{})

	go func() {
		select {
		case <-done:
		case <-time.After(time.Second * 30):
			b.Error("timeout while closing nodes")
		}
	}()

	wait.Wait()
	close(done)
}
