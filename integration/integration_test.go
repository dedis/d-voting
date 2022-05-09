package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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
	t.Run("10 nodes, 10 votes", getIntegrationTest(10, 10))
}

func getIntegrationTest(numNodes, numVotes int) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

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

		m := newTxManager(signer, nodes[0], time.Second*time.Duration(numNodes/2+1), numNodes*2)

		err = grantAccess(m, signer)
		require.NoError(t, err)

		for _, n := range nodes {
			err = grantAccess(m, n.GetShuffleSigner())
			require.NoError(t, err)
		}

		// ##### CREATE ELECTION #####
		electionID, err := createElection(m, "Three votes election", adminID)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		// ##### SETUP DKG #####
		actor, err := initDkg(nodes, electionID, m.m)
		require.NoError(t, err)

		// ##### OPEN ELECTION #####
		err = openElection(m, electionID)
		require.NoError(t, err)

		electionFac := types.NewElectionFactory(types.CiphervoteFactory{}, nodes[0].GetRosterFac())

		t.Logf("start casting votes")
		election, err := getElection(electionFac, electionID, nodes[0].GetOrdering())
		require.NoError(t, err)

		castedVotes, err := castVotesRandomly(m, actor, election, numVotes)
		require.NoError(t, err)

		fmt.Println("casted votes:", castedVotes)

		// ##### CLOSE ELECTION #####
		err = closeElection(m, electionID, adminID)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		// ##### SHUFFLE BALLOTS #####
		t.Logf("initializing shuffle")
		sActor, err := initShuffle(nodes)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		t.Logf("shuffling")
		err = sActor.Shuffle(electionID)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		// ##### SUBMIT PUBLIC SHARES #####
		t.Logf("submitting public shares")

		election, err = getElection(electionFac, electionID, nodes[0].GetOrdering())
		require.NoError(t, err)
		err = actor.ComputePubshares()
		require.NoError(t, err)

		time.Sleep(time.Millisecond * 5000 * time.Duration(numNodes))

		// ##### DECRYPT BALLOTS #####
		t.Logf("decrypting")

		election, err = getElection(electionFac, electionID, nodes[0].GetOrdering())
		t.Logf("PubsharesUnit: %v", election.PubsharesUnits)
		require.NoError(t, err)
		err = decryptBallots(m, actor, election)
		require.NoError(t, err)

		time.Sleep(time.Second * 1)

		t.Logf("get vote proof")
		election, err = getElection(electionFac, electionID, nodes[0].GetOrdering())
		require.NoError(t, err)

		fmt.Println("Title of the election : " + election.Configuration.MainTitle)
		fmt.Println("ID of the election : " + string(election.ElectionID))
		fmt.Println("Status of the election : " + strconv.Itoa(int(election.Status)))
		fmt.Println("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))

		require.Len(t, election.DecryptedBallots, len(castedVotes))

		for _, b := range election.DecryptedBallots {
			ok := false
			for _, casted := range castedVotes {
				if b.Equal(casted) {
					ok = true
					break
				}
			}
			require.True(t, ok)
		}

		closeNodes(t, nodes)
	}
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

func createElection(m txManager, title string, admin string) ([]byte, error) {
	// Define the configuration :
	configuration := fake.BasicConfiguration

	createElection := types.CreateElection{
		Configuration: configuration,
		AdminID:       admin,
	}

	data, err := createElection.Serialize(serdecontext)
	if err != nil {
		return nil, xerrors.Errorf("failed to serialize: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.ElectionArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCreateElection)},
	}

	txID, err := m.addAndWait(args...)
	if err != nil {
		return nil, xerrors.Errorf(addAndWaitErr, err)
	}

	// Calculate electionID from
	hash := sha256.New()
	hash.Write(txID)
	electionID := hash.Sum(nil)

	return electionID, nil
}

func openElection(m txManager, electionID []byte) error {
	openElection := &types.OpenElection{
		ElectionID: hex.EncodeToString(electionID),
	}

	data, err := openElection.Serialize(serdecontext)
	if err != nil {
		return xerrors.Errorf("failed to serialize open election: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.ElectionArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdOpenElection)},
	}

	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf(addAndWaitErr, err)
	}

	return nil
}

func getElection(electionFac serde.Factory, electionID []byte,
	service ordering.Service) (types.Election, error) {

	election := types.Election{}

	proof, err := service.GetProof(electionID)
	if err != nil {
		return election, xerrors.Errorf("failed to GetProof: %v", err)
	}

	if proof == nil {
		return election, xerrors.Errorf("election does not exist: %v", err)
	}

	message, err := electionFac.Deserialize(serdecontext, proof.GetValue())
	if err != nil {
		return election, xerrors.Errorf("failed to deserialize Election: %v", err)
	}

	election, ok := message.(types.Election)
	if !ok {
		return election, xerrors.Errorf("wrong message type: %T", message)
	}

	return election, nil
}

func castVotesRandomly(m txManager, actor dkg.Actor, election types.Election,
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

		ciphervote, err := marshallBallot(strings.NewReader(vote), actor, election.ChunksPerBallot())
		if err != nil {
			return nil, xerrors.Errorf("failed to marshallBallot: %v", err)
		}

		userID := "user " + strconv.Itoa(i)

		castVote := types.CastVote{
			ElectionID: election.ElectionID,
			UserID:     userID,
			Ballot:     ciphervote,
		}

		data, err := castVote.Serialize(serdecontext)
		if err != nil {
			return nil, xerrors.Errorf("failed to serialize cast vote: %v", err)
		}

		args := []txn.Arg{
			{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
			{Key: evoting.ElectionArg, Value: data},
			{Key: evoting.CmdArg, Value: []byte(evoting.CmdCastVote)},
		}

		_, err = m.addAndWait(args...)
		if err != nil {
			return nil, xerrors.Errorf(addAndWaitErr, err)
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

func closeElection(m txManager, electionID []byte, admin string) error {
	closeElection := &types.CloseElection{
		ElectionID: hex.EncodeToString(electionID),
		UserID:     admin,
	}

	data, err := closeElection.Serialize(serdecontext)
	if err != nil {
		return xerrors.Errorf("failed to serialize open election: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.ElectionArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCloseElection)},
	}

	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf("failed to Marshall closeElection: %v", err)
	}

	return nil
}

func initDkg(nodes []dVotingCosiDela, electionID []byte, m txn.Manager) (dkg.Actor, error) {
	var actor dkg.Actor
	var err error

	for _, node := range nodes {
		d := node.(dVotingNode).GetDkg()

		// put Listen in a goroutine to optimize for speed
		actor, err = d.Listen(electionID, m)
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
	if election.Status != types.PubSharesSubmitted {
		return xerrors.Errorf("cannot decrypt: not all pubShares submitted")
	}

	decryptBallots := types.CombineShares{
		ElectionID: election.ElectionID,
	}

	data, err := decryptBallots.Serialize(serdecontext)
	if err != nil {
		return xerrors.Errorf("failed to serialize ballots: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte(evoting.ContractName)},
		{Key: evoting.ElectionArg, Value: data},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCombineShares)},
	}

	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf(addAndWaitErr, err)
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
