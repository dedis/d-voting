package integration

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	delaPkg "go.dedis.ch/dela"
	"go.dedis.ch/dela/core/txn"
	"golang.org/x/xerrors"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// Check the shuffled votes versus the casted votes and a few nodes
func BenchmarkIntegration_CustomVotesScenario(b *testing.B) {

	numNodes := 3
	numVotes := 10
	numChunksPerBallot := 3

	adminID := "first admin"

	// ##### SETUP ENV #####
	// make tests reproducible
	rand.Seed(1)
	delaPkg.Logger = zerolog.New(os.Stdout).Level(zerolog.WarnLevel)

	dirPath, err := ioutil.TempDir(os.TempDir(), "d-voting-three-votes")
	require.NoError(b, err)

	defer os.RemoveAll(dirPath)

	b.Logf("using temp dir %s", dirPath)

	// ##### CREATE NODES #####
	nodes := setupDVotingNodes(b, numNodes, dirPath)

	signer := createDVotingAccess(b, nodes, dirPath)

	m := newTxManager(signer, nodes[0], time.Second*3)

	err = grantAccess(m, signer)
	require.NoError(b, err)

	// ##### CREATE ELECTION #####
	electionID, err := createElectionNChunks(m, "Three votes election", adminID, numChunksPerBallot)
	require.NoError(b, err)

	time.Sleep(time.Millisecond * 100)

	// ##### SETUP DKG #####
	actor, err := initDkg(nodes, electionID)
	require.NoError(b, err)

	// ##### OPEN ELECTION #####
	err = openElection(m, electionID)
	require.NoError(b, err)

	b.Logf("start casting votes")
	election, err := getElection(electionID, nodes[0].GetOrdering())
	require.NoError(b, err)

	castedVotes, err := castVotesNChunks(m, actor, electionID, numVotes,
		election.ChunksPerBallot())
	require.NoError(b, err)

	// ##### CLOSE ELECTION #####
	err = closeElection(m, electionID, adminID)
	require.NoError(b, err)

	time.Sleep(time.Millisecond * 100)

	// ##### SHUFFLE BALLOTS #####
	b.ResetTimer()

	b.Logf("initializing shuffle")
	sActor, err := initShuffle(nodes, signer)
	require.NoError(b, err)

	b.Logf("shuffling")
	err = sActor.Shuffle(electionID)
	require.NoError(b, err)

	b.StopTimer()
	// ##### DECRYPT BALLOTS #####
	time.Sleep(time.Second * 1)

	b.Logf("decrypting")

	election, err = getElection(electionID, nodes[0].GetOrdering())
	require.NoError(b, err)

	//b.ResetTimer()
	err = decryptBallots(m, actor, election)
	require.NoError(b, err)
	//b.StopTimer()

	b.Logf("get vote proof")
	election, err = getElection(electionID, nodes[0].GetOrdering())
	require.NoError(b, err)

	fmt.Println("Title of the election : " + election.Configuration.MainTitle)
	fmt.Println("ID of the election : " + string(election.ElectionID))
	fmt.Println("Status of the election : " + strconv.Itoa(int(election.Status)))
	fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))
	fmt.Println("Chunks per ballot : " + strconv.Itoa(election.ChunksPerBallot()))

	// TODO: check that decrypted ballots are equals to casted ballots (maybe through hashing)
	for _, b := range election.DecryptedBallots {
		fmt.Println("decrypted ballot:", b)
	}

	for _, c := range castedVotes {
		fmt.Println("casted ballot:", c)
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

	createSimpleElectionRequest := types.CreateElectionRequest{
		Configuration: configuration,
		AdminID:       admin,
	}

	createElectionBuf, err := json.Marshal(createSimpleElectionRequest)
	if err != nil {
		return nil, xerrors.Errorf("failed to create createElectionBuf: %v", err)
	}

	args := []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.CreateElectionArg, Value: createElectionBuf},
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

func castVotesNChunks(m txManager, actor dkg.Actor, electionID []byte, numberOfVotes int,
	chunkPerBallot int) ([]string, error) {

	ballotBuilder := strings.Builder{}

	ballotBuilder.Write([]byte("text:"))
	ballotBuilder.Write([]byte(encodeID("bb")))
	ballotBuilder.Write([]byte(":"))

	textSize := 29*chunkPerBallot - ballotBuilder.Len() - 3

	ballotBuilder.Write([]byte(strings.Repeat("=", textSize)))
	ballotBuilder.Write([]byte("\n\n"))

	vote := ballotBuilder.String()

	ballot, err := marshallBallot(strings.NewReader(vote), actor, chunkPerBallot)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshallBallot: %v", err)
	}

	votes := make([]string, numberOfVotes)

	for i := 0; i < numberOfVotes; i++ {

		userID := "user " + strconv.Itoa(i)
		castVoteTransaction := types.CastVoteTransaction{
			ElectionID: hex.EncodeToString(electionID),
			UserID:     userID,
			Ballot:     ballot,
		}

		castedVoteBuf, err := json.Marshal(castVoteTransaction)
		if err != nil {
			return nil, xerrors.Errorf("failed to Marshall vote transaction: %v", err)
		}

		args := []txn.Arg{
			{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
			{Key: evoting.CastVoteArg, Value: castedVoteBuf},
			{Key: evoting.CmdArg, Value: []byte(evoting.CmdCastVote)},
		}
		_, err = m.addAndWait(args...)
		if err != nil {
			return nil, xerrors.Errorf("failed to addAndWait: %v", err)
		}

		votes[i] = vote
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

func decryptBallots(m txManager, actor dkg.Actor, election types.Election) error {
	if election.Status != types.ShuffledBallots {
		return xerrors.Errorf("cannot decrypt: shuffle is not finished")
	}

	X, Y, err := election.ShuffleInstances[election.ShuffleThreshold-1].ShuffledBallots.GetElGPairs()
	if err != nil {
		return xerrors.Errorf("failed to get Elg pairs")
	}

	decryptedBallots := make([]types.Ballot, 0, len(election.ShuffleInstances))
	wrongBallots := 0

	for i := 0; i < len(X[0]); i++ {
		// decryption of one ballot:
		marshalledBallot := strings.Builder{}
		for j := 0; j < len(X); j++ {
			chunk, err := actor.Decrypt(X[j][i], Y[j][i])
			if err != nil {
				return xerrors.Errorf("failed to decrypt (K,C): %v", err)
			}
			marshalledBallot.Write(chunk)
		}

		var ballot types.Ballot
		err = ballot.Unmarshal(marshalledBallot.String(), election)
		if err != nil {
			wrongBallots++
		}

		decryptedBallots = append(decryptedBallots, ballot)
	}

	decryptBallotsTransaction := types.DecryptBallotsTransaction{
		ElectionID:       election.ElectionID,
		UserID:           election.AdminID,
		DecryptedBallots: decryptedBallots,
	}

	decryptBallotsBuf, err := json.Marshal(decryptBallotsTransaction)
	if err != nil {
		return xerrors.Errorf("failed to marshal DecryptBallotsTransaction: %v", err)
	}

	args := []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.DecryptBallotsArg, Value: decryptBallotsBuf},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdDecryptBallots)},
	}
	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf("failed to addAndWait: %v", err)
	}

	return nil
}
