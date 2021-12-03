package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/stretchr/testify/require"
	accessContract "go.dedis.ch/dela/contracts/access"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto/bls"
	"go.dedis.ch/dela/crypto/loader"
	"golang.org/x/xerrors"
)

// Start 3 nodes
// Use the value contract
// Check the state
func TestIntegration_Scenario(t *testing.T) {

	dir, err := ioutil.TempDir(os.TempDir(), "d-voting-integration-test")

	require.NoError(t, err)

	defer os.RemoveAll(dir)

	t.Logf("using temp dir %s", dir)

	// create nodes
	nodes := []dela{
		newDVotingNode(t, filepath.Join(dir, "node1"), 2001),
		newDVotingNode(t, filepath.Join(dir, "node2"), 2002),
		newDVotingNode(t, filepath.Join(dir, "node3"), 2003),
	}

	nodes[0].Setup(nodes[1:]...)

	l := loader.NewFileLoader(filepath.Join(dir, "private.key"))

	signerdata, err := l.LoadOrCreate(newKeyGenerator())
	require.NoError(t, err)

	signer, err := bls.NewSignerFromBytes(signerdata)
	require.NoError(t, err)

	pubKey := signer.GetPublicKey()
	cred := accessContract.NewCreds(aKey[:])

	for _, node := range nodes {
		node.GetAccessService().Grant(node.(dVotingNode).GetAccessStore(), cred, pubKey)
	}

	manager := signed.NewManager(signer, &txClient{})

	pubKeyBuf, err := signer.GetPublicKey().MarshalBinary()
	require.NoError(t, err)

	// grant access
	args := []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte("go.dedis.ch/dela.Access")},
		{Key: "access:grant_id", Value: []byte(hex.EncodeToString(evotingAccessKey[:]))},
		{Key: "access:grant_contract", Value: []byte("go.dedis.ch/dela.Evoting")},
		{Key: "access:grant_command", Value: []byte("all")},
		{Key: "access:identity", Value: []byte(base64.StdEncoding.EncodeToString(pubKeyBuf))},
		{Key: "access:command", Value: []byte("GRANT")},
	}
	addAndWait(t, manager, nodes[0].(dVotingNode), args...)

	// SC1: create election
	createElectionTransaction := types.CreateElectionTransaction{
		Title:   "Some Election",
		AdminID: "anAdminID",
		Format:  "majority",
	}
	election, err := json.Marshal(createElectionTransaction)
	require.NoError(t, err)
	args = []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.CreateElectionArg, Value: election},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCreateElection)},
	}
	txID := addAndWait(t, manager, nodes[0].(dVotingNode), args...)

	// Calculate electionID from
	hash := sha256.New()
	hash.Write(txID)
	electionID := hash.Sum(nil)

	// DK1: DKG init
	var dkg []dkg.DKG
	for _, node := range nodes {
		new_dkg := node.(dVotingNode).GetDkg()
		_, err := new_dkg.Listen()
		require.NoError(t, err)
		dkg = append(dkg, new_dkg)
	}

	// NS1: Neff shuffle init
	for _, node := range nodes {
		s := node.(dVotingNode).GetShuffle()
		_, err := s.Listen(signer)
		require.NoError(t, err)
	}

	// SC2: get election info, but not used for now

	// DK2: DKG setup
	actor, err := dkg[0].GetLastActor()
	require.NoError(t, err)
	actor.Setup(electionID)

	// SC3: open election
	args = []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.CreateElectionArg, Value: electionID},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdOpenElection)},
	}
	addAndWait(t, manager, nodes[0].(dVotingNode), args...)

	// SC4: cast vote1
	ballot1, err := marshallBallot("vote1", actor)
	require.NoError(t, err)

	castVoteTransaction := types.CastVoteTransaction{
		ElectionID: hex.EncodeToString(electionID),
		UserID:     "user 1",
		Ballot:     ballot1,
	}
	vote, err := json.Marshal(castVoteTransaction)
	require.NoError(t, err)

	args = []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.CreateElectionArg, Value: vote},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCastVote)},
	}
	addAndWait(t, manager, nodes[0].(dVotingNode), args...)

	// SC4: cast vote2
	ballot2, err := marshallBallot("vote2", actor)
	require.NoError(t, err)

	castVoteTransaction = types.CastVoteTransaction{
		ElectionID: hex.EncodeToString(electionID),
		UserID:     "user 2",
		Ballot:     ballot2,
	}
	vote, err = json.Marshal(castVoteTransaction)
	require.NoError(t, err)

	args = []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.CreateElectionArg, Value: vote},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCastVote)},
	}
	addAndWait(t, manager, nodes[0].(dVotingNode), args...)

	// SC4: cast vote3
	ballot3, err := marshallBallot("vote1", actor)
	require.NoError(t, err)

	castVoteTransaction = types.CastVoteTransaction{
		ElectionID: hex.EncodeToString(electionID),
		UserID:     "user 3",
		Ballot:     ballot3,
	}
	vote, err = json.Marshal(castVoteTransaction)
	require.NoError(t, err)

	args = []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.CreateElectionArg, Value: vote},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCastVote)},
	}
	addAndWait(t, manager, nodes[0].(dVotingNode), args...)

	// SC5: close election

	// SC6: shuffle

	// SC7: decrypt

	// SC8: get result

	// args = []txn.Arg{
	// 	{Key: "go.dedis.ch/dela.ContractArg", Value: []byte("go.dedis.ch/dela.Evoting")},
	// 	{Key: "value:key", Value: key1},
	// 	{Key: "value:value", Value: []byte("value1")},
	// 	{Key: "value:command", Value: []byte("WRITE")},
	// }
	// addAndWait(t, manager, nodes[0].(cosiDelaNode), args...)

	// proof, err := nodes[0].GetOrdering().GetProof(key1)
	// require.NoError(t, err)
	// require.Equal(t, []byte("value1"), proof.GetValue())
}

// -----------------------------------------------------------------------------
// Utility functions

func addAndWait(t *testing.T, manager txn.Manager, node dVotingNode, args ...txn.Arg) []byte {
	manager.Sync()

	tx, err := manager.Make(args...)
	require.NoError(t, err)
	txID := tx.GetID()

	err = node.GetPool().Add(tx)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	events := node.GetOrdering().Watch(ctx)

	for event := range events {
		for _, result := range event.Transactions {
			tx2 := result.GetTransaction()
			if bytes.Equal(txID, tx2.GetID()) {
				accepted, status := event.Transactions[0].GetStatus()
				require.Empty(t, status)

				require.True(t, accepted)
				return txID
			}
		}
	}

	// force failed test if transaction failed
	t.Error("transaction not found")

	return txID
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
