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

// Check the shuffled votes versus the casted votes an a few nodes
func TestIntegration_ThreeVotesScenario(t *testing.T) {
	numNodes := 3
	numVotes := 3
	adminID := "first admin"

	// ##### SETUP ENV #####
	// make tests reproducible
	rand.Seed(1)
	delaPkg.Logger = zerolog.New(os.Stdout).Level(zerolog.WarnLevel)

	dirPath, err := ioutil.TempDir(os.TempDir(), "d-voting-three-votes")
	require.NoError(t, err)

	defer os.RemoveAll(dirPath)

	t.Logf("using temp dir %s", dirPath)

	// ##### CREATE NODES #####
	nodes := setupDVotingNodes(t, numNodes, dirPath)

	signer := createDVotingAccess(t, nodes, dirPath)

	m := newTxManager(signer, nodes[0], time.Second*3)

	err = grantAccess(m, signer)
	require.NoError(t, err)

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
	sActor, err := initShuffle(nodes, signer)
	require.NoError(t, err)

	t.Logf("shuffling")
	err = sActor.Shuffle(electionID)
	require.NoError(t, err)

	// ##### DECRYPT BALLOTS #####
	time.Sleep(time.Second * 3)

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

// Check more shuffled votes versus the casted votes on more nodes.
func TestIntegration_ManyVotesScenario(t *testing.T) {
	// The following constants are limited by VSC built in debug function that times out after 30s.
	numNodes := 10
	numVotes := 10
	adminID := "I am an admin"

	// ##### SETUP ENV #####
	// make tests reproducible
	rand.Seed(2)
	delaPkg.Logger = zerolog.New(os.Stdout).Level(zerolog.WarnLevel)

	dirPath, err := ioutil.TempDir(os.TempDir(), "d-voting-many-votes")
	require.NoError(t, err)

	defer os.RemoveAll(dirPath)

	t.Logf("using temp dir %s", dirPath)

	// ##### CREATE NODES #####
	nodes := setupDVotingNodes(t, numNodes, dirPath)

	signer := createDVotingAccess(t, nodes, dirPath)

	m := newTxManager(signer, nodes[0], time.Second*3)

	err = grantAccess(m, signer)
	require.NoError(t, err)

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
	sActor, err := initShuffle(nodes, signer)
	require.NoError(t, err)

	t.Logf("shuffling")
	err = sActor.Shuffle(electionID)
	require.NoError(t, err)

	// ##### DECRYPT BALLOTS #####
	time.Sleep(time.Second * 3)

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
func newTxManager(signer crypto.Signer, firstNode dVotingCosiDela, timeout time.Duration) txManager {
	return txManager{
		m: signed.NewManager(signer, &txClient{}),
		n: firstNode,
		t: timeout,
	}
}

type txManager struct {
	m txn.Manager
	n dVotingCosiDela
	t time.Duration
}

func (m txManager) addAndWait(args ...txn.Arg) ([]byte, error) {
	err := m.m.Sync()
	if err != nil {
		return nil, xerrors.Errorf("failed to Sync: %v", err)
	}

	sentTxn, err := m.m.Make(args...)
	if err != nil {
		return nil, xerrors.Errorf("failed to Make: %v", err)
	}
	sentTxnID := sentTxn.GetID()

	err = m.n.GetPool().Add(sentTxn)
	if err != nil {
		return nil, xerrors.Errorf("failed to Add: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.t)
	defer cancel()

	events := m.n.GetOrdering().Watch(ctx)

	for event := range events {
		for _, result := range event.Transactions {
			fetchedTxnID := result.GetTransaction().GetID()

			if bytes.Equal(sentTxnID, fetchedTxnID) {
				accepted, status := event.Transactions[0].GetStatus()

				if !accepted {
					return nil, xerrors.Errorf("transaction has not been accepted: %s", status)
				}

				return sentTxnID, nil
			}
		}
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
				ID:       "0xaaa",
				Title:    "subject1",
				Order:    nil,
				Subjects: nil,
				Selects: []types.Select{
					{
						ID:      "0xbbb",
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
				ID:       "0xddd",
				Title:    "subject2",
				Order:    nil,
				Subjects: nil,
				Selects:  nil,
				Ranks:    nil,
				Texts: []types.Text{
					{
						ID:        "0xeee",
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
		"select:0xbbb:0,0,1,0\n" + "text:0xeee:eWVz\n\n", //encoding of "yes"
		"select:0xbbb:1,1,0,0\n" + "text:0xeee:amE=\n\n", //encoding of "ja
		"select:0xbbb:0,0,0,1\n" + "text:0xeee:b3Vp\n\n", //encoding of "oui"
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
		actor, err = d.Listen()
		if err != nil {
			return nil, xerrors.Errorf("failed to GetDkg: %v", err)
		}
	}

	_, err = actor.Setup(electionID)
	if err != nil {
		return nil, xerrors.Errorf("failed to Setup: %v", err)
	}

	return actor, nil
}

func initShuffle(nodes []dVotingCosiDela, signer crypto.AggregateSigner) (shuffle.Actor, error) {
	var sActor shuffle.Actor
	for _, node := range nodes {
		var err error
		s := node.GetShuffle()
		sActor, err = s.Listen(signer)
		if err != nil {
			return nil, xerrors.Errorf("failed to init Shuffle: %v", err)
		}
	}
	time.Sleep(time.Second * 1)

	return sActor, nil
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