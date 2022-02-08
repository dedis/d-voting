package evoting

// todo: json marshall and unmarshall branch is are not covered yet

import (
	"crypto/sha256"
	"encoding/hex"
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
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/bls"
	"go.dedis.ch/dela/serde"
	sjson "go.dedis.ch/dela/serde/json"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	"go.dedis.ch/kyber/v3/util/random"
)

var dummyElectionIDBuff = []byte("dummyID")
var fakeElectionID = hex.EncodeToString(dummyElectionIDBuff)
var fakeCommonSigner = bls.NewSigner()

var ctx serde.Context

func init() {
	ctx = sjson.NewContext()
	ctx = serde.WithFactory(ctx, types.ElectionKey{}, types.ElectionFactory{})
	ctx = serde.WithFactory(ctx, types.CiphervoteKey{}, types.CiphervoteFactory{})
	ctx = serde.WithFactory(ctx, types.TransactionKey{}, types.TransactionFactory{})
}

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
	require.EqualError(t, err, "\"evoting:command\" not found in tx arg")

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

	createElection := types.CreateElection{
		AdminID: "dummyAdminID",
	}

	data, err := createElection.Serialize(ctx)
	require.NoError(t, err)

	var evotingAccessKey = [32]byte{3}
	rosterKey := [32]byte{}

	service := fakeAccess{err: fake.GetError()}
	rosterFac := fakeAuthorityFactory{}

	contract := NewContract(evotingAccessKey[:], rosterKey[:], service, fakeDkg, rosterFac)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.createElection(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, "failed to get transaction: \"evoting:arg\" not found in tx arg")

	err = cmd.createElection(fake.NewSnapshot(), makeStep(t, ElectionArg, "dummy"))
	require.EqualError(t, err, "failed to get transaction: failed to deserialize transaction: failed to decode: failed to unmarshal transaction json: invalid character 'd' looking for beginning of value")

	err = cmd.createElection(fake.NewBadSnapshot(), makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "failed to get roster")

	snap := fake.NewSnapshot()
	step := makeStep(t, ElectionArg, string(data))
	err = cmd.createElection(snap, step)
	require.NoError(t, err)

	// recover election ID:
	h := sha256.New()
	h.Write(step.Current.GetID())
	electionIDBuff := h.Sum(nil)

	res, err := snap.Get(electionIDBuff)
	require.NoError(t, err)

	fac := ctx.GetFactory(types.ElectionKey{})

	message, err := fac.Deserialize(ctx, res)
	require.NoError(t, err)

	election, ok := message.(types.Election)
	require.True(t, ok)

	require.Equal(t, createElection.AdminID, election.AdminID)
	require.Equal(t, types.Initial, election.Status)
}

func TestCommand_CastVote(t *testing.T) {
	castVote := types.CastVote{
		ElectionID: fakeElectionID,
		UserID:     "dummyUserId",
		Ballot: types.Ciphervote{types.EGPair{
			K: suite.Point(),
			C: suite.Point(),
		}},
	}

	data, err := castVote.Serialize(ctx)
	require.NoError(t, err)

	dummyElection, contract := initElectionAndContract()

	electionBuf, err := dummyElection.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.castVote(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, "failed to get transaction: \"evoting:arg\" not found in tx arg")

	err = cmd.castVote(fake.NewSnapshot(), makeStep(t, ElectionArg, "dummy"))
	require.EqualError(t, err, "failed to get transaction: failed to deserialize transaction: failed to decode: failed to unmarshal transaction json: invalid character 'd' looking for beginning of value")

	err = cmd.castVote(fake.NewBadSnapshot(), makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	err = snap.Set(dummyElectionIDBuff, []byte("fake election"))
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to deserialize Election")

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, fmt.Sprintf("the election is not open, current status: %d", types.Initial))

	dummyElection.Status = types.Open
	dummyElection.BallotSize = 0

	electionBuf, err = dummyElection.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "the ballot has unexpected length: 1 != 0")

	dummyElection.BallotSize = 29

	electionBuf, err = dummyElection.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	// encrypt a real message :
	RandomStream := suite.RandomStream()
	h := suite.Scalar().Pick(RandomStream)
	pubKey := suite.Point().Mul(h, nil)

	M := suite.Point().Embed([]byte("fakeVote"), random.New())

	// ElGamal-encrypt the point to produce ciphertext (K,C).
	k := suite.Scalar().Pick(random.New()) // ephemeral private key
	K := suite.Point().Mul(k, nil)         // ephemeral DH public key
	S := suite.Point().Mul(k, pubKey)      // ephemeral DH shared secret
	C := S.Add(S, M)                       // message blinded with secret

	castVote.Ballot = types.Ciphervote{
		types.EGPair{
			K: K,
			C: C,
		},
	}

	castVote.ElectionID = "X"

	data, err = castVote.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "failed to get election: failed to decode electionIDHex: encoding/hex: invalid byte: U+0058 'X'")

	dummyElection.ElectionID = fakeElectionID

	electionBuf, err = dummyElection.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	castVote.ElectionID = fakeElectionID

	data, err = castVote.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, ElectionArg, string(data)))
	require.NoError(t, err)

	res, err := snap.Get(dummyElectionIDBuff)
	require.NoError(t, err)

	fac := ctx.GetFactory(types.ElectionKey{})

	message, err := fac.Deserialize(ctx, res)
	require.NoError(t, err)

	election, ok := message.(types.Election)
	require.True(t, ok)

	require.Len(t, election.Suffragia.Ciphervotes, 1)
	require.True(t, castVote.Ballot.Equal(election.Suffragia.Ciphervotes[0]))

	require.Equal(t, castVote.UserID,
		election.Suffragia.UserIDs[0])
}

func TestCommand_CloseElection(t *testing.T) {
	closeElection := types.CloseElection{
		ElectionID: fakeElectionID,
		UserID:     "dummyUserId",
	}

	data, err := closeElection.Serialize(ctx)
	require.NoError(t, err)

	dummyElection, contract := initElectionAndContract()
	dummyElection.ElectionID = fakeElectionID

	electionBuf, err := dummyElection.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.closeElection(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, "failed to get transaction: \"evoting:arg\" not found in tx arg")

	err = cmd.closeElection(fake.NewSnapshot(), makeStep(t, ElectionArg, "dummy"))
	require.EqualError(t, err, "failed to get transaction: failed to deserialize transaction: failed to decode: failed to unmarshal transaction json: invalid character 'd' looking for beginning of value")

	err = cmd.closeElection(fake.NewBadSnapshot(), makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	err = snap.Set(dummyElectionIDBuff, []byte("fake election"))
	require.NoError(t, err)

	err = cmd.closeElection(snap, makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to deserialize Election")

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.closeElection(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "only the admin can close the election")

	closeElection.UserID = hex.EncodeToString([]byte("dummyAdminID"))

	data, err = closeElection.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.closeElection(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, fmt.Sprintf("the election is not open, current status: %d", types.Initial))

	dummyElection.Status = types.Open

	electionBuf, err = dummyElection.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.closeElection(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "at least two ballots are required")

	dummyElection.Suffragia.CastVote("dummyUser1", types.Ciphervote{})
	dummyElection.Suffragia.CastVote("dummyUser2", types.Ciphervote{})

	electionBuf, err = dummyElection.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.closeElection(snap, makeStep(t, ElectionArg, string(data)))
	require.NoError(t, err)

	res, err := snap.Get(dummyElectionIDBuff)
	require.NoError(t, err)

	fac := ctx.GetFactory(types.ElectionKey{})

	message, err := fac.Deserialize(ctx, res)
	require.NoError(t, err)

	election, ok := message.(types.Election)
	require.True(t, ok)

	require.Equal(t, types.Closed, election.Status)
}

func TestCommand_ShuffleBallotsCannotShuffleTwice(t *testing.T) {
	k := 3

	election, shuffleBallots, contract := initGoodShuffleBallot(t, k)

	cmd := evotingCommand{
		Contract: &contract,
		prover:   fakeProver,
	}

	snap := fake.NewSnapshot()

	// Attempts to shuffle twice :
	shuffleBallots.Round = 1

	election.ShuffleInstances = make([]types.ShuffleInstance, 1)
	election.ShuffleInstances[0].ShuffledBallots = make([]types.Ciphervote, 3)

	Ks, Cs, _ := fakeKCPoints(k)
	for i := 0; i < k; i++ {
		ballot := types.Ciphervote{types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		election.ShuffleInstances[0].ShuffledBallots[i] = ballot
	}

	election.ShuffleInstances[0].ShufflerPublicKey = shuffleBallots.PublicKey

	electionBuff, err := election.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuff)
	require.NoError(t, err)

	data, err := shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "a node already submitted a shuffle that has been accepted in round 0")
}

func TestCommand_ShuffleBallotsValidScenarios(t *testing.T) {
	k := 3

	// Simple Shuffle from round 0 :
	election, shuffleBallots, contract := initGoodShuffleBallot(t, k)

	cmd := evotingCommand{
		Contract: &contract,
		prover:   fakeProver,
	}

	snap := fake.NewSnapshot()

	electionBuf, err := election.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	data, err := shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	step := makeStep(t, ElectionArg, string(data))

	err = cmd.shuffleBallots(snap, step)
	require.NoError(t, err)

	// Valid Shuffle is over :
	shuffleBallots.Round = k
	election.ShuffleInstances = make([]types.ShuffleInstance, k)

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	for i := 1; i <= k-1; i++ {
		election.ShuffleInstances[i].ShuffledBallots = make([]types.Ciphervote, 3)
	}

	Ks, Cs, _ := fakeKCPoints(k)
	for i := 0; i < k; i++ {
		ballot := types.Ciphervote{types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		election.ShuffleInstances[k-1].ShuffledBallots[i] = ballot
	}

	electionBuf, err = election.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.NoError(t, err)

	// Check the shuffle is over:
	electionTxIDBuff, err := hex.DecodeString(election.ElectionID)
	require.NoError(t, err)

	electionBuf, err = snap.Get(electionTxIDBuff)
	require.NoError(t, err)

	fac := ctx.GetFactory(types.ElectionKey{})

	message, err := fac.Deserialize(ctx, electionBuf)
	require.NoError(t, err)

	election, ok := message.(types.Election)
	require.True(t, ok)

	require.Equal(t, election.Status, types.ShuffledBallots)
}

func TestCommand_ShuffleBallotsFormatErrors(t *testing.T) {
	k := 3

	election, shuffleBallots, contract := initBadShuffleBallot(k)

	data, err := shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	electionBuf, err := election.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
		prover:   fakeProver,
	}

	err = cmd.shuffleBallots(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, "failed to get transaction: \"evoting:arg\" not found in tx arg")

	err = cmd.shuffleBallots(fake.NewSnapshot(), makeStep(t, ElectionArg, "dummy"))
	require.EqualError(t, err, "failed to get transaction: failed to deserialize transaction: failed to decode: failed to unmarshal transaction json: invalid character 'd' looking for beginning of value")

	err = cmd.shuffleBallots(fake.NewBadSnapshot(), makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	// Wrong election id format
	snap := fake.NewSnapshot()

	err = snap.Set(dummyElectionIDBuff, []byte("fake election"))
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to deserialize Election")

	// Election not closed
	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "the election is not closed")

	// Wrong round :
	election.Status = types.Closed

	electionBuf, err = election.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "wrong shuffle round: expected round '0', transaction is for round '2'")

	// Missing public key of shuffler:
	shuffleBallots.Round = 1
	election.ShuffleInstances = make([]types.ShuffleInstance, 1)
	shuffleBallots.PublicKey = []byte("wrong Key")

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	electionBuf, err = election.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "could not verify identity of shuffler : public key not associated to a member of the roster: 77726f6e67204b6579")

	// Right key, wrong signature:
	shuffleBallots.PublicKey, _ = fakeCommonSigner.GetPublicKey().MarshalBinary()

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "could node deserialize shuffle signature : couldn't decode signature: couldn't deserialize data: unexpected end of JSON input")

	// Wrong election ID (Hash of shuffle fails)
	signature, err := fakeCommonSigner.Sign([]byte("fake shuffle"))
	require.NoError(t, err)

	wrongSignature, err := signature.Serialize(contract.context)
	require.NoError(t, err)

	shuffleBallots.Signature = wrongSignature

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	// Signatures not matching:
	election.ElectionID = fakeElectionID

	electionBuf, err = election.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "signature does not match the Shuffle : bls verify failed: bls: invalid signature")

	// Good format, signature not updated thus not matching, no random vector yet :
	Ks, Cs, pubKey := fakeKCPoints(k)

	for i := 0; i < k; i++ {
		ballot := types.Ciphervote{types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		shuffleBallots.ShuffledBallots[i] = ballot
	}

	h := sha256.New()

	err = shuffleBallots.Fingerprint(h)
	require.NoError(t, err)

	hash := h.Sum(nil)

	signature, _ = fakeCommonSigner.Sign(hash)
	wrongSignature, _ = signature.Serialize(contract.context)

	shuffleBallots.Signature = wrongSignature

	election.BallotSize = 1

	electionBuf, err = election.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "randomVector has unexpected length : 0 != 1")

	// random vector with right length, but different value :
	lenRandomVector := election.ChunksPerBallot()
	e := make([]kyber.Scalar, lenRandomVector)
	for i := 0; i < lenRandomVector; i++ {
		v := suite.Scalar().Pick(suite.RandomStream())
		e[i] = v
	}

	err = shuffleBallots.RandomVector.LoadFromScalars(e)
	require.NoError(t, err)

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "random vector from shuffle transaction is different than expected random vector")

	// generate correct random vector:
	h = sha256.New()

	err = shuffleBallots.Fingerprint(h)
	require.NoError(t, err)

	hash = h.Sum(nil)

	semiRandomStream, err := NewSemiRandomStream(hash)
	require.NoError(t, err)

	e = make([]kyber.Scalar, lenRandomVector)
	for i := 0; i < lenRandomVector; i++ {
		v := suite.Scalar().Pick(semiRandomStream)
		e[i] = v
	}

	err = shuffleBallots.RandomVector.LoadFromScalars(e)
	require.NoError(t, err)

	// > With no casted ballot the shuffling can't happen

	election.Pubkey = pubKey
	shuffleBallots.Round = 0
	election.ShuffleInstances = make([]types.ShuffleInstance, 0)
	election.Suffragia.Ciphervotes = make([]types.Ciphervote, 0)

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	electionBuf, err = election.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "not enough votes: 0 < 2")

	// > With only one shuffled ballot the shuffling can't happen

	election.Suffragia.CastVote("user1", types.Ciphervote{
		types.EGPair{K: suite.Point(), C: suite.Point()},
	})

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	electionBuf, err = election.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "not enough votes: 1 < 2")
}

func TestCommand_DecryptBallots(t *testing.T) {
	ballot1 := types.Ballot{}
	ballot2 := types.Ballot{}

	decryptBallot := types.DecryptBallots{
		ElectionID:       fakeElectionID,
		UserID:           hex.EncodeToString([]byte("dummyUserId")),
		DecryptedBallots: []types.Ballot{ballot1, ballot2},
	}

	data, err := decryptBallot.Serialize(ctx)
	require.NoError(t, err)

	dummyElection, contract := initElectionAndContract()
	dummyElection.ElectionID = fakeElectionID

	electionBuf, err := dummyElection.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.decryptBallots(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, "failed to get transaction: \"evoting:arg\" not found in tx arg")

	err = cmd.decryptBallots(fake.NewSnapshot(), makeStep(t, ElectionArg, "dummy"))
	require.EqualError(t, err, "failed to get transaction: failed to deserialize transaction: failed to decode: failed to unmarshal transaction json: invalid character 'd' looking for beginning of value")

	err = cmd.decryptBallots(fake.NewBadSnapshot(), makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	err = snap.Set(dummyElectionIDBuff, []byte("fake election"))
	require.NoError(t, err)

	err = cmd.decryptBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to deserialize Election")

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.decryptBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "only the admin can decrypt the ballots")

	decryptBallot.UserID = hex.EncodeToString([]byte("dummyAdminID"))

	data, err = decryptBallot.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.decryptBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, fmt.Sprintf("the ballots are not shuffled, current status: %d", types.Initial))

	dummyElection.Status = types.ShuffledBallots

	electionBuf, err = dummyElection.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.decryptBallots(snap, makeStep(t, ElectionArg, string(data)))
	require.NoError(t, err)

	res, err := snap.Get(dummyElectionIDBuff)
	require.NoError(t, err)

	fac := ctx.GetFactory(types.ElectionKey{})

	message, err := fac.Deserialize(ctx, res)
	require.NoError(t, err)

	election, ok := message.(types.Election)
	require.True(t, ok)

	require.Equal(t, decryptBallot.DecryptedBallots, election.DecryptedBallots)
	require.Equal(t, types.ResultAvailable, election.Status)

}

func TestCommand_CancelElection(t *testing.T) {
	cancelElection := types.CancelElection{
		ElectionID: fakeElectionID,
		UserID:     "dummyUserId",
	}

	data, err := cancelElection.Serialize(ctx)
	require.NoError(t, err)

	dummyElection, contract := initElectionAndContract()
	dummyElection.ElectionID = fakeElectionID

	electionBuf, err := dummyElection.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.cancelElection(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, "failed to get transaction: \"evoting:arg\" not found in tx arg")

	err = cmd.cancelElection(fake.NewSnapshot(), makeStep(t, ElectionArg, "dummy"))
	require.EqualError(t, err, "failed to get transaction: failed to deserialize transaction: failed to decode: failed to unmarshal transaction json: invalid character 'd' looking for beginning of value")

	err = cmd.cancelElection(fake.NewBadSnapshot(), makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	err = snap.Set(dummyElectionIDBuff, []byte("fake election"))
	require.NoError(t, err)

	err = cmd.cancelElection(snap, makeStep(t, ElectionArg, string(data)))
	require.Contains(t, err.Error(), "failed to deserialize Election")

	err = snap.Set(dummyElectionIDBuff, electionBuf)
	require.NoError(t, err)

	err = cmd.cancelElection(snap, makeStep(t, ElectionArg, string(data)))
	require.EqualError(t, err, "only the admin can cancel the election")

	cancelElection.UserID = hex.EncodeToString([]byte("dummyAdminID"))

	data, err = cancelElection.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.cancelElection(snap, makeStep(t, ElectionArg, string(data)))
	require.NoError(t, err)

	res, err := snap.Get(dummyElectionIDBuff)
	require.NoError(t, err)

	fac := ctx.GetFactory(types.ElectionKey{})

	message, err := fac.Deserialize(ctx, res)
	require.NoError(t, err)

	election, ok := message.(types.Election)
	require.True(t, ok)

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
	adminID := hex.EncodeToString([]byte("dummyAdminID"))

	dummyElection := types.Election{
		ElectionID:       fakeElectionID,
		AdminID:          adminID,
		Status:           0,
		Pubkey:           nil,
		Suffragia:        types.Suffragia{},
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

func initGoodShuffleBallot(t *testing.T, k int) (types.Election, types.ShuffleBallots, Contract) {
	election, shuffleBallots, contract := initBadShuffleBallot(3)
	election.Status = types.Closed

	election.BallotSize = 1

	Ks, Cs, pubKey := fakeKCPoints(k)
	shuffleBallots.PublicKey, _ = fakeCommonSigner.GetPublicKey().MarshalBinary()

	// ShuffledBallots:
	for i := 0; i < k; i++ {
		ballot := types.Ciphervote{types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		shuffleBallots.ShuffledBallots[i] = ballot
	}

	// Encrypted ballots:
	election.Pubkey = pubKey
	shuffleBallots.Round = 0
	election.ShuffleInstances = make([]types.ShuffleInstance, 0)

	for i := 0; i < k; i++ {
		ballot := types.Ciphervote{types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		election.Suffragia.CastVote(fmt.Sprintf("user%d", i), ballot)
	}

	// Valid Signature of shuffle
	election.ElectionID = fakeElectionID

	h := sha256.New()
	shuffleBallots.Fingerprint(h)
	hash := h.Sum(nil)
	signature, err := fakeCommonSigner.Sign(hash)
	require.NoError(t, err)

	wrongSignature, err := signature.Serialize(contract.context)
	require.NoError(t, err)

	shuffleBallots.Signature = wrongSignature

	semiRandomStream, err := NewSemiRandomStream(hash)
	require.NoError(t, err)

	lenRandomVector := election.ChunksPerBallot()
	e := make([]kyber.Scalar, lenRandomVector)
	for i := 0; i < lenRandomVector; i++ {
		v := suite.Scalar().Pick(semiRandomStream)
		e[i] = v
	}
	shuffleBallots.RandomVector.LoadFromScalars(e)

	return election, shuffleBallots, contract
}

func initBadShuffleBallot(sizeOfElection int) (types.Election, types.ShuffleBallots, Contract) {
	FakePubKey := fake.NewBadPublicKey()
	FakePubKeyMarshalled, _ := FakePubKey.MarshalBinary()
	shuffledBallots := make([]types.Ciphervote, sizeOfElection)

	shuffleBallots := types.ShuffleBallots{
		ElectionID:      fakeElectionID,
		Round:           2,
		ShuffledBallots: shuffledBallots,
		Proof:           nil,
		PublicKey:       FakePubKeyMarshalled,
	}

	election, contract := initElectionAndContract()

	return election, shuffleBallots, contract
}

func fakeKCPoints(k int) ([]kyber.Point, []kyber.Point, kyber.Point) {
	RandomStream := suite.RandomStream()
	h := suite.Scalar().Pick(RandomStream)
	pubKey := suite.Point().Mul(h, nil)

	Ks := make([]kyber.Point, 0, k)
	Cs := make([]kyber.Point, 0, k)

	for i := 0; i < k; i++ {
		// Embed the message into a curve point
		message := "Ballot" + strconv.Itoa(i)
		M := suite.Point().Embed([]byte(message), random.New())

		// ElGamal-encrypt the point to produce ciphertext (K,C).
		k := suite.Scalar().Pick(random.New()) // ephemeral private key
		K := suite.Point().Mul(k, nil)         // ephemeral DH public key
		S := suite.Point().Mul(k, pubKey)      // ephemeral DH shared secret
		C := S.Add(S, M)                       // message blinded with secret

		Ks = append(Ks, K)
		Cs = append(Cs, C)
	}
	return Ks, Cs, pubKey
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

func (f fakeDKG) Listen(electionID []byte) (dkg.Actor, error) {
	return f.actor, f.err
}

func (f fakeDKG) GetActor(electionID []byte) (dkg.Actor, bool) {
	return f.actor, false
}

func (f fakeDKG) SetService(service ordering.Service) {
}

type fakeDkgActor struct {
	publicKey kyber.Point
	err       error
}

func (f fakeDkgActor) Setup() (pubKey kyber.Point, err error) {
	return nil, f.err
}

func (f fakeDkgActor) GetPublicKey() (kyber.Point, error) {
	return f.publicKey, f.err
}

func (f fakeDkgActor) Encrypt(message []byte) (K, C kyber.Point, remainder []byte, err error) {
	return nil, nil, nil, f.err
}

func (f fakeDkgActor) Decrypt(K, C kyber.Point) ([]byte, error) {
	return nil, f.err
}

func (f fakeDkgActor) Reshare() error {
	return f.err
}

func (f fakeDkgActor) MarshalJSON() ([]byte, error) {
	return nil, f.err
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

func (c fakeCmd) openElection(snap store.Snapshot, step execution.Step) error {
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
