package evoting

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/d-voting/contracts/evoting/types"
	"go.dedis.ch/d-voting/internal/testing/fake"
	"go.dedis.ch/d-voting/services/dkg"
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

var dummyFormIDBuff = []byte("dummyID")
var fakeFormID = hex.EncodeToString(dummyFormIDBuff)
var fakeCommonSigner = bls.NewSigner()

const getTransactionErr = "failed to get transaction: \"evoting:arg\" not found in tx arg"
const unmarshalTransactionErr = "failed to get transaction: failed to deserialize " +
	"transaction: failed to decode: failed to unmarshal transaction json: invalid " +
	"character 'd' looking for beginning of value"
const deserializeErr = "failed to deserialize Form"

var invalidForm = []byte("fake form")

var ctx serde.Context

var formFac serde.Factory
var transactionFac serde.Factory

func init() {
	ciphervoteFac := types.CiphervoteFactory{}
	formFac = types.NewFormFactory(ciphervoteFac, fakeAuthorityFactory{})
	transactionFac = types.NewTransactionFactory(ciphervoteFac)

	ctx = sjson.NewContext()
}

func fakeProver(proof.Suite, string, proof.Verifier, []byte) error {
	return nil
}

func TestExecute(t *testing.T) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}
	service := fakeAccess{err: fake.GetError()}
	rosterFac := fakeAuthorityFactory{}

	contract := NewContract(service, fakeDkg, rosterFac)

	err := contract.Execute(fakeStore{}, makeStep(t))
	require.EqualError(t, err, "identity not authorized: fake.PublicKey ("+fake.GetError().Error()+")")

	service = fakeAccess{}

	contract = NewContract(service, fakeDkg, rosterFac)
	err = contract.Execute(fakeStore{}, makeStep(t))
	require.EqualError(t, err, "\"evoting:command\" not found in tx arg")

	contract.cmd = fakeCmd{err: fake.GetError()}

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCreateForm)))
	require.EqualError(t, err, fake.Err("failed to create form"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCastVote)))
	require.EqualError(t, err, fake.Err("failed to cast vote"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCloseForm)))
	require.EqualError(t, err, fake.Err("failed to close form"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdShuffleBallots)))
	require.EqualError(t, err, fake.Err("failed to shuffle ballots"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCombineShares)))
	require.EqualError(t, err, fake.Err("failed to decrypt ballots"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCancelForm)))
	require.EqualError(t, err, fake.Err("failed to cancel form"))

	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, "fake"))
	require.EqualError(t, err, "unknown command: fake")

	contract.cmd = fakeCmd{}
	err = contract.Execute(fakeStore{}, makeStep(t, CmdArg, string(CmdCreateForm)))
	require.NoError(t, err)

}

func TestCommand_CreateForm(t *testing.T) {
	initMetrics()

	fakeActor := fakeDkgActor{
		publicKey: suite.Point(),
		err:       nil,
	}

	fakeDkg := fakeDKG{
		actor: fakeActor,
		err:   nil,
	}

	createForm := types.CreateForm{
		AdminID: "dummyAdminID",
	}

	data, err := createForm.Serialize(ctx)
	require.NoError(t, err)

	service := fakeAccess{err: fake.GetError()}
	rosterFac := fakeAuthorityFactory{}

	contract := NewContract(service, fakeDkg, rosterFac)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.createForm(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, getTransactionErr)

	err = cmd.createForm(fake.NewSnapshot(), makeStep(t, FormArg, "dummy"))
	require.EqualError(t, err, unmarshalTransactionErr)

	err = cmd.createForm(fake.NewBadSnapshot(), makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "failed to get roster")

	snap := fake.NewSnapshot()
	step := makeStep(t, FormArg, string(data))
	err = cmd.createForm(snap, step)
	require.NoError(t, err)

	// recover form ID:
	h := sha256.New()
	h.Write(step.Current.GetID())
	formIDBuff := h.Sum(nil)

	res, err := snap.Get(formIDBuff)
	require.NoError(t, err)

	message, err := formFac.Deserialize(ctx, res)
	require.NoError(t, err)

	form, ok := message.(types.Form)
	require.True(t, ok)

	require.Equal(t, types.Initial, form.Status)
	require.Equal(t, float64(types.Initial), testutil.ToFloat64(PromFormStatus))
}

func TestCommand_OpenForm(t *testing.T) {
	// TODO
}

func TestCommand_CastVote(t *testing.T) {
	initMetrics()

	castVote := types.CastVote{
		FormID: fakeFormID,
		UserID: "dummyUserId",
		Ballot: types.Ciphervote{types.EGPair{
			K: suite.Point(),
			C: suite.Point(),
		}},
	}

	data, err := castVote.Serialize(ctx)
	require.NoError(t, err)

	dummyForm, contract := initFormAndContract()

	formBuf, err := dummyForm.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.castVote(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, getTransactionErr)

	err = cmd.castVote(fake.NewSnapshot(), makeStep(t, FormArg, "dummy"))
	require.EqualError(t, err, unmarshalTransactionErr)

	err = cmd.castVote(fake.NewBadSnapshot(), makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	err = snap.Set(dummyFormIDBuff, invalidForm)
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), deserializeErr)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, fmt.Sprintf("the form is not open, current status: %d", types.Initial))

	dummyForm.Status = types.Open
	dummyForm.BallotSize = 0

	formBuf, err = dummyForm.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "the ballot has unexpected length: 1 != 0")

	dummyForm.BallotSize = 29

	formBuf, err = dummyForm.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
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

	castVote.FormID = "X"

	data, err = castVote.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "failed to get form: failed to decode "+
		"formIDHex: encoding/hex: invalid byte: U+0058 'X'")

	dummyForm.FormID = fakeFormID

	formBuf, err = dummyForm.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	castVote.FormID = fakeFormID

	data, err = castVote.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.castVote(snap, makeStep(t, FormArg, string(data)))
	require.NoError(t, err)

	res, err := snap.Get(dummyFormIDBuff)
	require.NoError(t, err)

	message, err := formFac.Deserialize(ctx, res)
	require.NoError(t, err)

	form, ok := message.(types.Form)
	require.True(t, ok)

	require.Equal(t, uint32(1), form.BallotCount)
	suff, err := form.Suffragia(ctx, snap)
	require.NoError(t, err)
	require.True(t, castVote.Ballot.Equal(suff.Ciphervotes[0]))

	require.Equal(t, castVote.UserID, suff.UserIDs[0])
	require.Equal(t, float64(form.BallotCount), testutil.ToFloat64(PromFormBallots))
}

func TestCommand_CloseForm(t *testing.T) {
	initMetrics()

	closeForm := types.CloseForm{
		FormID: fakeFormID,
		UserID: "dummyUserId",
	}

	data, err := closeForm.Serialize(ctx)
	require.NoError(t, err)

	dummyForm, contract := initFormAndContract()
	dummyForm.FormID = fakeFormID

	formBuf, err := dummyForm.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.closeForm(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, getTransactionErr)

	err = cmd.closeForm(fake.NewSnapshot(), makeStep(t, FormArg, "dummy"))
	require.EqualError(t, err, unmarshalTransactionErr)

	err = cmd.closeForm(fake.NewBadSnapshot(), makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	err = snap.Set(dummyFormIDBuff, invalidForm)
	require.NoError(t, err)

	err = cmd.closeForm(snap, makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), deserializeErr)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	closeForm.UserID = hex.EncodeToString([]byte("dummyAdminID"))

	data, err = closeForm.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.closeForm(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, fmt.Sprintf("the form is not open, "+
		"current status: %d", types.Initial))
	require.Equal(t, 0, testutil.CollectAndCount(PromFormStatus))

	dummyForm.Status = types.Open

	formBuf, err = dummyForm.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.closeForm(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "at least two ballots are required")

	require.NoError(t, dummyForm.CastVote(ctx, snap, "dummyUser1", types.Ciphervote{}))
	require.NoError(t, dummyForm.CastVote(ctx, snap, "dummyUser2", types.Ciphervote{}))

	formBuf, err = dummyForm.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.closeForm(snap, makeStep(t, FormArg, string(data)))
	require.NoError(t, err)
	require.Equal(t, float64(types.Closed), testutil.ToFloat64(PromFormStatus))

	res, err := snap.Get(dummyFormIDBuff)
	require.NoError(t, err)

	message, err := formFac.Deserialize(ctx, res)
	require.NoError(t, err)

	form, ok := message.(types.Form)
	require.True(t, ok)

	require.Equal(t, types.Closed, form.Status)
	require.Equal(t, float64(types.Closed), testutil.ToFloat64(PromFormStatus))
}

func TestCommand_ShuffleBallotsCannotShuffleTwice(t *testing.T) {
	k := 3

	snap, form, shuffleBallots, contract := initGoodShuffleBallot(t, k)

	cmd := evotingCommand{
		Contract: &contract,
		prover:   fakeProver,
	}

	// Attempts to shuffle twice :
	shuffleBallots.Round = 1

	form.ShuffleInstances = make([]types.ShuffleInstance, 1)
	form.ShuffleInstances[0].ShuffledBallots = make([]types.Ciphervote, 3)

	Ks, Cs, _ := fakeKCPoints(k)
	for i := 0; i < k; i++ {
		ballot := types.Ciphervote{types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		form.ShuffleInstances[0].ShuffledBallots[i] = ballot
	}

	form.ShuffleInstances[0].ShufflerPublicKey = shuffleBallots.PublicKey

	formBuff, err := form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuff)
	require.NoError(t, err)

	data, err := shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "a node already submitted a shuffle that has "+
		"been accepted in round 0")
}

func TestCommand_ShuffleBallotsValidScenarios(t *testing.T) {
	initMetrics()

	k := 3

	// Simple Shuffle from round 0 :
	snap, form, shuffleBallots, contract := initGoodShuffleBallot(t, k)

	cmd := evotingCommand{
		Contract: &contract,
		prover:   fakeProver,
	}

	formBuf, err := form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	data, err := shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	step := makeStep(t, FormArg, string(data))

	err = cmd.shuffleBallots(snap, step)
	require.NoError(t, err)
	require.Equal(t, float64(1), testutil.ToFloat64(PromFormShufflingInstances))

	// Valid Shuffle is over :
	shuffleBallots.Round = k
	form.ShuffleInstances = make([]types.ShuffleInstance, k)

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	for i := 1; i <= k-1; i++ {
		form.ShuffleInstances[i].ShuffledBallots = make([]types.Ciphervote, 3)
	}

	Ks, Cs, _ := fakeKCPoints(k)
	for i := 0; i < k; i++ {
		ballot := types.Ciphervote{types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		form.ShuffleInstances[k-1].ShuffledBallots[i] = ballot
	}

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.NoError(t, err)
	require.Equal(t, float64(k+1), testutil.ToFloat64(PromFormShufflingInstances))

	// Check the shuffle is over:
	formTxIDBuff, err := hex.DecodeString(form.FormID)
	require.NoError(t, err)

	formBuf, err = snap.Get(formTxIDBuff)
	require.NoError(t, err)

	message, err := formFac.Deserialize(ctx, formBuf)
	require.NoError(t, err)

	form, ok := message.(types.Form)
	require.True(t, ok)

	require.Equal(t, types.ShuffledBallots, form.Status)
	require.Equal(t, float64(types.ShuffledBallots), testutil.ToFloat64(PromFormStatus))
}

func TestCommand_ShuffleBallotsFormatErrors(t *testing.T) {
	k := 3

	form, shuffleBallots, contract := initBadShuffleBallot(k)

	data, err := shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	formBuf, err := form.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
		prover:   fakeProver,
	}

	err = cmd.shuffleBallots(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, getTransactionErr)

	err = cmd.shuffleBallots(fake.NewSnapshot(), makeStep(t, FormArg, "dummy"))
	require.EqualError(t, err, unmarshalTransactionErr)

	err = cmd.shuffleBallots(fake.NewBadSnapshot(), makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	// Wrong form id format
	snap := fake.NewSnapshot()

	err = snap.Set(dummyFormIDBuff, invalidForm)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), deserializeErr)

	// Form not closed
	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "the form is not in state closed (current: 0 != closed: 2)")

	// Wrong round :
	form.Status = types.Closed

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "wrong shuffle round: expected round '0', "+
		"transaction is for round '2'")

	// Missing public key of shuffler:
	shuffleBallots.Round = 1
	form.ShuffleInstances = make([]types.ShuffleInstance, 1)
	shuffleBallots.PublicKey = []byte("wrong Key")

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "could not verify identity of shuffler : "+
		"public key not associated to a member of the roster: 77726f6e67204b6579")

	// Right key, wrong signature:
	shuffleBallots.PublicKey, _ = fakeCommonSigner.GetPublicKey().MarshalBinary()

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "could node deserialize shuffle signature : "+
		"couldn't decode signature: couldn't deserialize data: unexpected end of JSON input")

	// Wrong form ID (Hash of shuffle fails)
	signature, err := fakeCommonSigner.Sign([]byte("fake shuffle"))
	require.NoError(t, err)

	wrongSignature, err := signature.Serialize(contract.context)
	require.NoError(t, err)

	shuffleBallots.Signature = wrongSignature

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	// Signatures not matching:
	form.FormID = fakeFormID

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "signature does not match the Shuffle : "+
		"bls verify failed: bls: invalid signature")

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

	form.BallotSize = 1

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "randomVector has unexpected length : 0 != 1")

	// random vector with right length, but different value :
	lenRandomVector := form.ChunksPerBallot()
	e := make([]kyber.Scalar, lenRandomVector)
	for i := 0; i < lenRandomVector; i++ {
		v := suite.Scalar().Pick(suite.RandomStream())
		e[i] = v
	}

	err = shuffleBallots.RandomVector.LoadFromScalars(e)
	require.NoError(t, err)

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "random vector from shuffle transaction is "+
		"different than expected random vector")

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

	form.Pubkey = pubKey
	shuffleBallots.Round = 0
	form.ShuffleInstances = make([]types.ShuffleInstance, 0)

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "not enough votes: 0 < 2")

	// > With only one shuffled ballot the shuffling can't happen

	require.NoError(t, form.CastVote(ctx, snap, "user1", types.Ciphervote{
		types.EGPair{K: suite.Point(), C: suite.Point()},
	}))

	data, err = shuffleBallots.Serialize(ctx)
	require.NoError(t, err)

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.shuffleBallots(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "not enough votes: 1 < 2")
}

func TestCommand_RegisterPubShares(t *testing.T) {
	registerPubShares := types.RegisterPubShares{
		FormID:    fakeFormID,
		Index:     0,
		Pubshares: make([][]types.Pubshare, 0),
		Signature: []byte{},
		PublicKey: []byte{},
	}

	data, err := registerPubShares.Serialize(ctx)
	require.NoError(t, err)

	form, contract := initFormAndContract()
	form.FormID = fakeFormID

	formBuf, err := form.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.registerPubshares(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, getTransactionErr)

	err = cmd.registerPubshares(fake.NewSnapshot(), makeStep(t, FormArg, "dummy"))
	require.EqualError(t, err, unmarshalTransactionErr)

	err = cmd.registerPubshares(fake.NewBadSnapshot(), makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	err = snap.Set(dummyFormIDBuff, invalidForm)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), deserializeErr)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "the ballots have not been shuffled")

	// Requirements:
	form.Status = types.ShuffledBallots
	form.PubsharesUnits = types.PubsharesUnits{
		Pubshares: make([]types.PubsharesUnit, 0),
		PubKeys:   make([][]byte, 0),
		Indexes:   make([]int, 0),
	}
	form.ShuffleInstances = make([]types.ShuffleInstance, 1)
	form.ShuffleInstances[0] = types.ShuffleInstance{
		ShuffledBallots: make([]types.Ciphervote, 1),
	}
	form.ShuffleInstances[0].ShuffledBallots[0] = types.Ciphervote{types.EGPair{
		K: suite.Point(),
		C: suite.Point(),
	}}

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "could not verify identity of node : public key not associated to a member of the roster: ")

	registerPubShares.PublicKey, err = fakeCommonSigner.GetPublicKey().MarshalBinary()
	require.NoError(t, err)

	data, err = registerPubShares.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "could node deserialize pubShare signature:"+
		" couldn't decode signature: couldn't deserialize data: unexpected end of JSON input")

	signature, err := fakeCommonSigner.Sign([]byte("fake shares"))
	require.NoError(t, err)

	wrongSignature, err := signature.Serialize(ctx)
	require.NoError(t, err)

	registerPubShares.Signature = wrongSignature

	data, err = registerPubShares.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "signature does not match the PubsharesUnit:"+
		" bls verify failed: bls: invalid signature ")

	h := sha256.New()

	err = registerPubShares.Fingerprint(h)
	require.NoError(t, err)

	hash := h.Sum(nil)

	signature, err = fakeCommonSigner.Sign(hash)
	require.NoError(t, err)
	correctSignature, err := signature.Serialize(ctx)
	require.NoError(t, err)

	registerPubShares.Signature = correctSignature

	data, err = registerPubShares.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "unexpected size of pubshares submission: 0 != 1")

	registerPubShares.Pubshares = make([][]types.Pubshare, 1)

	data, err = registerPubShares.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "unexpected size of pubshares submission: 0 != 1")

	registerPubShares.Pubshares[0] = make([]types.Pubshare, 1)
	registerPubShares.Pubshares[0][0] = suite.Point()

	// update signature:

	h = sha256.New()

	err = registerPubShares.Fingerprint(h)
	require.NoError(t, err)

	hash = h.Sum(nil)

	signature, err = fakeCommonSigner.Sign(hash)
	require.NoError(t, err)
	correctSignature, err = signature.Serialize(ctx)
	require.NoError(t, err)

	registerPubShares.Signature = correctSignature

	data, err = registerPubShares.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.NoError(t, err)
	require.Equal(t, float64(1), testutil.ToFloat64(PromFormPubShares))

	// With the public key already used:
	form.PubsharesUnits.PubKeys = append(form.PubsharesUnits.PubKeys,
		registerPubShares.PublicKey)

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, fmt.Sprintf("'%x' already made a submission",
		registerPubShares.PublicKey))

	// With the index already used:
	form.PubsharesUnits.Indexes = append(form.PubsharesUnits.Indexes,
		registerPubShares.Index)
	form.PubsharesUnits.PubKeys = make([][]byte, 0)

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, "a submission has already been made for index 0")

	// All good:
	form.PubsharesUnits.Indexes = make([]int, 0)

	formBuf, err = form.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	err = cmd.registerPubshares(snap, makeStep(t, FormArg, string(data)))
	require.NoError(t, err)
	require.Equal(t, float64(1), testutil.ToFloat64(PromFormPubShares))

	res, err := snap.Get(dummyFormIDBuff)
	require.NoError(t, err)

	message, err := formFac.Deserialize(ctx, res)
	require.NoError(t, err)

	resultForm, ok := message.(types.Form)
	require.True(t, ok)

	require.Equal(t, types.PubSharesSubmitted, resultForm.Status)

	require.Equal(t, resultForm.PubsharesUnits.PubKeys[0], registerPubShares.PublicKey)
	require.Equal(t, resultForm.PubsharesUnits.Indexes[0], registerPubShares.Index)
}

func TestCommand_DecryptBallots(t *testing.T) {
	decryptBallot := types.CombineShares{
		FormID: fakeFormID,
		UserID: hex.EncodeToString([]byte("dummyUserId")),
	}

	data, err := decryptBallot.Serialize(ctx)
	require.NoError(t, err)

	dummyForm, contract := initFormAndContract()

	formBuf, err := dummyForm.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.combineShares(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, getTransactionErr)

	err = cmd.combineShares(fake.NewSnapshot(), makeStep(t, FormArg, "dummy"))
	require.EqualError(t, err, unmarshalTransactionErr)

	err = cmd.combineShares(fake.NewBadSnapshot(), makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	err = snap.Set(dummyFormIDBuff, invalidForm)
	require.NoError(t, err)

	err = cmd.combineShares(snap, makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), deserializeErr)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	decryptBallot.UserID = hex.EncodeToString([]byte("dummyAdminID"))

	data, err = decryptBallot.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.combineShares(snap, makeStep(t, FormArg, string(data)))
	require.EqualError(t, err, fmt.Sprintf("the public shares have not"+
		" been submitted, current status: %d", types.Initial))

	dummyForm.Status = types.PubSharesSubmitted

	// Avoid panic (will always be the case in practice):
	dummyForm.ShuffleInstances = make([]types.ShuffleInstance, 1)
	dummyForm.ShuffleInstances[0] = types.ShuffleInstance{
		ShuffledBallots:   make([]types.Ciphervote, 1),
		ShuffleProofs:     nil,
		ShufflerPublicKey: nil,
	}

	dummyForm.ShuffleInstances[0].ShuffledBallots[0] = types.Ciphervote{}

	formBuf, err = dummyForm.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	// Nothing to decrypt
	err = cmd.combineShares(snap, makeStep(t, FormArg, string(data)))
	require.NoError(t, err)

	dummyForm.ShuffleInstances[0].ShuffledBallots[0] = make([]types.EGPair, 1)
	dummyForm.ShuffleInstances[0].ShuffledBallots[0][0] = types.EGPair{
		K: suite.Point(),
		C: suite.Point(),
	}

	formBuf, err = dummyForm.Serialize(ctx)
	require.NoError(t, err)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	// Decrypt empty ballot
	err = cmd.combineShares(snap, makeStep(t, FormArg, string(data)))
	require.NoError(t, err)

	res, err := snap.Get(dummyFormIDBuff)
	require.NoError(t, err)

	message, err := formFac.Deserialize(ctx, res)
	require.NoError(t, err)

	form, ok := message.(types.Form)
	require.True(t, ok)

	require.Equal(t, types.Ballot{}, form.DecryptedBallots[0])
	require.Equal(t, types.ResultAvailable, form.Status)
	require.Equal(t, float64(types.ResultAvailable), testutil.ToFloat64(PromFormStatus))
}

func TestCommand_CancelForm(t *testing.T) {
	cancelForm := types.CancelForm{
		FormID: fakeFormID,
		UserID: "dummyUserId",
	}

	data, err := cancelForm.Serialize(ctx)
	require.NoError(t, err)

	dummyForm, contract := initFormAndContract()
	dummyForm.FormID = fakeFormID

	formBuf, err := dummyForm.Serialize(ctx)
	require.NoError(t, err)

	cmd := evotingCommand{
		Contract: &contract,
	}

	err = cmd.cancelForm(fake.NewSnapshot(), makeStep(t))
	require.EqualError(t, err, getTransactionErr)

	err = cmd.cancelForm(fake.NewSnapshot(), makeStep(t, FormArg, "dummy"))
	require.EqualError(t, err, unmarshalTransactionErr)

	err = cmd.cancelForm(fake.NewBadSnapshot(), makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), "failed to get key")

	snap := fake.NewSnapshot()

	err = snap.Set(dummyFormIDBuff, invalidForm)
	require.NoError(t, err)

	err = cmd.cancelForm(snap, makeStep(t, FormArg, string(data)))
	require.Contains(t, err.Error(), deserializeErr)

	err = snap.Set(dummyFormIDBuff, formBuf)
	require.NoError(t, err)

	cancelForm.UserID = hex.EncodeToString([]byte("dummyAdminID"))

	data, err = cancelForm.Serialize(ctx)
	require.NoError(t, err)

	err = cmd.cancelForm(snap, makeStep(t, FormArg, string(data)))
	require.NoError(t, err)

	res, err := snap.Get(dummyFormIDBuff)
	require.NoError(t, err)

	message, err := formFac.Deserialize(ctx, res)
	require.NoError(t, err)

	form, ok := message.(types.Form)
	require.True(t, ok)

	require.Equal(t, types.Canceled, form.Status)
	require.Equal(t, float64(types.Canceled), testutil.ToFloat64(PromFormStatus))
}

func TestRegisterContract(t *testing.T) {
	RegisterContract(native.NewExecution(), Contract{})
}

// -----------------------------------------------------------------------------
// Utility functions

func initMetrics() {
	PromFormStatus.Reset()
	PromFormBallots.Reset()
	PromFormShufflingInstances.Reset()
	PromFormPubShares.Reset()
}

func initFormAndContract() (types.Form, Contract) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}

	dummyForm := types.Form{
		FormID:           fakeFormID,
		Status:           0,
		Pubkey:           nil,
		ShuffleInstances: make([]types.ShuffleInstance, 0),
		DecryptedBallots: nil,
		ShuffleThreshold: 0,
		Roster:           fake.Authority{},
	}

	service := fakeAccess{err: fake.GetError()}
	rosterFac := fakeAuthorityFactory{}

	contract := NewContract(service, fakeDkg, rosterFac)

	return dummyForm, contract
}

func initGoodShuffleBallot(t *testing.T, k int) (store.Snapshot, types.Form, types.ShuffleBallots, Contract) {
	form, shuffleBallots, contract := initBadShuffleBallot(3)
	form.Status = types.Closed

	form.BallotSize = 1

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
	form.Pubkey = pubKey
	shuffleBallots.Round = 0
	form.ShuffleInstances = make([]types.ShuffleInstance, 0)

	snap := fake.NewSnapshot()
	for i := 0; i < k; i++ {
		ballot := types.Ciphervote{types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		require.NoError(t, form.CastVote(ctx, snap, fmt.Sprintf("user%d", i), ballot))
	}

	// Valid Signature of shuffle
	form.FormID = fakeFormID

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

	lenRandomVector := form.ChunksPerBallot()
	e := make([]kyber.Scalar, lenRandomVector)
	for i := 0; i < lenRandomVector; i++ {
		v := suite.Scalar().Pick(semiRandomStream)
		e[i] = v
	}
	shuffleBallots.RandomVector.LoadFromScalars(e)

	return snap, form, shuffleBallots, contract
}

func initBadShuffleBallot(sizeOfForm int) (types.Form, types.ShuffleBallots, Contract) {
	FakePubKey := fake.NewBadPublicKey()
	FakePubKeyMarshalled, _ := FakePubKey.MarshalBinary()
	shuffledBallots := make([]types.Ciphervote, sizeOfForm)

	shuffleBallots := types.ShuffleBallots{
		FormID:          fakeFormID,
		Round:           2,
		ShuffledBallots: shuffledBallots,
		Proof:           nil,
		PublicKey:       FakePubKeyMarshalled,
	}

	form, contract := initFormAndContract()

	return form, shuffleBallots, contract
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

func (f fakeDKG) Listen(formID []byte, txmanager txn.Manager) (dkg.Actor, error) {
	return f.actor, f.err
}

func (f fakeDKG) GetActor(formID []byte) (dkg.Actor, bool) {
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

func (f fakeDkgActor) ComputePubshares() error {
	return f.err
}

func (f fakeDkgActor) Status() dkg.Status {
	return dkg.Status{}
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

func (c fakeCmd) createForm(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) openForm(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) castVote(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) closeForm(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) shuffleBallots(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) combineShares(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) cancelForm(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) deleteForm(snap store.Snapshot, step execution.Step) error {
	return c.err
}

func (c fakeCmd) registerPubshares(snap store.Snapshot, step execution.Step) error {
	return c.err
}

type fakeAuthorityFactory struct {
	serde.Factory
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

func (fakeAuthority) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
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
