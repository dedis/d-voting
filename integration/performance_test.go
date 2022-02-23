package integration

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
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

	dirPath, err := ioutil.TempDir(os.TempDir(), "d-voting-three-votes")
	require.NoError(b, err)

	defer os.RemoveAll(dirPath)

	b.Logf("using temp dir %s", dirPath)

	// ##### CREATE NODES #####
	nodes := setupDVotingNodes(b, numNodes, dirPath)

	signer := createDVotingAccess(b, nodes, dirPath)

	m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*2)

	err = grantAccess(m, signer)
	require.NoError(b, err)

	for _, n := range nodes {
		err = grantAccess(m, n.GetShuffleSigner())
		require.NoError(b, err)
	}

	// ##### CREATE ELECTION #####
	electionID, err := createElectionNChunks(m, "Three votes election", adminID, numChunksPerBallot)
	require.NoError(b, err)

	time.Sleep(time.Millisecond * 1000)

	// ##### SETUP DKG #####
	actor, err := initDkg(nodes, electionID, m.m)
	require.NoError(b, err)

	// ##### OPEN ELECTION #####
	err = openElection(m, electionID)
	require.NoError(b, err)

	electionFac := types.NewElectionFactory(types.CiphervoteFactory{}, nodes[0].GetRosterFac())

	b.Logf("start casting votes")
	election, err := getElection(electionFac, electionID, nodes[0].GetOrdering())
	require.NoError(b, err)

	castedVotes, err := castVotesNChunks(m, actor, election, numVotes)
	require.NoError(b, err)

	// ##### CLOSE ELECTION #####
	err = closeElection(m, electionID, adminID)
	require.NoError(b, err)

	time.Sleep(time.Millisecond * 1000)

	// ##### SHUFFLE BALLOTS ####

	//b.ResetTimer()

	b.Logf("initializing shuffle")
	sActor, err := initShuffle(nodes)
	require.NoError(b, err)

	b.Logf("shuffling")
	err = sActor.Shuffle(electionID)
	require.NoError(b, err)

	//b.StopTimer()

	// ##### DECRYPT BALLOTS #####
	time.Sleep(time.Second * 1)

	b.Logf("decrypting")

	//b.ResetTimer()

	election, err = getElection(electionFac, electionID, nodes[0].GetOrdering())
	require.NoError(b, err)

	err = decryptBallots(m, actor, election)
	require.NoError(b, err)

	//b.StopTimer()

	time.Sleep(time.Second * 1)

	b.Logf("get vote proof")
	election, err = getElection(electionFac, electionID, nodes[0].GetOrdering())
	require.NoError(b, err)

	fmt.Println("Title of the election : " + election.Configuration.MainTitle)
	fmt.Println("ID of the election : " + string(election.ElectionID))
	fmt.Println("Status of the election : " + strconv.Itoa(int(election.Status)))
	fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))
	fmt.Println("Chunks per ballot : " + strconv.Itoa(election.ChunksPerBallot()))

	require.Len(b, election.DecryptedBallots, len(castedVotes))

	for _, ballot := range election.DecryptedBallots {
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

func createElectionNChunks(m txManager, title string, admin string, numChunks int) ([]byte, error) {

	defaultBallotContent := "text:" + encodeID("bb") + ":\n\n"
	textSize := 29*numChunks - len(defaultBallotContent)

	// Define the configuration :
	configuration := types.Configuration{
		MainTitle: title,
		Scaffold: []types.Subject{
			{
				ID:       encodeID("aa"),
				Title:    "subject1",
				Order:    nil,
				Subjects: nil,
				Selects:  nil,
				Ranks:    []types.Rank{},
				Texts: []types.Text{{
					ID:        encodeID("bb"),
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

	createElection := types.CreateElection{
		Configuration: configuration,
		AdminID:       admin,
	}

	data, err := createElection.Serialize(serdecontext)
	if err != nil {
		return nil, xerrors.Errorf("failed to serialize create election: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.ElectionArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCreateElection)},
	}

	txID, err := m.addAndWait(args...)
	if err != nil {
		return nil, xerrors.Errorf("failed to addAndWait: %v", err)
	}

	// Calculate electionID from
	hash := sha256.New()
	hash.Write(txID)
	electionID := hash.Sum(nil)

	return electionID, nil
}

func castVotesNChunks(m txManager, actor dkg.Actor, election types.Election,
	numberOfVotes int) ([]types.Ballot, error) {

	ballotBuilder := strings.Builder{}

	ballotBuilder.Write([]byte("text:"))
	ballotBuilder.Write([]byte(encodeID("bb")))
	ballotBuilder.Write([]byte(":"))

	textSize := 29*election.ChunksPerBallot() - ballotBuilder.Len() - 3

	ballotBuilder.Write([]byte(strings.Repeat("=", textSize)))
	ballotBuilder.Write([]byte("\n\n"))

	vote := ballotBuilder.String()

	ballot, err := marshallBallot(strings.NewReader(vote), actor, election.ChunksPerBallot())
	if err != nil {
		return nil, xerrors.Errorf("failed to marshallBallot: %v", err)
	}

	votes := make([]types.Ballot, numberOfVotes)

	for i := 0; i < numberOfVotes; i++ {

		userID := "user " + strconv.Itoa(i)

		castVote := types.CastVote{
			ElectionID: election.ElectionID,
			UserID:     userID,
			Ballot:     ballot,
		}

		data, err := castVote.Serialize(serdecontext)
		if err != nil {
			return nil, xerrors.Errorf("failed to serialize castVote: %v", err)
		}

		args := []txn.Arg{
			{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
			{Key: evoting.ElectionArg, Value: data},
			{Key: evoting.CmdArg, Value: []byte(evoting.CmdCastVote)},
		}
		_, err = m.addAndWait(args...)
		if err != nil {
			return nil, xerrors.Errorf("failed to addAndWait: %v", err)
		}

		var ballot types.Ballot
		err = ballot.Unmarshal(vote, election)
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
