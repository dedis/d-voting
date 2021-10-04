package evoting

// todo: json marshall and unmarshall branch is are not covered yet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	"go.dedis.ch/kyber/v3/util/random"
)

var dummyElectionIdBuff = []byte("dummyID")
var fakeElectionID = hex.EncodeToString(dummyElectionIdBuff)

func fakeProver(proof.Suite, string, proof.Verifier, []byte) error {
	return nil
}

func TestExecute(t *testing.T) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}
	contract := NewContract([]byte{}, fakeAccess{err: fake.GetError()}, fakeDkg)

	err := contract.Execute(fakeStore{}, makeStep(t))
	require.EqualError(t, err, "identity not authorized: fake.PublicKey ("+fake.GetError().Error()+")")

	contract = NewContract([]byte{}, fakeAccess{}, fakeDkg)
	err = contract.Execute(fakeStore{}, makeStep(t))
	require.EqualError(t, err, fmt.Sprintf(errArgNotFound, CmdArg))

	contract.cmd = fakeCmd{err: fake.GetError()}

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCreateElection)))
	require.EqualError(t, err, fake.Err("failed to create election"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCastVote)))
	require.EqualError(t, err, fake.Err("failed to cast vote"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCloseElection)))
	require.EqualError(t, err, fake.Err("failed to close election"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdShuffleBallots)))
	require.EqualError(t, err, fake.Err("failed to shuffle ballots"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdDecryptBallots)))
	require.EqualError(t, err, fake.Err("failed to decrypt ballots"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCancelElection)))
	require.EqualError(t, err, fake.Err("failed to cancel election"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, "fake"))
	require.EqualError(t, err, "unknown command: fake")

	contract.cmd = fakeCmd{}
	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCreateElection)))
	require.NoError(t, err)

}

func TestCommand_CreateElection(t *testing.T) {
	fakeActor := fakeDkgActor{
		publicKey: suite.Point(),
		err:       nil,
	}

	fakeDKG := fakeDKG{
		actor: fakeActor,
		err:   nil,
	}

	dummyCreateElectionTransaction := types.CreateElectionTransaction{
		ElectionID: "dummyID",
		Title:      "dummyTitle",
		AdminId:    "dummyAdminId",
	}

	js, _ := json.Marshal(dummyCreateElectionTransaction)

	contract := NewContract([]byte{}, fakeAccess{}, fakeDKG)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err := cmd.createElection(fake.NewSnapshot(), makeStep(t), fakeActor)
	require.EqualError(t, err, fmt.Sprintf(errArgNotFound, CreateElectionArg))

	err = cmd.createElection(fake.NewSnapshot(), makeStep(t, CreateElectionArg, "dummy"), fakeActor)
	require.EqualError(t, err, "failed to unmarshal CreateElectionTransaction : "+
		"invalid character 'd' looking for beginning of value")

	err = cmd.createElection(fake.NewBadSnapshot(), makeStep(t, CreateElectionArg, string(js)), fakeActor)
	require.EqualError(t, err, "failed to decode Election ID: encoding/hex: invalid byte: U+0075 'u'")

	dummyCreateElectionTransaction.ElectionID = hex.EncodeToString([]byte("dummyID"))
	js, _ = json.Marshal(dummyCreateElectionTransaction)
	err = cmd.createElection(fake.NewBadSnapshot(), makeStep(t, CreateElectionArg, string(js)), fakeActor)
	require.EqualError(t, err, fake.Err("failed to set value"))

	snap := fake.NewSnapshot()
	err = cmd.createElection(snap, makeStep(t, CreateElectionArg, string(js)), fakeActor)
	require.NoError(t, err)

	dummyElectionIdBuff, _ := hex.DecodeString(dummyCreateElectionTransaction.ElectionID)
	res, err := snap.Get(dummyElectionIdBuff)
	require.NoError(t, err)

	election := new(types.Election)
	_ = json.NewDecoder(bytes.NewBuffer(res)).Decode(election)

	require.Equal(t, dummyCreateElectionTransaction.ElectionID, string(election.ElectionID))
	require.Equal(t, dummyCreateElectionTransaction.Title, election.Title)
	require.Equal(t, dummyCreateElectionTransaction.AdminId, election.AdminId)
	require.Equal(t, types.Open, election.Status)

}

func TestCommand_CastVote(t *testing.T) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}

	dummyElectionIdBuff := []byte("dummyID")

	dummyCastVoteTransaction := types.CastVoteTransaction{
		ElectionID: fakeElectionID,
		UserId:     "dummyUserId",
		Ballot:     []byte{10},
	}
	jsCastVoteTransaction, _ := json.Marshal(dummyCastVoteTransaction)

	dummyElection := types.Election{
		Title:            "dummyTitle",
		ElectionID:       types.ID(fakeElectionID),
		AdminId:          "dummyAdminId",
		Status:           0,
		Pubkey:           []byte{},
		EncryptedBallots: &types.EncryptedBallots{},
		ShuffledBallots:  nil,
		ShuffledProofs:   nil,
		DecryptedBallots: nil,
		ShuffleThreshold: 0,
	}
	jsElection, err := json.Marshal(dummyElection)
	require.NoError(t, err)

	contract := NewContract([]byte{}, fakeAccess{}, fakeDkg)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.castVote(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, fmt.Sprintf(errArgNotFound, CastVoteArg))

	err = cmd.castVote(fake.NewSnapshot(), makeStep(t, CastVoteArg, "dummy"))
	require.EqualError(t, err, "failed to unmarshal CastVoteTransaction: "+
		"invalid character 'd' looking for beginning of value")

	err = cmd.castVote(fake.NewBadSnapshot(), makeStep(t, CastVoteArg, string(jsCastVoteTransaction)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	_ = snap.Set(dummyElectionIdBuff, []byte("fake election"))
	err = cmd.castVote(snap, makeStep(t, CastVoteArg, string(jsCastVoteTransaction)))
	require.Contains(t, err.Error(), "failed to unmarshal Election")

	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.castVote(snap, makeStep(t, CastVoteArg, string(jsCastVoteTransaction)))
	require.EqualError(t, err, fmt.Sprintf("the election is not open, current status: %d", types.Initial))

	dummyElection.Status = types.Open

	jsElection, _ = json.Marshal(dummyElection)

	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.castVote(snap, makeStep(t, CastVoteArg, string(jsCastVoteTransaction)))
	require.NoError(t, err)

	res, err := snap.Get(dummyElectionIdBuff)
	require.NoError(t, err)

	election := new(types.Election)
	_ = json.NewDecoder(bytes.NewBuffer(res)).Decode(election)

	require.Equal(t, dummyCastVoteTransaction.Ballot, election.EncryptedBallots.Ballots[0])
	require.Equal(t, dummyCastVoteTransaction.UserId, election.EncryptedBallots.UserIDs[0])
}

func TestCommand_CloseElection(t *testing.T) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}

	dummyCloseElectionTransaction := types.CloseElectionTransaction{
		ElectionID: fakeElectionID,
		UserId:     "dummyUserId",
	}
	jsCloseElectionTransaction, _ := json.Marshal(dummyCloseElectionTransaction)

	dummyElection := types.Election{
		Title:            "dummyTitle",
		ElectionID:       types.ID(fakeElectionID),
		AdminId:          "dummyAdminId",
		Status:           0,
		Pubkey:           nil,
		EncryptedBallots: &types.EncryptedBallots{},
		ShuffledBallots:  nil,
		ShuffledProofs:   nil,
		DecryptedBallots: nil,
		ShuffleThreshold: 0,
	}
	jsElection, err := json.Marshal(dummyElection)
	require.NoError(t, err)

	contract := NewContract([]byte{}, fakeAccess{}, fakeDkg)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.closeElection(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, fmt.Sprintf(errArgNotFound, CloseElectionArg))

	err = cmd.closeElection(fake.NewSnapshot(), makeStep(t, CloseElectionArg, "dummy"))
	require.EqualError(t, err, "failed to unmarshal CloseElectionTransaction: "+
		"invalid character 'd' looking for beginning of value")

	err = cmd.closeElection(fake.NewBadSnapshot(), makeStep(t, CloseElectionArg, string(jsCloseElectionTransaction)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	_ = snap.Set(dummyElectionIdBuff, []byte("fake election"))
	err = cmd.closeElection(snap, makeStep(t, CloseElectionArg, string(jsCloseElectionTransaction)))
	require.Contains(t, err.Error(), "failed to unmarshal Election")

	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.closeElection(snap, makeStep(t, CloseElectionArg, string(jsCloseElectionTransaction)))
	require.EqualError(t, err, "only the admin can close the election")

	dummyCloseElectionTransaction.UserId = "dummyAdminId"
	jsCloseElectionTransaction, _ = json.Marshal(dummyCloseElectionTransaction)
	err = cmd.closeElection(snap, makeStep(t, CloseElectionArg, string(jsCloseElectionTransaction)))
	require.EqualError(t, err, fmt.Sprintf("the election is not open, current status: %d", types.Initial))

	dummyElection.Status = types.Open
	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)

	err = cmd.closeElection(snap, makeStep(t, CloseElectionArg, string(jsCloseElectionTransaction)))
	require.EqualError(t, err, "at least two ballots are required")

	// dummyElection.EncryptedBallots["dummyUser1"] = []byte("dummyBallot1")
	// dummyElection.EncryptedBallots["dummyUser2"] = []byte("dummyBallot2")
	dummyElection.EncryptedBallots.CastVote("dummyUser1", []byte("dummyBallot1"))
	dummyElection.EncryptedBallots.CastVote("dummyUser2", []byte("dummyBallot2"))

	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)

	err = cmd.closeElection(snap, makeStep(t, CloseElectionArg, string(jsCloseElectionTransaction)))
	require.NoError(t, err)

	res, err := snap.Get(dummyElectionIdBuff)
	require.NoError(t, err)

	election := new(types.Election)
	_ = json.NewDecoder(bytes.NewBuffer(res)).Decode(election)

	require.Equal(t, types.Closed, election.Status)
}

func TestCommand_ShuffleBallots(t *testing.T) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}

	k := 3

	dummyShuffleBallotsTransaction := types.ShuffleBallotsTransaction{
		ElectionID:      fakeElectionID,
		Round:           2,
		ShuffledBallots: make([][]byte, 3),
		Proof:           nil,
	}
	jsShuffleBallotsTransaction, _ := json.Marshal(dummyShuffleBallotsTransaction)

	dummyElection := types.Election{
		Title:            "dummyTitle",
		ElectionID:       types.ID(fakeElectionID),
		AdminId:          "dummyAdminId",
		Status:           0,
		Pubkey:           nil,
		EncryptedBallots: &types.EncryptedBallots{},
		ShuffledBallots:  [][][]byte{},
		ShuffledProofs:   [][]byte{},
		DecryptedBallots: nil,
		ShuffleThreshold: 0,
	}
	jsElection, _ := json.Marshal(dummyElection)

	contract := NewContract([]byte{}, fakeAccess{}, fakeDkg)

	cmd := evotingCommand{
		Contract: &contract,
		prover:   fakeProver,
	}

	err := cmd.shuffleBallots(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, fmt.Sprintf(errArgNotFound, ShuffleBallotsArg))

	err = cmd.shuffleBallots(fake.NewSnapshot(), makeStep(t, ShuffleBallotsArg, "dummy"))
	require.EqualError(t, err, "failed to unmarshal ShuffleBallotsTransaction: "+
		"invalid character 'd' looking for beginning of value")

	err = cmd.shuffleBallots(fake.NewBadSnapshot(), makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	_ = snap.Set(dummyElectionIdBuff, []byte("fake election"))
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.Contains(t, err.Error(), "failed to unmarshal Election")

	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "the election is not closed")

	dummyElection.Status = types.Closed
	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "wrong number of shuffled ballots: expected '1', got '0'")

	dummyShuffleBallotsTransaction.Round = 1

	RandomStream := suite.RandomStream()
	h := suite.Scalar().Pick(RandomStream)
	pubKey := suite.Point().Mul(h, nil)

	KsMarshalled := make([][]byte, 0, k)
	CsMarshalled := make([][]byte, 0, k)

	for i := 0; i < k; i++ {
		// Embed the message into a curve point
		message := "Ballot" + strconv.Itoa(i)
		M := suite.Point().Embed([]byte(message), random.New())

		// ElGamal-encrypt the point to produce ciphertext (K,C).
		k := suite.Scalar().Pick(random.New()) // ephemeral private key
		K := suite.Point().Mul(k, nil)         // ephemeral DH public key
		S := suite.Point().Mul(k, pubKey)      // ephemeral DH shared secret
		C := S.Add(S, M)                       // message blinded with secret

		Kmarshalled, _ := K.MarshalBinary()
		Cmarshalled, _ := C.MarshalBinary()

		KsMarshalled = append(KsMarshalled, Kmarshalled)
		CsMarshalled = append(CsMarshalled, Cmarshalled)
	}

	for i := 0; i < k; i++ {
		dummyShuffleBallotsTransaction.ShuffledBallots[i] = []byte("badCiphertext")
	}
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to unmarshal Ciphertext: invalid character 'b' looking for beginning of value")

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: []byte("fakeVoteK"),
			C: []byte("fakeVoteC"),
		}
		js, _ := json.Marshal(ballot)
		dummyShuffleBallotsTransaction.ShuffledBallots[i] = js
	}
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to unmarshal K kyber.Point: invalid Ed25519 curve point")

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: []byte("fakeVoteC"),
		}
		js, _ := json.Marshal(ballot)
		dummyShuffleBallotsTransaction.ShuffledBallots[i] = js
	}
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to unmarshal C kyber.Point: invalid Ed25519 curve point")

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		}
		js, _ := json.Marshal(ballot)
		dummyShuffleBallotsTransaction.ShuffledBallots[i] = js
	}
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to unmarshal public key: invalid Ed25519 curve point")

	pubKeyMarshalled, _ := pubKey.MarshalBinary()
	dummyElection.Pubkey = pubKeyMarshalled

	for i := 0; i < k; i++ {
		ballot := []byte("badCiphertext")
		dummyElection.EncryptedBallots.CastVote(fmt.Sprintf("user%d", i), ballot)
	}

	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to unmarshal Ciphertext: invalid character 'b' looking for beginning of value")

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: []byte("fakeVoteK"),
			C: []byte("fakeVoteC"),
		}
		js, _ := json.Marshal(ballot)
		dummyElection.EncryptedBallots.CastVote(fmt.Sprintf("user%d", i), js)
	}

	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to unmarshal K kyber.Point: invalid Ed25519 curve point")

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: []byte("fakeVoteC"),
		}
		js, _ := json.Marshal(ballot)
		dummyElection.EncryptedBallots.CastVote(fmt.Sprintf("user%d", i), js)
	}

	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to unmarshal C kyber.Point: invalid Ed25519 curve point")

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		}
		js, _ := json.Marshal(ballot)
		dummyElection.EncryptedBallots.CastVote(fmt.Sprintf("user%d", i), js)
	}

	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.NoError(t, err)

	dummyShuffleBallotsTransaction.Round = k
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)

	for i := 1; i <= k-1; i++ {
		dummyElection.ShuffledBallots[i] = make([][]byte, 3)
	}

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		}
		js, _ := json.Marshal(ballot)
		dummyElection.ShuffledBallots[k-1][i] = js
	}

	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.NoError(t, err)

}

func TestCommand_DecryptBallots(t *testing.T) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}

	ballot1 := types.Ballot{Vote: "vote1"}
	ballot2 := types.Ballot{Vote: "vote2"}

	dummyDecryptBallotsTransaction := types.DecryptBallotsTransaction{
		ElectionID:       fakeElectionID,
		UserId:           "dummyUserId",
		DecryptedBallots: []types.Ballot{ballot1, ballot2},
	}
	jsDecryptBallotsTransaction, _ := json.Marshal(dummyDecryptBallotsTransaction)

	dummyElection := types.Election{
		Title:            "dummyTitle",
		ElectionID:       types.ID(fakeElectionID),
		AdminId:          "dummyAdminId",
		Status:           0,
		Pubkey:           nil,
		EncryptedBallots: &types.EncryptedBallots{},
		ShuffledBallots:  nil,
		ShuffledProofs:   nil,
		DecryptedBallots: []types.Ballot{},
		ShuffleThreshold: 0,
	}
	jsElection, _ := json.Marshal(dummyElection)

	contract := NewContract([]byte{}, fakeAccess{}, fakeDkg)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err := cmd.decryptBallots(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, fmt.Sprintf(errArgNotFound, DecryptBallotsArg))

	err = cmd.decryptBallots(fake.NewSnapshot(), makeStep(t, DecryptBallotsArg, "dummy"))
	require.EqualError(t, err, "failed to unmarshal DecryptBallotsTransaction: "+
		"invalid character 'd' looking for beginning of value")

	err = cmd.decryptBallots(fake.NewBadSnapshot(), makeStep(t, DecryptBallotsArg, string(jsDecryptBallotsTransaction)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	_ = snap.Set(dummyElectionIdBuff, []byte("fake election"))
	err = cmd.decryptBallots(snap, makeStep(t, DecryptBallotsArg, string(jsDecryptBallotsTransaction)))
	require.Contains(t, err.Error(), "failed to unmarshal Election")

	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.decryptBallots(snap, makeStep(t, DecryptBallotsArg, string(jsDecryptBallotsTransaction)))
	require.EqualError(t, err, "only the admin can decrypt the ballots")

	dummyDecryptBallotsTransaction.UserId = "dummyAdminId"
	jsDecryptBallotsTransaction, _ = json.Marshal(dummyDecryptBallotsTransaction)
	err = cmd.decryptBallots(snap, makeStep(t, DecryptBallotsArg, string(jsDecryptBallotsTransaction)))
	require.EqualError(t, err, fmt.Sprintf("the ballots are not shuffled, current status: %d", types.Initial))

	dummyElection.Status = types.ShuffledBallots

	jsElection, _ = json.Marshal(dummyElection)

	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.decryptBallots(snap, makeStep(t, DecryptBallotsArg, string(jsDecryptBallotsTransaction)))
	require.NoError(t, err)

	res, err := snap.Get(dummyElectionIdBuff)
	require.NoError(t, err)

	election := new(types.Election)
	_ = json.NewDecoder(bytes.NewBuffer(res)).Decode(election)

	require.Equal(t, dummyDecryptBallotsTransaction.DecryptedBallots, election.DecryptedBallots)
	require.Equal(t, types.ResultAvailable, election.Status)

}

func TestCommand_CancelElection(t *testing.T) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}

	dummyCancelElectionTransaction := types.CancelElectionTransaction{
		ElectionID: fakeElectionID,
		UserId:     "dummyUserId",
	}
	jsCancelElectionTransaction, _ := json.Marshal(dummyCancelElectionTransaction)

	dummyElection := types.Election{
		Title:            "dummyTitle",
		ElectionID:       types.ID(fakeElectionID),
		AdminId:          "dummyAdminId",
		Status:           1,
		Pubkey:           nil,
		EncryptedBallots: &types.EncryptedBallots{},
		ShuffledBallots:  nil,
		ShuffledProofs:   nil,
		DecryptedBallots: nil,
		ShuffleThreshold: 0,
	}
	jsElection, _ := json.Marshal(dummyElection)

	contract := NewContract([]byte{}, fakeAccess{}, fakeDkg)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err := cmd.cancelElection(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, fmt.Sprintf(errArgNotFound, CancelElectionArg))

	err = cmd.cancelElection(fake.NewSnapshot(), makeStep(t, CancelElectionArg, "dummy"))
	require.EqualError(t, err, "failed to unmarshal CancelElectionTransaction: "+
		"invalid character 'd' looking for beginning of value")

	err = cmd.cancelElection(fake.NewBadSnapshot(), makeStep(t, CancelElectionArg, string(jsCancelElectionTransaction)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	_ = snap.Set(dummyElectionIdBuff, []byte("fake election"))
	err = cmd.cancelElection(snap, makeStep(t, CancelElectionArg, string(jsCancelElectionTransaction)))
	require.Contains(t, err.Error(), "failed to unmarshal Election")

	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.cancelElection(snap, makeStep(t, CancelElectionArg, string(jsCancelElectionTransaction)))
	require.EqualError(t, err, "only the admin can cancel the election")

	dummyCancelElectionTransaction.UserId = "dummyAdminId"
	jsCancelElectionTransaction, _ = json.Marshal(dummyCancelElectionTransaction)
	err = cmd.cancelElection(snap, makeStep(t, CancelElectionArg, string(jsCancelElectionTransaction)))
	require.NoError(t, err)

	res, err := snap.Get(dummyElectionIdBuff)
	require.NoError(t, err)

	election := new(types.Election)
	_ = json.NewDecoder(bytes.NewBuffer(res)).Decode(election)

	require.Equal(t, types.Canceled, election.Status)

}

func TestRegisterContract(t *testing.T) {
	RegisterContract(native.NewExecution(), Contract{})
}

// -----------------------------------------------------------------------------
// Utility functions

func makeStep(t *testing.T, args ...string) execution.Step {
	return execution.Step{Current: makeTx(t, args...)}
}

func makeTx(t *testing.T, args ...string) txn.Transaction {
	options := []signed.TransactionOption{}
	for i := 0; i < len(args)-1; i += 2 {
		options = append(options, signed.WithArg(args[i], []byte(args[i+1])))
	}

	tx, err := signed.NewTransaction(0, fake.PublicKey{}, options...)
	require.NoError(t, err)

	return tx
}

type fakeDKG struct {
	actor fakeDkgActor
	err   error
}

func (f fakeDKG) Listen() (dkg.Actor, error) {
	return f.actor, f.err
}

func (f fakeDKG) GetLastActor() (dkg.Actor, error) {
	return f.actor, f.err
}

func (f fakeDKG) SetService(service ordering.Service) {
}

type fakeDkgActor struct {
	publicKey kyber.Point
	err       error
}

func (f fakeDkgActor) Setup(co crypto.CollectiveAuthority, threshold int) (pubKey kyber.Point, err error) {
	return nil, f.err
}

func (f fakeDkgActor) GetPublicKey() (kyber.Point, error) {
	return f.publicKey, f.err
}

func (f fakeDkgActor) Encrypt(message []byte) (K, C kyber.Point, remainder []byte, err error) {
	return nil, nil, nil, f.err
}

func (f fakeDkgActor) Decrypt(K, C kyber.Point, electionId string) ([]byte, error) {
	return nil, f.err
}

func (f fakeDkgActor) Reshare() error {
	return f.err
}

type fakeAccess struct {
	access.Service

	err error
}

func (srvc fakeAccess) Match(store.Readable, access.Credential, ...access.Identity) error {
	return srvc.err
}

func (srvc fakeAccess) Grant(store.Snapshot, access.Credential, ...access.Identity) error {
	return srvc.err
}

type fakeStore struct {
	store.Snapshot
}

func (s fakeStore) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (s fakeStore) Set(key, value []byte) error {
	return nil
}

type fakeCmd struct {
	err error
}

func (c fakeCmd) createElection(snap store.Snapshot, step execution.Step, dkgActor dkg.Actor) error {
	return c.err
}

func (c fakeCmd) castVote(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) closeElection(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) shuffleBallots(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) decryptBallots(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) cancelElection(snap store.Snapshot, step execution.Step) error {
	return c.err
}
