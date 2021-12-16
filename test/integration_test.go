package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strconv"
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
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"golang.org/x/xerrors"
)

// Check the shuffled votes versus the casted votes an a few nodes
func TestIntegration_ThreeVotesScenario(t *testing.T) {
	const N_NODES int = 3
	const N_VOTES int = 3

	delaPkg.Logger = zerolog.New(os.Stdout).Level(zerolog.WarnLevel)

	dirPath, err := ioutil.TempDir(os.TempDir(), "d-voting-3-votes")
	require.NoError(t, err)

	defer os.RemoveAll(dirPath)

	t.Logf("using temp dir %s", dirPath)

	nodes := setupDVotingNodes(t, N_NODES, dirPath)

	signer := createDVotingAccess(t, nodes, dirPath)

	m := newTxManager(signer, nodes[0].(dVotingNode), time.Second*3)

	err = grantAccess(m, signer)
	require.NoError(t, err)

	adminID := "anAdminID"
	electionID, err := createElection(m, "Three votes election", adminID, "majority")
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	actor, err := initDkg(nodes, electionID)
	require.NoError(t, err)

	err = openElection(m, electionID)
	require.NoError(t, err)

	possibleVotes := []string{"vote1", "vote2"}
	castedVotes, err := castVotesRandomly(m, actor, electionID, possibleVotes, N_VOTES)
	require.NoError(t, err)

	err = closeElection(m, electionID, adminID)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	sActor, err := initShuffle(nodes, signer)
	require.NoError(t, err)

	// SC6: shuffle
	err = sActor.Shuffle(electionID)
	require.NoError(t, err)

	// SC7: decrypt
	proof, err := nodes[0].(dVotingNode).GetOrdering().GetProof(electionID)
	require.NoError(t, err)

	electionMarshaled := proof.GetValue()
	election := &types.Election{}

	err = json.Unmarshal(electionMarshaled, election)
	require.NoError(t, err)

	shuffleInstances := election.ShuffleInstances
	nShuffleInstances := len(shuffleInstances)
	if nShuffleInstances == 0 {
		t.Errorf("Shuffle instances cannot be zero: %v", shuffleInstances)
	}

	shuffleLast := shuffleInstances[nShuffleInstances-1]

	// decrypt ballots
	ks, cs, err := shuffleLast.ShuffledBallots.GetKsCs()
	require.NoError(t, err)

	shuffledVotes := make([]string, len(ks))

	for i, k := range ks {
		c := cs[i]
		var ballot types.Ciphertext
		err = ballot.FromPoints(k, c)
		require.NoError(t, err)

		message, err := actor.Decrypt(k, c, electionID)
		require.NoError(t, err)

		shuffledVotes[i] = string(message)
	}

	// TODO: create transaction to add decrypted ballots on the blockchain

	sort.Strings(shuffledVotes)
	t.Logf("Shuffled votes: %v", shuffledVotes)
	sort.Strings(castedVotes)
	t.Logf("Casted votes: %v", castedVotes)

	require.Equal(t, castedVotes, shuffledVotes)

	t.Logf("Shuffled votes are equivalent to casted votes!")
}

// Check more shuffled votes versus the casted votes on more nodes.
func TestIntegration_ManyVotesScenario(t *testing.T) {
	// The following constants are limited by VSC built in debug function that times out after 30s.
	const N_NODES int = 10
	const N_VOTES int = 10

	delaPkg.Logger = zerolog.New(os.Stdout).Level(zerolog.WarnLevel)

	dirPath, err := ioutil.TempDir(os.TempDir(), "d-voting-many-votes")
	require.NoError(t, err)

	defer os.RemoveAll(dirPath)

	t.Logf("using temp dir %s", dirPath)

	nodes := setupDVotingNodes(t, N_NODES, dirPath)

	signer := createDVotingAccess(t, nodes, dirPath)

	m := newTxManager(signer, nodes[0].(dVotingNode), time.Second*3)

	err = grantAccess(m, signer)
	require.NoError(t, err)

	adminID := "anotherAdminID"
	electionID, err := createElection(m, "Many votes election", adminID, "majority")
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	actor, err := initDkg(nodes, electionID)
	require.NoError(t, err)

	err = openElection(m, electionID)
	require.NoError(t, err)

	t.Logf("start casting votes")
	possibleVotes := []string{
		"vote1",
		"vote2",
		"vote3",
		"vote4",
		"vote5",
	}
	castedVotes, err := castVotesRandomly(m, actor, electionID, possibleVotes, N_VOTES)
	require.NoError(t, err)

	err = closeElection(m, electionID, adminID)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	t.Logf("initializing shuffle")
	sActor, err := initShuffle(nodes, signer)
	require.NoError(t, err)

	t.Logf("shuffling")
	err = sActor.Shuffle(electionID)
	require.NoError(t, err)

	// SC7: decrypt
	t.Logf("get vote proof")
	proof, err := nodes[0].(dVotingNode).GetOrdering().GetProof(electionID)
	require.NoError(t, err)

	electionMarshaled := proof.GetValue()
	election := &types.Election{}

	err = json.Unmarshal(electionMarshaled, election)
	require.NoError(t, err)

	shuffleInstances := election.ShuffleInstances
	nShuffleInstances := len(shuffleInstances)
	if nShuffleInstances == 0 {
		t.Errorf("Shuffle instances cannot be zero: %v", shuffleInstances)
	}

	shuffleLast := shuffleInstances[nShuffleInstances-1]

	// decrypt ballots
	ks, cs, err := shuffleLast.ShuffledBallots.GetKsCs()
	require.NoError(t, err)

	shuffledVotes := make([]string, len(ks))

	for i, k := range ks {
		c := cs[i]
		var ballot types.Ciphertext
		err = ballot.FromPoints(k, c)
		require.NoError(t, err)

		message, err := actor.Decrypt(k, c, electionID)
		require.NoError(t, err)

		shuffledVotes[i] = string(message)
	}

	// TODO: create transaction to add decrypted ballots on the blockchain

	sort.Strings(shuffledVotes)
	t.Logf("Shuffled votes: %v", shuffledVotes)
	sort.Strings(castedVotes)
	t.Logf("Casted votes: %v", castedVotes)

	require.Equal(t, castedVotes, shuffledVotes)

	t.Logf("Shuffled votes are equivalent to casted votes!")
}

// -----------------------------------------------------------------------------
// Utility functions
func newTxManager(signer crypto.Signer, firstNode dVotingNode, timeout time.Duration) txManager {
	return txManager{
		m: signed.NewManager(signer, &txClient{}),
		n: firstNode,
		t: timeout,
	}
}

type txManager struct {
	m txn.Manager
	n dVotingNode
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

func createElection(m txManager, title, admin, format string) ([]byte, error) {
	electionData := types.CreateElectionTransaction{
		Title:   title,
		AdminID: admin,
		Format:  format,
	}

	createElectionBuf, err := json.Marshal(electionData)
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

func castVotesRandomly(m txManager, actor dkg.Actor, electionID []byte, possibleVotes []string,
	numberOfVotes int) ([]string, error) {

	var votes []string = make([]string, numberOfVotes)
	possibilities := len(possibleVotes)

	for i := 0; i < numberOfVotes; i++ {
		randomIndex := rand.Intn(possibilities)
		vote := possibleVotes[randomIndex]

		ballot, err := marshallBallot(vote, actor)
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

func marshallBallot(vote string, actor dkg.Actor) (types.Ciphertext, error) {
	K, C, _, err := actor.Encrypt([]byte(vote))
	if err != nil {
		return types.Ciphertext{}, xerrors.Errorf("failed to encrypt the plaintext: %v", err)
	}

	var ballot types.Ciphertext
	err = ballot.FromPoints(K, C)
	if err != nil {
		return types.Ciphertext{}, err
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
