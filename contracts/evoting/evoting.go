// Code generated ...

package evoting

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"strings"

	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering/cosipbft/contracts/viewchange"
	"go.dedis.ch/kyber/v3/share"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/cosi/threshold"
	"go.dedis.ch/dela/crypto/bls"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3/proof"
	"go.dedis.ch/kyber/v3/shuffle"
	"golang.org/x/xerrors"
)

const (
	shufflingProtocolName = "PairShuffle"
	errGetTransaction     = "failed to get transaction: %v"
	errGetForm            = "failed to get form: %v"
	errIsRole             = "failed check the permission: %v"
	errNoOwnerPerms       = "The user %v doesn't have the Owner permission on the form."
	errNoVoterPerms       = "The user %v doesn't have the Voter permission on the form."
	errWrongTx            = "wrong type of transaction: %T"
)

// evotingCommand implements the commands of the Evoting contract.
//
// - implements commands
type evotingCommand struct {
	*Contract

	prover prover
}

type Role int

const (
	Voters Role = iota + 1
	Owners
)

type prover func(suite proof.Suite, protocolName string, verifier proof.Verifier, proof []byte) error

// createForm implements commands. It performs the CREATE_FORM command
func (e evotingCommand) createForm(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.CreateForm)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	rosterBuf, err := snap.Get(viewchange.GetRosterKey())
	if err != nil {
		return xerrors.Errorf("failed to get roster")
	}

	// Check if has Admin Right to create a form
	isAdmin, _, err := e.fetchAdmin(snap, tx.UserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return xerrors.Errorf("The performing user is not an admin.")
	}

	roster, err := e.rosterFac.AuthorityOf(e.context, rosterBuf)
	if err != nil {
		return xerrors.Errorf("failed to get roster: %v", err)
	}

	// Get the formID, which is the SHA256 of the transaction ID
	h := sha256.New()
	h.Write(step.Current.GetID())
	formIDBuf := h.Sum(nil)

	if !tx.Configuration.IsValid() {
		return xerrors.Errorf("configuration of form is incoherent or has duplicated IDs")
	}

	units := types.PubsharesUnits{
		Pubshares: make([]types.PubsharesUnit, 0),
		PubKeys:   make([][]byte, 0),
		Indexes:   make([]int, 0),
	}

	// Initial owner is the creator
	owners := make([]int, 1)

	sciperInt, err := types.SciperToInt(tx.UserID)
	if err != nil {
		return xerrors.Errorf("failed to convert SCIPER to integer: %v", err)
	}

	owners[0] = sciperInt

	form := types.Form{
		FormID:        hex.EncodeToString(formIDBuf),
		Configuration: tx.Configuration,
		Status:        types.Initial,
		// Pubkey is set by the opening command
		BallotSize:       tx.Configuration.MaxBallotSize(),
		PubsharesUnits:   units,
		ShuffleInstances: []types.ShuffleInstance{},
		DecryptedBallots: []types.Ballot{},
		// We set the participant in the e-voting once for all. If it happens
		// that 1/3 of the participants go away, the form will never end.
		Roster:           roster,
		ShuffleThreshold: threshold.ByzantineThreshold(roster.Len()),
		Owners:           owners,
		Voters:           make([]int, 0),
	}

	PromFormStatus.WithLabelValues(form.FormID).Set(float64(form.Status))

	formBuf, err := form.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form : %v", err)
	}

	err = snap.Set(formIDBuf, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	err = updateFormMetadataStore(snap, form.FormID)
	if err != nil {
		return xerrors.Errorf("failed to update the metadata in the store: %v", err)
	}

	return nil
}

// updateFormMetadataStore Update the form metadata store
func updateFormMetadataStore(snap store.Snapshot, formID string) error {
	formsMetadataBuf, err := snap.Get([]byte(FormsMetadataKey))
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", formsMetadataBuf, err)
	}

	formsMetadata := &types.FormsMetadata{
		FormsIDs: types.FormIDs{},
	}

	if len(formsMetadataBuf) != 0 {
		err := json.Unmarshal(formsMetadataBuf, formsMetadata)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal FormsMetadata: %v", err)
		}
	}

	err = formsMetadata.FormsIDs.Add(formID)
	if err != nil {
		return xerrors.Errorf("couldn't add new form: %v", err)
	}

	formMetadataJSON, err := json.Marshal(formsMetadata)
	if err != nil {
		return xerrors.Errorf("failed to marshal FormsMetadata: %v", err)
	}

	err = snap.Set([]byte(FormsMetadataKey), formMetadataJSON)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}
	return nil
}

// openForm set the public key on the form. The public key is fetched
// from the DKG actor. It works only if DKG is set up.
func (e evotingCommand) openForm(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.OpenForm)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	form, formID, err := e.getForm(tx.FormID, snap)
	if err != nil {
		return xerrors.Errorf(errGetForm, err)
	}

	isOwner, err := e.isRole(form, tx.UserID, Owners)
	if err != nil {
		return xerrors.Errorf(errIsRole, err)
	}

	if !isOwner {
		return xerrors.Errorf(errNoOwnerPerms, tx.UserID)
	}

	if form.Status != types.Initial {
		return xerrors.Errorf("the form was opened before, current status: %d", form.Status)
	}

	form.Status = types.Open
	PromFormStatus.WithLabelValues(form.FormID).Set(float64(form.Status))

	if form.Pubkey != nil {
		return xerrors.Errorf("pubkey is already set: %s", form.Pubkey)
	}

	dkgActor, exists := e.pedersen.GetActor(formID)
	if !exists {
		return xerrors.Errorf("failed to get actor for form %q", form.FormID)
	}

	pubkey, err := dkgActor.GetPublicKey()
	if err != nil {
		return xerrors.Errorf("failed to get pubkey: %v", err)
	}

	form.Pubkey = pubkey

	formBuf, err := form.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form : %v", err)
	}

	err = snap.Set(formID, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// castVote implements commands. It performs the CAST_VOTE command
func (e evotingCommand) castVote(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.CastVote)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	form, formID, err := e.getForm(tx.FormID, snap)
	if err != nil {
		return xerrors.Errorf(errGetForm, err)
	}

	if form.Status != types.Open {
		return xerrors.Errorf("the form is not open, current status: %d", form.Status)
	}

	isOwner, err := e.isRole(form, tx.VoterID, Voters)
	if err != nil {
		return xerrors.Errorf(errIsRole, err)
	}

	if !isOwner {
		return xerrors.Errorf(errNoVoterPerms, tx.VoterID)
	}

	if len(tx.Ballot) != form.ChunksPerBallot() {
		return xerrors.Errorf("the ballot has unexpected length: %d != %d",
			len(tx.Ballot), form.ChunksPerBallot())
	}

	err = form.CastVote(e.context, snap, tx.VoterID, tx.Ballot)
	if err != nil {
		return xerrors.Errorf("couldn't cast vote: %v", err)
	}

	formBuf, err := form.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form : %v", err)
	}

	err = snap.Set(formID, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	PromFormBallots.WithLabelValues(form.FormID).Set(float64(form.BallotCount))

	return nil
}

// shuffleBallots implements commands. It performs the SHUFFLE_BALLOTS command
func (e evotingCommand) shuffleBallots(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.ShuffleBallots)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	err = e.checkPreviousTransactions(step, tx.Round)
	if err != nil {
		return xerrors.Errorf("check previous transactions failed: %v", err)
	}

	form, formID, err := e.getForm(tx.FormID, snap)
	if err != nil {
		return xerrors.Errorf(errGetForm, err)
	}

	if form.Status != types.Closed {
		return xerrors.Errorf("the form is not in state closed (current: %d != closed: %d)",
			form.Status, types.Closed)
	}

	isOwner, err := e.isRole(form, tx.UserID, Owners)
	if err != nil {
		return xerrors.Errorf(errIsRole, err)
	}

	if !isOwner {
		return xerrors.Errorf(errNoOwnerPerms, tx.UserID)
	}

	// Round starts at 0
	expectedRound := len(form.ShuffleInstances)

	if tx.Round != expectedRound {
		return xerrors.Errorf("wrong shuffle round: expected round '%d', "+
			"transaction is for round '%d'", expectedRound, tx.Round)
	}

	shufflerPublicKey := tx.PublicKey

	err = isMemberOf(form.Roster, shufflerPublicKey)
	if err != nil {
		return xerrors.Errorf("could not verify identity of shuffler : %v", err)
	}

	// Check the node who submitted the shuffle did not already submit an
	// accepted shuffle
	for i, shuffleInstance := range form.ShuffleInstances {
		if bytes.Equal(shufflerPublicKey, shuffleInstance.ShufflerPublicKey) {
			return xerrors.Errorf("a node already submitted a shuffle that "+
				"has been accepted in round %d", i)
		}
	}

	// Check the shuffler indeed signed the transaction:
	signerPubKey, err := bls.NewPublicKey(tx.PublicKey)
	if err != nil {
		return xerrors.Errorf("could not decode public key of signer : %v ", err)
	}

	txSignature := tx.Signature

	signature, err := bls.NewSignatureFactory().SignatureOf(e.context, txSignature)
	if err != nil {
		return xerrors.Errorf("could node deserialize shuffle signature : %v", err)
	}

	h := sha256.New()

	err = tx.Fingerprint(h)
	if err != nil {
		return xerrors.Errorf("failed to get fingerprint: %v", err)
	}

	hash := h.Sum(nil)

	// Check the signature matches the shuffle using the shuffler's public key
	err = signerPubKey.Verify(hash, signature)
	if err != nil {
		return xerrors.Errorf("signature does not match the Shuffle : %v", err)
	}

	// Retrieve the random vector (ie the Scalar vector)
	randomVector, err := tx.RandomVector.Unmarshal()
	if err != nil {
		return xerrors.Errorf("failed to unmarshal random vector: %v", err)
	}

	// Check that the random vector is correct
	semiRandomStream, err := NewSemiRandomStream(hash)
	if err != nil {
		return xerrors.Errorf("could not create semi-random stream: %v", err)
	}

	if form.ChunksPerBallot() != len(randomVector) {
		return xerrors.Errorf("randomVector has unexpected length : %v != %v",
			len(randomVector), form.ChunksPerBallot())
	}

	for i := 0; i < form.ChunksPerBallot(); i++ {
		v := suite.Scalar().Pick(semiRandomStream)
		if !randomVector[i].Equal(v) {
			return xerrors.Errorf("random vector from shuffle transaction is " +
				"different than expected random vector")
		}
	}

	if len(tx.ShuffledBallots) == 0 {
		return xerrors.Errorf("there are no shuffled ballots")
	}

	XX, YY := types.CiphervotesToPairs(tx.ShuffledBallots)

	var ciphervotes []types.Ciphervote

	if tx.Round == 0 {
		suff, err := form.Suffragia(e.context, snap)
		if err != nil {
			return xerrors.Errorf("couldn't get ballots: %v", err)
		}
		ciphervotes = suff.Ciphervotes
	} else {
		// get the form's last shuffled ballots
		lastIndex := len(form.ShuffleInstances) - 1
		ciphervotes = form.ShuffleInstances[lastIndex].ShuffledBallots
	}

	if len(ciphervotes) < 2 {
		return xerrors.Errorf("not enough votes: %d < 2", len(ciphervotes))
	}

	X, Y := types.CiphervotesToPairs(ciphervotes)

	XXUp, YYUp, XXDown, YYDown := shuffle.GetSequenceVerifiable(suite, X, Y, XX,
		YY, randomVector)

	verifier := shuffle.Verifier(suite, nil, form.Pubkey, XXUp, YYUp, XXDown, YYDown)

	err = e.prover(suite, shufflingProtocolName, verifier, tx.Proof)
	if err != nil {
		return xerrors.Errorf("proof verification failed: %v", err)
	}

	// append the new shuffled ballots and the proof to the lists
	currentShuffleInstance := types.ShuffleInstance{
		ShuffledBallots:   tx.ShuffledBallots,
		ShuffleProofs:     tx.Proof,
		ShufflerPublicKey: shufflerPublicKey,
	}

	form.ShuffleInstances = append(form.ShuffleInstances, currentShuffleInstance)

	PromFormShufflingInstances.WithLabelValues(form.FormID).Set(float64(len(form.ShuffleInstances)))

	// in case we have enough shuffled ballots, we update the status
	if len(form.ShuffleInstances) >= form.ShuffleThreshold {
		form.Status = types.ShuffledBallots
		PromFormStatus.WithLabelValues(form.FormID).Set(float64(form.Status))
	}

	formBuf, err := form.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form : %v", err)
	}

	err = snap.Set(formID, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// checkPreviousTransactions checks if a ShuffleBallotsTransaction has already
// been accepted and executed for a specific round.
func (e evotingCommand) checkPreviousTransactions(step execution.Step, round int) error {
	for _, tx := range step.Previous {
		// skip tx not concerning the evoting contract
		if string(tx.GetArg(native.ContractArg)) != ContractName {
			continue
		}

		// skip tx that does not contain the form argument
		if string(tx.GetArg(CmdArg)) != FormArg {
			continue
		}

		msg, err := e.getTransaction(step.Current)
		if err != nil {
			return xerrors.Errorf(errGetTransaction, err)
		}

		// skip if not a shuffling ballot transaction
		shuffleBallots, ok := msg.(types.ShuffleBallots)
		if !ok {
			continue
		}

		if shuffleBallots.Round == round {
			return xerrors.Errorf("shuffle is already happening in this round")
		}
	}

	return nil
}

// closeForm implements commands. It performs the CLOSE_FORM command
func (e evotingCommand) closeForm(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.CloseForm)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	form, formID, err := e.getForm(tx.FormID, snap)
	if err != nil {
		return xerrors.Errorf(errGetForm, err)
	}

	if form.Status != types.Open {
		return xerrors.Errorf("the form is not open, current status: %d", form.Status)
	}

	isOwner, err := e.isRole(form, tx.UserID, Owners)
	if err != nil {
		return xerrors.Errorf(errIsRole, err)
	}

	if !isOwner {
		return xerrors.Errorf(errNoOwnerPerms, tx.UserID)
	}

	if form.BallotCount <= 1 {
		return xerrors.Errorf("at least two ballots are required")
	}

	form.Status = types.Closed
	PromFormStatus.WithLabelValues(form.FormID).Set(float64(form.Status))

	formBuf, err := form.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form : %v", err)
	}

	err = snap.Set(formID, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// registerPubshares implements commands. It performs the
// REGISTER_PUB_SHARES command
func (e evotingCommand) registerPubshares(snap store.Snapshot, step execution.Step) error {
	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.RegisterPubShares)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	form, formID, err := e.getForm(tx.FormID, snap)
	if err != nil {
		return xerrors.Errorf(errGetForm, err)
	}

	if form.Status != types.ShuffledBallots {
		return xerrors.Errorf("the ballots have not been shuffled")
	}

	err = isMemberOf(form.Roster, tx.PublicKey)
	if err != nil {
		return xerrors.Errorf("could not verify identity of node : %v", err)
	}

	signerPubKey, err := bls.NewPublicKey(tx.PublicKey)
	if err != nil {
		return xerrors.Errorf("could not recover public key from tx: %v", err)
	}

	// Check the node indeed signed the transaction:
	txSignature := tx.Signature

	signature, err := bls.NewSignatureFactory().SignatureOf(e.context, txSignature)
	if err != nil {
		return xerrors.Errorf("could node deserialize pubShare signature: %v", err)
	}

	h := sha256.New()

	err = tx.Fingerprint(h)
	if err != nil {
		return xerrors.Errorf("failed to get fingerprint: %v", err)
	}

	hash := h.Sum(nil)

	// Check the signature matches the pubshares using the node's public key
	err = signerPubKey.Verify(hash, signature)
	if err != nil {
		return xerrors.Errorf("signature does not match the PubsharesUnit: %v ", err)
	}

	// coherence check on the length of the shares submitted
	shuffledBallots := form.ShuffleInstances[len(form.ShuffleInstances)-1].
		ShuffledBallots
	if len(tx.Pubshares) != len(shuffledBallots) {
		return xerrors.Errorf("unexpected size of pubshares submission: %d != %d",
			len(tx.Pubshares), len(shuffledBallots))
	}

	for i, ballot := range shuffledBallots {
		if len(ballot) != len(tx.Pubshares[i]) {
			return xerrors.Errorf("unexpected size of pubshares submission: %d != %d",
				len(tx.Pubshares[i]), len(ballot))
		}
	}

	units := &form.PubsharesUnits

	// Check the node hasn't made any other submissions
	for _, key := range units.PubKeys {
		if bytes.Equal(key, tx.PublicKey) {
			return xerrors.Errorf("'%x' already made a submission", key)
		}
	}

	for _, index := range units.Indexes {
		if index == tx.Index {
			return xerrors.Errorf("a submission has already been made for index %d", index)
		}
	}

	// Add the pubshares to the form
	units.Pubshares = append(units.Pubshares, tx.Pubshares)
	units.PubKeys = append(units.PubKeys, tx.PublicKey)
	units.Indexes = append(units.Indexes, tx.Index)

	nbrSubmissions := len(units.Pubshares)

	PromFormPubShares.WithLabelValues(form.FormID).Set(float64(nbrSubmissions))

	if nbrSubmissions >= form.ShuffleThreshold {
		form.Status = types.PubSharesSubmitted
		PromFormStatus.WithLabelValues(form.FormID).Set(float64(form.Status))
	}

	formBuf, err := form.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form: %v", err)
	}

	err = snap.Set(formID, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// combineShares implements commands. It performs the COMBINE_SHARES command
func (e evotingCommand) combineShares(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.CombineShares)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	form, formID, err := e.getForm(tx.FormID, snap)
	if err != nil {
		return xerrors.Errorf(errGetForm, err)
	}

	if form.Status != types.PubSharesSubmitted {
		return xerrors.Errorf("the public shares have not been submitted,"+
			" current status: %d", form.Status)
	}

	isOwner, err := e.isRole(form, tx.UserID, Owners)
	if err != nil {
		return xerrors.Errorf(errIsRole, err)
	}

	if !isOwner {
		return xerrors.Errorf(errNoOwnerPerms, tx.UserID)
	}

	allPubShares := form.PubsharesUnits.Pubshares

	shufflesSize := len(form.ShuffleInstances)

	shuffledBallotsSize := len(form.ShuffleInstances[shufflesSize-1].ShuffledBallots)
	ballotSize := len(form.ShuffleInstances[shufflesSize-1].ShuffledBallots[0])

	decryptedBallots := make([]types.Ballot, shuffledBallotsSize)

	for i := 0; i < shuffledBallotsSize; i++ {
		// decryption of one ballot:
		marshalledBallot := strings.Builder{}

		for j := 0; j < ballotSize; j++ {
			chunk, err := decrypt(i, j, allPubShares, form.PubsharesUnits.Indexes)
			if err != nil {
				return xerrors.Errorf("failed to decrypt (K, C): %v", err)
			}

			marshalledBallot.Write(chunk)
		}

		var ballot types.Ballot
		err = ballot.Unmarshal(marshalledBallot.String(), form)

		if err != nil {
			dela.Logger.Warn().Msgf("Failed to unmarshal a ballot: %v", err)
		}

		decryptedBallots[i] = ballot
	}

	form.DecryptedBallots = decryptedBallots

	form.Status = types.ResultAvailable
	PromFormStatus.WithLabelValues(form.FormID).Set(float64(form.Status))

	formBuf, err := form.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form : %v", err)
	}

	err = snap.Set(formID, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// cancelForm implements commands. It performs the CANCEL_FORM command
func (e evotingCommand) cancelForm(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.CancelForm)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	form, formID, err := e.getForm(tx.FormID, snap)
	if err != nil {
		return xerrors.Errorf(errGetForm, err)
	}

	isOwner, err := e.isRole(form, tx.UserID, Owners)
	if err != nil {
		return xerrors.Errorf(errIsRole, err)
	}

	if !isOwner {
		return xerrors.Errorf(errNoOwnerPerms, tx.UserID)
	}

	form.Status = types.Canceled
	PromFormStatus.WithLabelValues(form.FormID).Set(float64(form.Status))

	formBuf, err := form.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form : %v", err)
	}

	err = snap.Set(formID, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// deleteForm implements commands. It performs the DELETE_FORM command
func (e evotingCommand) deleteForm(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.DeleteForm)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	form, formID, err := e.getForm(tx.FormID, snap)
	if err != nil {
		return xerrors.Errorf(errGetForm, err)
	}

	isOwner, err := e.isRole(form, tx.UserID, Owners)
	if err != nil {
		return xerrors.Errorf(errIsRole, err)
	}

	if !isOwner {
		return xerrors.Errorf(errNoOwnerPerms, tx.UserID)
	}

	err = snap.Delete(formID)
	if err != nil {
		return xerrors.Errorf("failed to delete form: %v", err)
	}

	err = updateFormMetadataStore(snap, form.FormID)
	if err != nil {
		return xerrors.Errorf("failed to update the metadata in the store: %v", err)
	}

	return nil
}

// manageAdminList implements commands. It performs the ADD or REMOVE ADMIN command
func (e evotingCommand) manageAdminList(snap store.Snapshot, step execution.Step) error {
	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	var list types.AdminList

	h := sha256.New()
	h.Write([]byte(AdminListId))
	formIDBuf := h.Sum(nil)

	txAddAdmin, okAddAdmin := msg.(types.AddAdmin)
	txRemoveAdmin, okRemoveAdmin := msg.(types.RemoveAdmin)

	if okAddAdmin {
		isAdmin, listRetrieved, err := e.fetchAdmin(snap, txAddAdmin.PerformingUserID)
		list = listRetrieved
		if err != nil {
			// Exact string matching of the error
			if err.Error() != "failed to get the AdminList: No list found" {
				return xerrors.Errorf("failed to get AdminList: %v", err)
			}

			// Trust On First Use System -> if no AdminList, will create one by default.

			intSciper, err := types.SciperToInt(txAddAdmin.TargetUserID)
			if err != nil {
				return xerrors.Errorf("Invalid Sciper: %v", err)
			}

			err = initializeAdminList(snap, intSciper, e.context)
			if err != nil {
				return xerrors.Errorf("Failed to initialize admin list: %v", err)
			}

			// Adding the initial admin is performed by the initialize Admin List
			// Therefore return
			return nil
		}
		if !isAdmin {
			return xerrors.Errorf("The performing user is not an admin.")
		}

		err = list.AddAdmin(txAddAdmin.TargetUserID)
		if err != nil {
			return xerrors.Errorf("couldn't add admin: %v", err)
		}
	} else if okRemoveAdmin {
		isAdmin, listRetrieved, err := e.fetchAdmin(snap, txRemoveAdmin.PerformingUserID)
		list = listRetrieved
		if err != nil {
			return err
		}
		if !isAdmin {
			return xerrors.Errorf("The performing user is not an admin.")
		}

		err = list.RemoveAdmin(txRemoveAdmin.TargetUserID)
		if err != nil {
			return xerrors.Errorf("couldn't remove admin: %v", err)
		}
	} else {
		return xerrors.Errorf(errWrongTx, msg)
	}

	formBuf, err := list.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form : %v", err)
	}

	err = snap.Set(formIDBuf, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// isRole check whether the txPerformingUser has the role in the provided form
func (e evotingCommand) isRole(form types.Form, txPerformingUser string, role Role) (bool, error) {
	sciperInt, err := types.SciperToInt(txPerformingUser)
	if err != nil {
		return false, xerrors.Errorf("Failed to convert SCIPER to int: %v", err)
	}

	if role == Voters {
		for i := 0; i < len(form.Voters); i++ {
			if form.Voters[i] == sciperInt {
				return true, nil
			}
		}
	} else if role == Owners {
		for i := 0; i < len(form.Owners); i++ {
			if form.Owners[i] == sciperInt {
				return true, nil
			}
		}
	}

	return false, nil
}

// fetchAdmin Check whether a user is in an Admin List
func (e evotingCommand) fetchAdmin(snap store.Snapshot, txPerformingUser string) (bool, types.AdminList, error) {
	// If it found the AdminList
	// Check that the performing user is Admin
	form, err := e.getAdminList(snap)
	if err != nil {
		return false, types.AdminList{}, err
	}

	performingUserPerm, err := form.GetAdminIndex(txPerformingUser)
	if err != nil {
		return false, form, xerrors.Errorf("couldn't retrieve admin permission of the performing user: %v", err)
	}

	if performingUserPerm < 0 {
		return false, form, nil
	}
	return true, form, nil
}

// initializeAdminList initialize an AdminList on the blockchain. It is called the first time that
// we attempt to add an admin.
func initializeAdminList(snap store.Snapshot, initialAdmin int, ctx serde.Context) error {
	h := sha256.New()
	h.Write([]byte(AdminListId))
	formIDBuf := h.Sum(nil)

	adminList := types.AdminList{
		AdminList: []int{initialAdmin},
	}

	formBuf, err := adminList.Serialize(ctx)
	if err != nil {
		return xerrors.Errorf("failed to marshal AdminList : %v", err)
	}

	err = snap.Set(formIDBuf, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	err = updateFormMetadataStore(snap, hex.EncodeToString(formIDBuf))
	if err != nil {
		return xerrors.Errorf("failed to update the metadata in the store: %v", err)
	}

	return nil
}

// manageVotersForm implements commands.
// It performs the ADD or REMOVE VOTERS/OWNERS command
func (e evotingCommand) manageOwnersVotersForm(snap store.Snapshot, step execution.Step) error {
	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	var form types.Form
	var formID []byte

	txAddVoter, okAddVoter := msg.(types.AddVoter)
	txRemoveVoter, okRemoveVoter := msg.(types.RemoveVoter)
	txAddOwner, okAddOwner := msg.(types.AddOwner)
	txRemoveOwner, okRemoveOwner := msg.(types.RemoveOwner)

	if okAddVoter {
		form, formID, err = e.getForm(txAddVoter.FormID, snap)
		if err != nil {
			return xerrors.Errorf(errGetForm, err)
		}

		isOwner, err := e.isRole(form, txAddVoter.PerformingUserID, Owners)
		if err != nil {
			return xerrors.Errorf(errIsRole, err)
		}

		if !isOwner {
			return xerrors.Errorf(errNoOwnerPerms, txAddVoter.PerformingUserID)
		}

		err = form.AddVoter(txAddVoter.TargetUserID)
		if err != nil {
			return xerrors.Errorf("couldn't add voter: %v", err)
		}
	} else if okRemoveVoter {
		form, formID, err = e.getForm(txRemoveVoter.FormID, snap)
		if err != nil {
			return xerrors.Errorf(errGetForm, err)
		}

		isOwner, err := e.isRole(form, txRemoveVoter.PerformingUserID, Owners)
		if err != nil {
			return xerrors.Errorf(errIsRole, err)
		}

		if !isOwner {
			return xerrors.Errorf(errNoOwnerPerms, txRemoveVoter.PerformingUserID)
		}

		err = form.RemoveVoter(txRemoveVoter.TargetUserID)
		if err != nil {
			return xerrors.Errorf("couldn't remove voter: %v", err)
		}
	} else if okAddOwner {
		form, formID, err = e.getForm(txAddOwner.FormID, snap)
		if err != nil {
			return xerrors.Errorf(errGetForm, err)
		}

		isOwner, err := e.isRole(form, txAddOwner.PerformingUserID, Owners)
		if err != nil {
			return xerrors.Errorf(errIsRole, err)
		}

		if !isOwner {
			return xerrors.Errorf(errNoOwnerPerms, txAddOwner.PerformingUserID)
		}

		err = form.AddOwner(txAddOwner.TargetUserID)
		if err != nil {
			return xerrors.Errorf("couldn't add owner: %v", err)
		}
	} else if okRemoveOwner {
		form, formID, err = e.getForm(txRemoveOwner.FormID, snap)
		if err != nil {
			return xerrors.Errorf(errGetForm, err)
		}

		isOwner, err := e.isRole(form, txRemoveOwner.PerformingUserID, Owners)
		if err != nil {
			return xerrors.Errorf(errIsRole, err)
		}

		if !isOwner {
			return xerrors.Errorf(errNoOwnerPerms, txRemoveOwner.PerformingUserID)
		}

		err = form.RemoveOwner(txRemoveOwner.TargetUserID)
		if err != nil {
			return xerrors.Errorf("couldn't remove owner: %v", err)
		}
	} else {
		return xerrors.Errorf(errWrongTx, msg)
	}

	formBuf, err := form.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Form : %v", err)
	}

	err = snap.Set(formID, formBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// isMemberOf is a utility function to verify if a public key is associated to a
// member of the roster or not. Returns nil if it's the case.
func isMemberOf(roster authority.Authority, publicKey []byte) error {
	pubKeyIterator := roster.PublicKeyIterator()
	isAMember := false

	for pubKeyIterator.HasNext() {
		key, err := pubKeyIterator.GetNext().MarshalBinary()
		if err != nil {
			return xerrors.Errorf("failed to serialize a public key from the roster : %v ", err)
		}

		if bytes.Equal(publicKey, key) {
			isAMember = true
		}
	}

	if !isAMember {
		return xerrors.Errorf("public key not associated to a member of the roster: %x", publicKey)
	}

	return nil
}

// SemiRandomStream implements cipher.Stream
type SemiRandomStream struct {
	// Seed is the seed on which should be based our random number generation
	seed []byte

	stream *rand.Rand
}

// NewSemiRandomStream returns a new initialized semi-random struct based on
// math.Rand. This random stream is not cryptographically safe.
//
// - implements cipher.Stream
func NewSemiRandomStream(seed []byte) (SemiRandomStream, error) {
	if len(seed) > 8 {
		seed = seed[0:8]
	}

	s, n := binary.Varint(seed)
	if n <= 0 {
		return SemiRandomStream{}, xerrors.Errorf("the seed has a wrong size (too small)")
	}

	source := rand.NewSource(s)
	stream := rand.New(source)

	return SemiRandomStream{stream: stream, seed: seed}, nil
}

// XORKeyStream implements cipher.Stream
func (s SemiRandomStream) XORKeyStream(dst, src []byte) {
	key := make([]byte, len(src))

	_, err := s.stream.Read(key)
	if err != nil {
		panic("error reading into semi random stream :" + err.Error())
	}

	xof := suite.XOF(key)
	xof.XORKeyStream(dst, src)
}

// getForm gets the form from the snap. Returns the form ID NOT hex
// encoded.
func (e evotingCommand) getForm(formIDHex string,
	snap store.Snapshot) (types.Form, []byte, error) {

	var form types.Form

	formIDBuf, err := hex.DecodeString(formIDHex)
	if err != nil {
		return form, nil, xerrors.Errorf("failed to decode formIDHex: %v", err)
	}

	form, err = types.FormFromStore(e.context, e.formFac, formIDHex, snap)
	if err != nil {
		return form, nil, xerrors.Errorf("failed to get key %q: %v", formIDBuf, err)
	}

	return form, formIDBuf, nil
}

// getAdminList gets the AdminList from the snap. Returns the form ID NOT hex
// encoded.
func (e evotingCommand) getAdminList(snap store.Snapshot) (types.AdminList, error) {

	var form types.AdminList

	form, err := types.AdminListFromStore(e.context, e.adminListFac, snap, AdminListId)
	if err != nil {
		return form, xerrors.Errorf("failed to get the AdminList: %v", err)
	}

	return form, nil
}

// getTransaction extracts the argument from the transaction.
func (e evotingCommand) getTransaction(tx txn.Transaction) (serde.Message, error) {
	buff := tx.GetArg(FormArg)
	if len(buff) == 0 {
		return nil, xerrors.Errorf("%q not found in tx arg", FormArg)
	}

	message, err := e.transactionFac.Deserialize(e.context, buff)
	if err != nil {
		return nil, xerrors.Errorf("failed to deserialize transaction: %v", err)
	}

	return message, nil
}

// decrypt combines the public shares to reconstruct the secret
// (i.e. encrypted ballots).
func decrypt(ballot int, pair int, allPubShares []types.PubsharesUnit, indexes []int) (
	[]byte, error) {
	pubShares := make([]*share.PubShare, 0)

	for i := 0; i < len(allPubShares); i++ {
		if allPubShares[i] != nil { // can be nil since not all nodes need to submit
			pubShare := allPubShares[i][ballot][pair]

			pubShares = append(pubShares, &share.PubShare{
				I: indexes[i],
				V: pubShare,
			})
		}
	}

	res, err := share.RecoverCommit(suite, pubShares, len(pubShares), len(pubShares))
	if err != nil {
		return nil, xerrors.Errorf("failed to recover commit: %v", err)
	}

	decryptedMessage, err := res.Data()
	if err != nil {
		return nil, xerrors.Errorf("failed to get embedded data: %v", err)
	}

	return decryptedMessage, nil
}
