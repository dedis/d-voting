package integration

import (
	"bytes"
	"context"
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

	// create election
	election, err := json.Marshal(types.CreateElectionTransaction{
		Title:   "Some Election",
		AdminID: "anAdminID",
		Format:  "majority",
	})
	require.NoError(t, err)
	args = []txn.Arg{
		{Key: "go.dedis.ch/dela.ContractArg", Value: []byte(evoting.ContractName)},
		{Key: evoting.CreateElectionArg, Value: election},
		{Key: evoting.CmdArg, Value: []byte(evoting.CmdCreateElection)},
	}
	addAndWait(t, manager, nodes[0].(dVotingNode), args...)

	// DKG init
	var dkg dkg.DKG
	for _, node := range nodes {
		dkg = node.(dVotingNode).GetDkg()
		_, err := dkg.Listen()
		require.NoError(t, err)
	}

	// Shuffle init

	// DKG setup

	// pubKey, err := dkg.Setup(
	// require.NoError(t, err)

	// for _, node := range nodes {
	// 	s := node.(dVotingNode).GetShuffle()
	// 	shuffleActor, err := s.Listen()
	// 	require.NoError(t, err)
	// }

	// key1 := make([]byte, 32)

	// _, err = rand.Read(key1)
	// require.NoError(t, err)

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

	// key2 := make([]byte, 32)

	// _, err = rand.Read(key2)
	// require.NoError(t, err)

	// args = []txn.Arg{
	// 	{Key: "go.dedis.ch/dela.ContractArg", Value: []byte("go.dedis.ch/dela.Value")},
	// 	{Key: "value:key", Value: key2},
	// 	{Key: "value:value", Value: []byte("value2")},
	// 	{Key: "value:command", Value: []byte("WRITE")},
	// }
	// addAndWait(t, manager, nodes[0].(cosiDelaNode), args...)
}

// -----------------------------------------------------------------------------
// Utility functions

func addAndWait(t *testing.T, manager txn.Manager, node dVotingNode, args ...txn.Arg) {
	manager.Sync()

	tx, err := manager.Make(args...)
	require.NoError(t, err)

	err = node.GetPool().Add(tx)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	events := node.GetOrdering().Watch(ctx)

	for event := range events {
		for _, result := range event.Transactions {
			tx := result.GetTransaction()

			if bytes.Equal(tx.GetID(), tx.GetID()) {
				accepted, err := event.Transactions[0].GetStatus()
				require.Empty(t, err)

				require.True(t, accepted)
				return
			}
		}
	}

	t.Error("transaction not found")
}
