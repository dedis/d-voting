package evoting

// todo: json marshall and unmarshall branch is are not covered yet

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/bls"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	"go.dedis.ch/kyber/v3/util/random"
	"strconv"
	"testing"
)

var dummyElectionIdBuff = []byte("dummyID")
var fakeElectionID = hex.EncodeToString(dummyElectionIdBuff)
var fakeCommonSigner = bls.NewSigner()

func fakeProver(proof.Suite, string, proof.Verifier, []byte) error {
	return nil
}

func TestExecute(t *testing.T) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}
	var evotingAccessKey = [32]byte{3}
	rosterKey := [32]byte{}

	service := fakeAccess{err: fake.GetError()}
	rosterFac := fakeAuthorityFactory{}

	contract := NewContract(evotingAccessKey[:], rosterKey[:], service, fakeDkg, rosterFac)

	err := contract.Execute(fakeStore{}, makeStep(t))
	require.EqualError(t, err, "identity not authorized: fake.PublicKey ("+fake.GetError().Error()+")")

	service = fakeAccess{}

	contract = NewContract(evotingAccessKey[:], rosterKey[:], service, fakeDkg, rosterFac)
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

	fakeDkg := fakeDKG{
		actor: fakeActor,
		err:   nil,
	}

	dummyCreateElectionTransaction := types.CreateElectionTransaction{
		Title:   "dummyTitle",
		AdminID: "dummyAdminID",
	}

	js, _ := json.Marshal(dummyCreateElectionTransaction)

	var evotingAccessKey = [32]byte{3}
	rosterKey := [32]byte{}

	service := fakeAccess{err: fake.GetError()}
	rosterFac := fakeAuthorityFactory{}

	contract := NewContract(evotingAccessKey[:], rosterKey[:], service, fakeDkg, rosterFac)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err := cmd.createElection(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, fmt.Sprintf(errArgNotFound, CreateElectionArg))

	err = cmd.createElection(fake.NewSnapshot(), makeStep(t, CreateElectionArg, "dummy"))
	require.EqualError(t, err, "failed to unmarshal CreateElectionTransaction : "+
		"invalid character 'd' looking for beginning of value")

	err = cmd.createElection(fake.NewBadSnapshot(), makeStep(t, CreateElectionArg, string(js)))
	require.EqualError(t, err, "failed to get roster")

	snap := fake.NewSnapshot()
	step := makeStep(t, CreateElectionArg, string(js))
	err = cmd.createElection(snap, step)
	require.NoError(t, err)

	// recover election ID:
	h := sha256.New()
	h.Write(step.Current.GetID())
	electionIdBuff := h.Sum(nil)

	res, err := snap.Get(electionIdBuff)
	require.NoError(t, err)

	election := new(types.Election)
	_ = json.NewDecoder(bytes.NewBuffer(res)).Decode(election)

	require.Equal(t, dummyCreateElectionTransaction.Title, election.Title)
	require.Equal(t, dummyCreateElectionTransaction.AdminID, election.AdminID)
	require.Equal(t, types.Open, election.Status)
}

func TestCommand_CastVote(t *testing.T) {
	dummyCastVoteTransaction := types.CastVoteTransaction{
		ElectionID: fakeElectionID,
		UserID:     "dummyUserId",
		Ballot: types.Ciphertext{
			K: []byte{},
			C: []byte{},
		},
	}

	jsCastVoteTransaction, _ := json.Marshal(dummyCastVoteTransaction)

	dummyElection, contract := initElectionAndContract()

	jsElection, err := json.Marshal(dummyElection)
	require.NoError(t, err)

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
	require.EqualError(t, err, "failed to decode Election ID: encoding/hex: invalid byte: U+0075 'u'")

	dummyElection.ElectionID = fakeElectionID
	jsElection, _ = json.Marshal(dummyElection)

	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.castVote(snap, makeStep(t, CastVoteArg, string(jsCastVoteTransaction)))
	require.NoError(t, err)

	res, err := snap.Get(dummyElectionIdBuff)
	require.NoError(t, err)

	election := new(types.Election)
	_ = json.NewDecoder(bytes.NewBuffer(res)).Decode(election)

	require.Equal(t, dummyCastVoteTransaction.Ballot,
		election.EncryptedBallots.Ballots[0])
	require.Equal(t, dummyCastVoteTransaction.UserID,
		election.EncryptedBallots.UserIDs[0])
}

func TestCommand_CloseElection(t *testing.T) {
	dummyCloseElectionTransaction := types.CloseElectionTransaction{
		ElectionID: fakeElectionID,
		UserID:     "dummyUserId",
	}
	jsCloseElectionTransaction, _ := json.Marshal(dummyCloseElectionTransaction)

	dummyElection, contract := initElectionAndContract()
	dummyElection.ElectionID = fakeElectionID

	jsElection, err := json.Marshal(dummyElection)
	require.NoError(t, err)

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

	dummyCloseElectionTransaction.UserID = "dummyAdminID"
	jsCloseElectionTransaction, _ = json.Marshal(dummyCloseElectionTransaction)

	err = cmd.closeElection(snap, makeStep(t, CloseElectionArg, string(jsCloseElectionTransaction)))
	require.EqualError(t, err, fmt.Sprintf("the election is not open, current status: %d", types.Initial))

	dummyElection.Status = types.Open
	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)

	err = cmd.closeElection(snap, makeStep(t, CloseElectionArg, string(jsCloseElectionTransaction)))
	require.EqualError(t, err, "at least two ballots are required")

	dummyElection.EncryptedBallots.CastVote("dummyUser1", types.Ciphertext{})
	dummyElection.EncryptedBallots.CastVote("dummyUser2", types.Ciphertext{})

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

func TestCommand_ShuffleBallotsCannotShuffleTwice(t *testing.T) {
	k := 3

	dummyElection, dummyShuffleBallotsTransaction, contract := initGoodShuffleBallot(k)

	cmd := evotingCommand{
		Contract: &contract,
		prover:   fakeProver,
	}

	snap := fake.NewSnapshot()

	// Attempts to shuffle twice :
	dummyShuffleBallotsTransaction.Round = 1

	dummyElection.ShuffleInstances = make([]types.ShuffleInstance, 1)
	dummyElection.ShuffleInstances[0].ShuffledBallots = make(types.Ciphertexts, 3)

	KsMarshalled, CsMarshalled, _ := fakeKCPointsMarshalled(k)
	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		}
		dummyElection.ShuffleInstances[0].ShuffledBallots[i] = ballot
	}

	dummyElection.ShuffleInstances[0].ShufflerPublicKey = dummyShuffleBallotsTransaction.PublicKey

	jsElection, _ := json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)
	jsShuffleBallotsTransaction, _ := json.Marshal(dummyShuffleBallotsTransaction)

	err := cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "a node already submitted a shuffle that has been accepted in round 0")
}

func TestCommand_ShuffleBallotsValidScenarios(t *testing.T) {
	k := 3

	// Simple Shuffle from round 0 :
	dummyElection, dummyShuffleBallotsTransaction, contract := initGoodShuffleBallot(k)

	cmd := evotingCommand{
		Contract: &contract,
		prover:   fakeProver,
	}

	snap := fake.NewSnapshot()
	jsElection, _ := json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)
	jsShuffleBallotsTransaction, _ := json.Marshal(dummyShuffleBallotsTransaction)

	err := cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.NoError(t, err)

	// Valid Shuffle is over :
	dummyShuffleBallotsTransaction.Round = k
	dummyElection.ShuffleInstances = make([]types.ShuffleInstance, k)
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)

	for i := 1; i <= k-1; i++ {
		dummyElection.ShuffleInstances[i].ShuffledBallots = make(types.Ciphertexts, 3)
	}

	KsMarshalled, CsMarshalled, _ := fakeKCPointsMarshalled(k)
	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		}
		dummyElection.ShuffleInstances[k-1].ShuffledBallots[i] = ballot
	}

	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.NoError(t, err)

	// Check the shuffle is over:
	electionTxIDBuff, _ := hex.DecodeString(dummyElection.ElectionID)
	electionMarshaled, _ := snap.Get(electionTxIDBuff)
	election := &types.Election{}
	_ = json.Unmarshal(electionMarshaled, election)

	require.Equal(t, election.Status, types.ShuffledBallots)
}

func TestCommand_ShuffleBallotsFormatErrors(t *testing.T) {
	k := 3

	dummyElection, dummyShuffleBallotsTransaction, contract := initBadShuffleBallot(k)
	jsShuffleBallotsTransaction, _ := json.Marshal(dummyShuffleBallotsTransaction)
	jsElection, _ := json.Marshal(dummyElection)

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

	// Wrong election id format
	snap := fake.NewSnapshot()
	_ = snap.Set(dummyElectionIdBuff, []byte("fake election"))

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.Contains(t, err.Error(), "failed to unmarshal Election")

	// Election not closed
	_ = snap.Set(dummyElectionIdBuff, jsElection)
	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "the election is not closed")

	// Wrong round :
	dummyElection.Status = types.Closed
	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "wrong shuffle round: expected round '0', transaction is for round '2'")

	// Missing public key of shuffler:
	dummyShuffleBallotsTransaction.Round = 1
	dummyElection.ShuffleInstances = make([]types.ShuffleInstance, 1)
	dummyShuffleBallotsTransaction.PublicKey = []byte("wrong Key")

	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)
	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "public key of the shuffler not found in roster: 77726f6e67204b6579")

	// Right key, wrong signature:
	dummyShuffleBallotsTransaction.PublicKey, _ = fakeCommonSigner.GetPublicKey().MarshalBinary()
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "could node deserialize shuffle signature : couldn't decode signature: couldn't deserialize data: unexpected end of JSON input")

	// Wrong election ID (Hash of shuffle fails)
	signature, _ := fakeCommonSigner.Sign([]byte("fake shuffle"))
	wrongSignature, _ := signature.Serialize(contract.context)
	dummyShuffleBallotsTransaction.Signature = wrongSignature
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "could not hash shuffle : Could not decode electionId : encoding/hex: invalid byte: U+0075 'u'")

	// Signatures not matching:
	dummyId := hex.EncodeToString([]byte("dummyId"))
	dummyElection.ElectionID = dummyId
	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "signature does not match the Shuffle : bls verify failed: bls: invalid signature ")

	// Wrong data in Shuffled Ballot (and fixed signatures)
	hash, _ := dummyShuffleBallotsTransaction.HashShuffle(dummyId)
	signature, _ = fakeCommonSigner.Sign(hash)
	wrongSignature, _ = signature.Serialize(contract.context)

	dummyShuffleBallotsTransaction.Signature = wrongSignature
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to get ks, cs: failed to get points: failed to unmarshal K: invalid Ed25519 curve point")

	// Good format, signature not updated thus not matching :
	KsMarshalled, CsMarshalled, pubKey := fakeKCPointsMarshalled(k)

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		}
		dummyShuffleBallotsTransaction.ShuffledBallots[i] = ballot
	}

	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "signature does not match the Shuffle : bls verify failed: bls: invalid signature ")

	// Signature matches, election bad public key :
	hash, _ = dummyShuffleBallotsTransaction.HashShuffle(dummyId)
	signature, _ = fakeCommonSigner.Sign(hash)
	wrongSignature, _ = signature.Serialize(contract.context)

	dummyShuffleBallotsTransaction.Signature = wrongSignature
	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to unmarshal public key: invalid Ed25519 curve point")

	// Wrong format in encrypted ballots :
	pubKeyMarshalled, _ := pubKey.MarshalBinary()
	dummyElection.Pubkey = pubKeyMarshalled
	dummyShuffleBallotsTransaction.Round = 0
	dummyElection.ShuffleInstances = make([]types.ShuffleInstance, 0)

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: []byte("fakeVoteK"),
			C: []byte("fakeVoteC"),
		}
		dummyElection.EncryptedBallots.CastVote(fmt.Sprintf("user%d", i), ballot)
	}

	jsShuffleBallotsTransaction, _ = json.Marshal(dummyShuffleBallotsTransaction)
	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to get ks, cs: failed to get points: failed to unmarshal K: invalid Ed25519 curve point")

	// C of encrypted ballots is wrong
	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: []byte("fakeVoteC"),
		}
		dummyElection.EncryptedBallots.CastVote(fmt.Sprintf("user%d", i), ballot)
	}

	jsElection, _ = json.Marshal(dummyElection)
	_ = snap.Set(dummyElectionIdBuff, jsElection)

	err = cmd.shuffleBallots(snap, makeStep(t, ShuffleBallotsArg, string(jsShuffleBallotsTransaction)))
	require.EqualError(t, err, "failed to get ks, cs: failed to get points: failed to unmarshal C: invalid Ed25519 curve point")

}

func TestCommand_DecryptBallots(t *testing.T) {
	ballot1 := types.Ballot{Vote: "vote1"}
	ballot2 := types.Ballot{Vote: "vote2"}

	dummyDecryptBallotsTransaction := types.DecryptBallotsTransaction{
		ElectionID:       fakeElectionID,
		UserID:           "dummyUserId",
		DecryptedBallots: []types.Ballot{ballot1, ballot2},
	}
	jsDecryptBallotsTransaction, _ := json.Marshal(dummyDecryptBallotsTransaction)

	dummyElection, contract := initElectionAndContract()
	dummyElection.ElectionID = fakeElectionID

	jsElection, err := json.Marshal(dummyElection)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.decryptBallots(fake.NewSnapshot(), makeStep(t))
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

	dummyDecryptBallotsTransaction.UserID = "dummyAdminID"
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
	dummyCancelElectionTransaction := types.CancelElectionTransaction{
		ElectionID: fakeElectionID,
		UserID:     "dummyUserId",
	}
	jsCancelElectionTransaction, _ := json.Marshal(dummyCancelElectionTransaction)

	dummyElection, contract := initElectionAndContract()
	dummyElection.ElectionID = fakeElectionID

	jsElection, err := json.Marshal(dummyElection)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.cancelElection(fake.NewSnapshot(), makeStep(t))
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

	dummyCancelElectionTransaction.UserID = "dummyAdminID"
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

func initElectionAndContract() (types.Election, Contract) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}

	dummyElection := types.Election{
		Title:            "dummyTitle",
		ElectionID:       "dummyID",
		AdminID:          "dummyAdminID",
		Status:           0,
		Pubkey:           nil,
		EncryptedBallots: types.EncryptedBallots{},
		ShuffleInstances: make([]types.ShuffleInstance, 0),
		DecryptedBallots: nil,
		ShuffleThreshold: 0,
	}

	var evotingAccessKey = [32]byte{3}
	rosterKey := [32]byte{}

	service := fakeAccess{err: fake.GetError()}
	rosterFac := fakeAuthorityFactory{}

	contract := NewContract(evotingAccessKey[:], rosterKey[:], service, fakeDkg, rosterFac)

	return dummyElection, contract
}

func initGoodShuffleBallot(k int) (types.Election, types.ShuffleBallotsTransaction, Contract) {
	dummyElection, dummyShuffleBallotsTransaction, contract := initBadShuffleBallot(3)
	dummyElection.Status = types.Closed

	KsMarshalled, CsMarshalled, pubKey := fakeKCPointsMarshalled(k)
	dummyShuffleBallotsTransaction.PublicKey, _ = fakeCommonSigner.GetPublicKey().MarshalBinary()

	// ShuffledBallots:
	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		}
		dummyShuffleBallotsTransaction.ShuffledBallots[i] = ballot
	}

	// Encrypted ballots:
	pubKeyMarshalled, _ := pubKey.MarshalBinary()
	dummyElection.Pubkey = pubKeyMarshalled
	dummyShuffleBallotsTransaction.Round = 0
	dummyElection.ShuffleInstances = make([]types.ShuffleInstance, 0)

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		}
		dummyElection.EncryptedBallots.CastVote(fmt.Sprintf("user%d", i), ballot)
	}

	// Valid Signature of shuffle
	dummyId := hex.EncodeToString([]byte("dummyId"))
	dummyElection.ElectionID = dummyId
	hash, _ := dummyShuffleBallotsTransaction.HashShuffle(dummyId)
	signature, _ := fakeCommonSigner.Sign(hash)
	wrongSignature, _ := signature.Serialize(contract.context)
	dummyShuffleBallotsTransaction.Signature = wrongSignature

	return dummyElection, dummyShuffleBallotsTransaction, contract
}

func initBadShuffleBallot(sizeOfElection int) (types.Election, types.ShuffleBallotsTransaction, Contract) {
	FakePubKey := fake.NewBadPublicKey()
	FakePubKeyMarshalled, _ := FakePubKey.MarshalBinary()
	shuffledBallots := make(types.Ciphertexts, sizeOfElection)

	dummyShuffleBallotsTransaction := types.ShuffleBallotsTransaction{
		ElectionID:      fakeElectionID,
		Round:           2,
		ShuffledBallots: shuffledBallots,
		Proof:           nil,
		PublicKey:       FakePubKeyMarshalled,
	}

	dummyElection, contract := initElectionAndContract()

	return dummyElection, dummyShuffleBallotsTransaction, contract

}

func fakeKCPointsMarshalled(k int) ([][]byte, [][]byte, kyber.Point) {
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
	return KsMarshalled, CsMarshalled, pubKey
}

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

func (f fakeDkgActor) Setup(electionID []byte) (pubKey kyber.Point, err error) {
	return nil, f.err
}

func (f fakeDkgActor) GetPublicKey() (kyber.Point, error) {
	return f.publicKey, f.err
}

func (f fakeDkgActor) Encrypt(message []byte) (K, C kyber.Point, remainder []byte, err error) {
	return nil, nil, nil, f.err
}

func (f fakeDkgActor) Decrypt(K, C kyber.Point, electionId []byte) ([]byte, error) {
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

func (c fakeCmd) createElection(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) openElection(snap store.Snapshot, step execution.Step, dkgActor dkg.Actor) error {
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

type fakeAuthorityFactory struct {
	serde.Factory

	//AuthorityOf(serde.Context, []byte) (authority.Authority, error)
}

func (f fakeAuthorityFactory) AuthorityOf(ctx serde.Context, rosterBuf []byte) (authority.Authority, error) {
	fakeAuthority := fakeAuthority{}
	return fakeAuthority, nil
}

type fakeAuthority struct {
	serde.Message
	serde.Fingerprinter
	crypto.CollectiveAuthority
}

func (f fakeAuthority) Apply(c authority.ChangeSet) authority.Authority {
	return nil
}

// Diff should return the change set to apply to get the given authority.
func (f fakeAuthority) Diff(a authority.Authority) authority.ChangeSet {
	return nil
}

func (f fakeAuthority) PublicKeyIterator() crypto.PublicKeyIterator {
	signers := make([]crypto.Signer, 2)
	signers[0] = fake.NewSigner()
	signers[1] = fakeCommonSigner

	return fake.NewPublicKeyIterator(signers)
}

func (f fakeAuthority) Len() int {
	return 0
}
