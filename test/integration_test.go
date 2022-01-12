package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
	_ "github.com/dedis/d-voting/services/dkg/pedersen/json"
	"github.com/dedis/d-voting/services/shuffle"
	_ "github.com/dedis/d-voting/services/shuffle/neff/json"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	delaPkg "go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

// Check the shuffled votes versus the cast votes on a few nodes
func TestIntegration_ThreeVotesScenario(t *testing.T) {
	numNodes := 3
	numVotes := 3
	adminID := "first admin"

	// ##### SETUP ENV #####
	// make tests reproducible
	rand.Seed(1)

	delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

	dirPath, err := ioutil.TempDir(os.TempDir(), "d-voting-three-votes")
	require.NoError(t, err)

	defer os.RemoveAll(dirPath)

	t.Logf("using temp dir %s", dirPath)

	// ##### CREATE NODES #####
	nodes := setupDVotingNodes(t, numNodes, dirPath)

	signer := createDVotingAccess(t, nodes, dirPath)

	m := newTxManager(signer, nodes[0], time.Second*10, 10)

	err = grantAccess(m, signer)
	require.NoError(t, err)

	for _, n := range nodes {
		err = grantAccess(m, n.GetShuffleSigner())
		require.NoError(t, err)
	}

	// ##### CREATE ELECTION #####
	electionID, err := createElection(m, "Three votes election", adminID)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	// ##### SETUP DKG #####
	actor, err := initDkg(nodes, electionID)
	require.NoError(t, err)

	// ##### OPEN ELECTION #####
	err = openElection(m, electionID)
	require.NoError(t, err)

	t.Logf("start casting votes")
	election, err := getElection(electionID, nodes[0].GetOrdering())
	require.NoError(t, err)

	castedVotes, err := castVotesRandomly(m, actor, electionID, numVotes,
		election.ChunksPerBallot())
	require.NoError(t, err)

	// ##### CLOSE ELECTION #####
	err = closeElection(m, electionID, adminID)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	// ##### SHUFFLE BALLOTS #####
	t.Logf("initializing shuffle")
	sActor, err := initShuffle(nodes)
	require.NoError(t, err)

	time.Sleep(time.Second * 1)

	t.Logf("shuffling")
	err = sActor.Shuffle(electionID)
	require.NoError(t, err)

	// ##### DECRYPT BALLOTS #####
	time.Sleep(time.Second * 1)

	t.Logf("decrypting")

	election, err = getElection(electionID, nodes[0].GetOrdering())
	require.NoError(t, err)

	err = decryptBallots(m, actor, election)
	require.NoError(t, err)

	time.Sleep(time.Second * 1)

	t.Logf("get vote proof")
	election, err = getElection(electionID, nodes[0].GetOrdering())
	require.NoError(t, err)

	fmt.Println("Title of the election : " + election.Configuration.MainTitle)
	fmt.Println("ID of the election : " + string(election.ElectionID))
	fmt.Println("Status of the election : " + strconv.Itoa(int(election.Status)))
	fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))

	// TODO: check that decrypted ballots are equals to cast ballots (maybe through hashing)
	for _, b := range election.DecryptedBallots {
		fmt.Println("decrypted ballot:", b)
	}

	for _, c := range castedVotes {
		fmt.Println("casted ballot:", c)
	}

	closeNodes(t, nodes)
}

// Check more shuffled votes versus the cast votes on more nodes.
func TestIntegration_ManyVotesScenario(t *testing.T) {
	// The following constants are limited by VSC build in debug function that times out after 30s.
	numNodes := 10
	numVotes := 10
	adminID := "I am an admin"

	// ##### SETUP ENV #####
	// make tests reproducible
	rand.Seed(2)

	delaPkg.Logger = delaPkg.Logger.Level(zerolog.WarnLevel)

	dirPath, err := ioutil.TempDir(os.TempDir(), "d-voting-many-votes")
	require.NoError(t, err)

	defer os.RemoveAll(dirPath)

	t.Logf("using temp dir %s", dirPath)

	// ##### CREATE NODES #####
	nodes := setupDVotingNodes(t, numNodes, dirPath)

	signer := createDVotingAccess(t, nodes, dirPath)

	m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*2)

	err = grantAccess(m, signer)
	require.NoError(t, err)

	for _, n := range nodes {
		err = grantAccess(m, n.GetShuffleSigner())
		require.NoError(t, err)
	}

	// ##### CREATE ELECTION #####
	electionID, err := createElection(m, "Many votes election", adminID)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	// ##### SETUP DKG #####
	actor, err := initDkg(nodes, electionID)
	require.NoError(t, err)

	// ##### OPEN ELECTION #####
	err = openElection(m, electionID)
	require.NoError(t, err)

	t.Logf("start casting votes")
	election, err := getElection(electionID, nodes[0].GetOrdering())
	require.NoError(t, err)

	castedVotes, err := castVotesRandomly(m, actor, electionID, numVotes,
		election.ChunksPerBallot())
	require.NoError(t, err)

	// ##### CLOSE ELECTION #####
	err = closeElection(m, electionID, adminID)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	// ##### SHUFFLE BALLOTS #####
	t.Logf("initializing shuffle")
	sActor, err := initShuffle(nodes)
	require.NoError(t, err)

	time.Sleep(time.Second * 1)

	t.Logf("shuffling")
	err = sActor.Shuffle(electionID)
	require.NoError(t, err)

	// ##### DECRYPT BALLOTS #####
	time.Sleep(time.Second * 10)

	election, err = getElection(electionID, nodes[0].GetOrdering())
	require.NoError(t, err)
	err = decryptBallots(m, actor, election)
	require.NoError(t, err)

	time.Sleep(time.Second * 1)

	t.Logf("get vote proof")
	election, err = getElection(electionID, nodes[0].GetOrdering())
	require.NoError(t, err)

	fmt.Println("Title of the election : " + election.Configuration.MainTitle)
	fmt.Println("ID of the election : " + string(election.ElectionID))
	fmt.Println("Status of the election : " + strconv.Itoa(int(election.Status)))
	fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))

	// TODO: check that decrypted ballots are equals to casted ballots (maybe through hashing)
	for _, b := range election.DecryptedBallots {
		fmt.Println("decrypted ballot:", b)
	}

	for _, c := range castedVotes {
		fmt.Println("casted ballot:", c)
	}

	closeNodes(t, nodes)
}

// -----------------------------------------------------------------------------
// Utility functions
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

	events:
		for event := range events {
			for _, result := range event.Transactions {
				fetchedTxnID := result.GetTransaction().GetID()

				if bytes.Equal(sentTxnID, fetchedTxnID) {
					accepted, status := event.Transactions[0].GetStatus()

					if !accepted {
						if i+1 == m.retry {
							return nil, xerrors.Errorf("transaction not accepted: %s", status)
						}

						// let's retry and sync the nonce in case
						err = m.m.Sync()
						if err != nil {
							return nil, xerrors.Errorf("failed to sync: %v", err)
						}

						break events
					}

					return sentTxnID, nil
				}
			}
		}

		cancel()
	}

	return nil, xerrors.Errorf("transaction not included after timeout: %v", args)
}

func grantAccess(m txManager, signer crypto.Signer) error {
	pubKeyBuf, err := signer.GetPublicKey().MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to GetPublicKey: %v", err)
	}

	args := []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte("go.dedis.ch/dela.Access")},
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

func createElection(m txManager, title string, admin string) ([]byte, error) {
	// Define the configuration :
	configuration := types.Configuration{
		MainTitle: title,
		Scaffold: []types.Subject{
			{
				ID:       encodeID("aa"),
				Title:    "subject1",
				Order:    nil,
				Subjects: nil,
				Selects: []types.Select{
					{
						ID:      encodeID("bb"),
						Title:   "Select your favorite snacks",
						MaxN:    3,
						MinN:    0,
						Choices: []string{"snickers", "mars", "vodka", "babibel"},
					},
				},
				Ranks: []types.Rank{},
				Texts: nil,
			},
			{
				ID:       encodeID("dd"),
				Title:    "subject2",
				Order:    nil,
				Subjects: nil,
				Selects:  nil,
				Ranks:    nil,
				Texts: []types.Text{
					{
						ID:        encodeID("ee"),
						Title:     "dissertation",
						MaxN:      1,
						MinN:      1,
						MaxLength: 3,
						Regex:     "",
						Choices:   []string{"write yes in your language"},
					},
				},
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

func openElection(m txManager, electionID []byte) error {
	openElectTransaction := &types.OpenElectionTransaction{
		ElectionID: hex.EncodeToString(electionID),
	}
	openElectionBuf, err := json.Marshal(openElectTransaction)
	if err != nil {
		return xerrors.Errorf("failed to Marshall: %v", err)
	}

	args := []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.OpenElectionArg, Value: openElectionBuf},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdOpenElection)},
	}
	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf("failed to addAndWait: %v", err)
	}

	return nil
}

func getElection(electionID []byte, service ordering.Service) (types.Election, error) {
	election := types.Election{}

	proof, err := service.GetProof(electionID)
	if err != nil {
		return election, xerrors.Errorf("failed to GetProof: %v", err)
	}

	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(&election)
	if err != nil {
		return election, xerrors.Errorf("failed to unmarshal Election: %v", err)
	}

	return election, nil
}

func castVotesRandomly(m txManager, actor dkg.Actor, electionID []byte, numberOfVotes int, chunkPerBallot int) ([]string, error) {
	possibleBallots := []string{
		string("select:" + encodeID("bb") + ":0,0,1,0\n" +
			"text:" + encodeID("ee") + ":eWVz\n\n"), //encoding of "yes"
		string("select:" + encodeID("bb") + ":1,1,0,0\n" +
			"text:" + encodeID("ee") + ":amE=\n\n"), //encoding of "ja
		string("select:" + encodeID("bb") + ":0,0,0,1\n" +
			"text:" + encodeID("ee") + "b3Vp\n\n"), //encoding of "oui"
	}

	votes := make([]string, numberOfVotes)

	for i := 0; i < numberOfVotes; i++ {
		randomIndex := rand.Intn(len(possibleBallots))
		vote := possibleBallots[randomIndex]

		ballot, err := marshallBallot(strings.NewReader(vote), actor, chunkPerBallot)
		if err != nil {
			return nil, xerrors.Errorf("failed to marshallBallot: %v", err)
		}

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

func marshallBallot(vote io.Reader, actor dkg.Actor, chunks int) (types.EncryptedBallot, error) {

	var ballot = make([]types.Ciphertext, chunks)

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
			return types.EncryptedBallot{}, xerrors.Errorf("failed to encrypt the plaintext: %v", err)
		}

		var chunk types.Ciphertext

		err = chunk.FromPoints(K, C)
		if err != nil {
			return types.EncryptedBallot{}, err
		}

		ballot[i] = chunk
	}

	return ballot, nil
}

func closeElection(m txManager, electionID []byte, admin string) error {
	closeElectTransaction := &types.CloseElectionTransaction{
		ElectionID: hex.EncodeToString(electionID),
		UserID:     admin,
	}
	closeElectionBuf, err := json.Marshal(closeElectTransaction)
	if err != nil {
		return xerrors.Errorf("failed to Marshall closeElection: %v", err)
	}

	args := []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.CloseElectionArg, Value: closeElectionBuf},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCloseElection)},
	}
	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf("failed to Marshall closeElection: %v", err)
	}

	return nil
}

func initDkg(nodes []dVotingCosiDela, electionID []byte) (dkg.Actor, error) {
	var actor dkg.Actor
	var err error

	for _, node := range nodes {
		d := node.(dVotingNode).GetDkg()

		// put Listen in a goroutine to optimize for speed
		actor, err = d.Listen(electionID)
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

func closeNodes(t *testing.T, nodes []dVotingCosiDela) {
	wait := sync.WaitGroup{}
	wait.Add(len(nodes))

	for _, n := range nodes {
		go func(node dVotingNode) {
			defer wait.Done()
			node.GetOrdering().Close()
		}(n.(dVotingNode))
	}

	t.Log("stopping nodes...")

	done := make(chan struct{})

	go func() {
		select {
		case <-done:
		case <-time.After(time.Second * 30):
			t.Error("timeout while closing nodes")
		}
	}()

	wait.Wait()
	close(done)
}

func encodeID(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
}
