// Code generated ...

package evoting

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela"
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
	errGetElection        = "failed to get election: %v"
	errWrongTx            = "wrong type of transaction: %T"
)

// parameters for the Unikernel connection
const (
	dialTimeout  = time.Second * 60
	writeTimeout = time.Second * 60
	readTimeout  = time.Second * 60
)

// evotingCommand implements the commands of the Evoting contract.
//
// - implements commands
type evotingCommand struct {
	*Contract

	prover prover
}

type prover func(suite proof.Suite, protocolName string, verifier proof.Verifier, proof []byte) error

// DialerFactory defines the function used by the smart contract to get a
// dialer.
type DialerFactory func(network, address string, timeout time.Duration) (net.Conn, error)

// createElection implements commands. It performs the CREATE_ELECTION command
func (e evotingCommand) createElection(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.CreateElection)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	rosterBuf, err := snap.Get(e.rosterKey)
	if err != nil {
		return xerrors.Errorf("failed to get roster")
	}

	roster, err := e.rosterFac.AuthorityOf(e.context, rosterBuf)
	if err != nil {
		return xerrors.Errorf("failed to get roster: %v", err)
	}

	// Get the electionID, which is the SHA256 of the transaction ID
	h := sha256.New()
	h.Write(step.Current.GetID())
	electionIDBuf := h.Sum(nil)

	if !tx.Configuration.IsValid() {
		return xerrors.Errorf("configuration of election is incoherent or has duplicated IDs")
	}

	units := types.PubsharesUnits{
		Pubshares: make([]types.PubsharesUnit, 0),
		PubKeys:   make([][]byte, 0),
		Indexes:   make([]int, 0),
	}

	election := types.Election{
		ElectionID:    hex.EncodeToString(electionIDBuf),
		Configuration: tx.Configuration,
		AdminID:       tx.AdminID,
		Status:        types.Initial,
		// Pubkey is set by the opening command
		BallotSize:       tx.Configuration.MaxBallotSize(),
		Suffragia:        types.Suffragia{},
		PubsharesUnits:   units,
		ShuffleInstances: []types.ShuffleInstance{},
		DecryptedBallots: []types.Ballot{},
		// We set the participant in the e-voting once for all. If it happens
		// that 1/3 of the participants go away, the election will never end.
		Roster:           roster,
		ShuffleThreshold: threshold.ByzantineThreshold(roster.Len()),
	}

	electionBuf, err := election.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionIDBuf, electionBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	// Update the election metadata store

	electionsMetadataBuf, err := snap.Get([]byte(ElectionsMetadataKey))
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionsMetadataBuf, err)
	}

	electionsMetadata := &types.ElectionsMetadata{
		ElectionsIDs: types.ElectionIDs{},
	}

	if len(electionsMetadataBuf) != 0 {
		err := json.Unmarshal(electionsMetadataBuf, electionsMetadata)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal ElectionsMetadata: %v", err)
		}
	}

	electionsMetadata.ElectionsIDs.Add(election.ElectionID)

	electionMetadataJSON, err := json.Marshal(electionsMetadata)
	if err != nil {
		return xerrors.Errorf("failed to marshal ElectionsMetadata: %v", err)
	}

	err = snap.Set([]byte(ElectionsMetadataKey), electionMetadataJSON)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// openElection set the public key on the election. The public key is fetched
// from the DKG actor. It works only if DKG is set up.
func (e evotingCommand) openElection(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.OpenElection)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	election, electionID, err := e.getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.Status != types.Initial {
		return xerrors.Errorf("the election was opened before, current status: %d", election.Status)
	}

	election.Status = types.Open

	if election.Pubkey != nil {
		return xerrors.Errorf("pubkey is already set: %s", election.Pubkey)
	}

	dkgActor, exists := e.pedersen.GetActor(electionID)
	if !exists {
		return xerrors.Errorf("failed to get actor for election %q", election.ElectionID)
	}

	pubkey, err := dkgActor.GetPublicKey()
	if err != nil {
		return xerrors.Errorf("failed to get pubkey: %v", err)
	}

	election.Pubkey = pubkey

	electionBuf, err := election.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionID, electionBuf)
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

	election, electionID, err := e.getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.Status != types.Open {
		return xerrors.Errorf("the election is not open, current status: %d", election.Status)
	}

	if len(tx.Ballot) != election.ChunksPerBallot() {
		return xerrors.Errorf("the ballot has unexpected length: %d != %d",
			len(tx.Ballot), election.ChunksPerBallot())
	}

	election.Suffragia.CastVote(tx.UserID, tx.Ballot)

	electionBuf, err := election.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionID, electionBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

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

	election, electionID, err := e.getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.Status != types.Closed {
		return xerrors.Errorf("the election is not closed")
	}

	// Round starts at 0
	expectedRound := len(election.ShuffleInstances)

	if tx.Round != expectedRound {
		return xerrors.Errorf("wrong shuffle round: expected round '%d', "+
			"transaction is for round '%d'", expectedRound, tx.Round)
	}

	shufflerPublicKey := tx.PublicKey

	err = isMemberOf(election.Roster, shufflerPublicKey)
	if err != nil {
		return xerrors.Errorf("could not verify identity of shuffler : %v", err)
	}

	// Check the node who submitted the shuffle did not already submit an
	// accepted shuffle
	for i, shuffleInstance := range election.ShuffleInstances {
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

	if election.ChunksPerBallot() != len(randomVector) {
		return xerrors.Errorf("randomVector has unexpected length : %v != %v",
			len(randomVector), election.ChunksPerBallot())
	}

	for i := 0; i < election.ChunksPerBallot(); i++ {
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
		ciphervotes = election.Suffragia.Ciphervotes
	} else {
		// get the election's last shuffled ballots
		lastIndex := len(election.ShuffleInstances) - 1
		ciphervotes = election.ShuffleInstances[lastIndex].ShuffledBallots
	}

	if len(ciphervotes) < 2 {
		return xerrors.Errorf("not enough votes: %d < 2", len(ciphervotes))
	}

	X, Y := types.CiphervotesToPairs(ciphervotes)

	XXUp, YYUp, XXDown, YYDown := shuffle.GetSequenceVerifiable(suite, X, Y, XX,
		YY, randomVector)

	verifier := shuffle.Verifier(suite, nil, election.Pubkey, XXUp, YYUp, XXDown, YYDown)

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

	election.ShuffleInstances = append(election.ShuffleInstances, currentShuffleInstance)

	// in case we have enough shuffled ballots, we update the status
	if len(election.ShuffleInstances) >= election.ShuffleThreshold {
		election.Status = types.ShuffledBallots
	}

	electionBuf, err := election.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionID, electionBuf)
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

		// skip tx that does not contain the election argument
		if string(tx.GetArg(CmdArg)) != ElectionArg {
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

// closeElection implements commands. It performs the CLOSE_ELECTION command
func (e evotingCommand) closeElection(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.CloseElection)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	election, electionID, err := e.getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.AdminID != tx.UserID {
		return xerrors.Errorf("only the admin can close the election")
	}

	if election.Status != types.Open {
		return xerrors.Errorf("the election is not open, current status: %d", election.Status)
	}

	if len(election.Suffragia.Ciphervotes) <= 1 {
		return xerrors.Errorf("at least two ballots are required")
	}

	election.Status = types.Closed

	electionBuf, err := election.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionID, electionBuf)
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

	election, electionID, err := e.getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.Status != types.ShuffledBallots {
		return xerrors.Errorf("the ballots have not been shuffled")
	}

	err = isMemberOf(election.Roster, tx.PublicKey)
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
	shuffledBallots := election.ShuffleInstances[len(election.ShuffleInstances)-1].
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

	// Check the node hasn't made any other submissions
	for _, key := range election.PubsharesUnits.PubKeys {
		if bytes.Equal(key, tx.PublicKey) {
			return xerrors.Errorf("'%x' already made a submission", key)
		}
	}

	for _, index := range election.PubsharesUnits.Indexes {
		if index == tx.Index {
			return xerrors.Errorf("a submission has already been made for index %d", index)
		}
	}

	// Add the pubshares to the election
	election.PubsharesUnits.Pubshares = append(election.PubsharesUnits.Pubshares, tx.Pubshares)
	election.PubsharesUnits.PubKeys = append(election.PubsharesUnits.PubKeys, tx.PublicKey)
	election.PubsharesUnits.Indexes = append(election.PubsharesUnits.Indexes, tx.Index)

	nbrSubmissions := len(election.PubsharesUnits.Pubshares)

	// For the Unikernel we must have as many shares as nodes. This is because
	// we are not storing the indexes of each node and the Unikernel is expected
	// all shares in order of indexes.
	if nbrSubmissions >= election.Roster.Len() {
		election.Status = types.PubSharesSubmitted
	}

	electionBuf, err := election.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election: %v", err)
	}

	err = snap.Set(electionID, electionBuf)
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

	election, electionID, err := e.getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.AdminID != tx.UserID {
		return xerrors.Errorf("only the admin can decrypt the ballots")
	}

	if election.Status != types.PubSharesSubmitted {
		return xerrors.Errorf("the public shares have not been submitted,"+
			" current status: %d", election.Status)
	}

	numNodes := len(election.PubsharesUnits.Pubshares)

	numBallots := len(election.PubsharesUnits.Pubshares[0])
	numChunks := len(election.PubsharesUnits.Pubshares[0][0])

	err = os.MkdirAll(e.exportFolder, os.ModePerm)
	if err != nil {
		return xerrors.Errorf("failed to create dir: %v", err)
	}

	// we must pass to the Unikernel the RELATIVE folder path from the one
	// mounted, i.e. e.exportFolder. This folder MUST be uniq for each
	// transaction.
	txIDHex := hex.EncodeToString(step.Current.GetID())
	tmp := fmt.Sprintf("decrypted-ballots-%s-%s-%d/", tx.ElectionID[:4], txIDHex[:4], time.Now().UnixNano())

	dela.Logger.Info().Msgf("using the following folder: %s", tmp)

	absPath := filepath.Join(e.exportFolder, tmp)

	err = exportPubshares(absPath, election.PubsharesUnits, numBallots, numChunks)
	if err != nil {
		return xerrors.Errorf("failed to export pubshares: %v", err)
	}

	conn, err := e.dialer("tcp", e.unikernelAddr, dialTimeout)
	if err != nil {
		return xerrors.Errorf("failed to dial TCP: %v", err)
	}

	if numChunks > 1<<32 {
		return xerrors.Errorf("numChunks can't be encoded to 32 bytes")
	}

	if numNodes > 1<<32 {
		return xerrors.Errorf("numNodes can't be encoded to 32 bytes")
	}

	buf := bytes.Buffer{}

	nc := make([]byte, 4)
	binary.LittleEndian.PutUint32(nc, uint32(numChunks))

	nn := make([]byte, 4)
	binary.LittleEndian.PutUint32(nn, uint32(numNodes))

	// Datagram sent to the Unikernel
	//
	//   ┌─────────────────────────────────────────── ─ ─ ─┐
	//   │                    256 bytes                    │
	//   ├─────────────┬─────────────┬─────────────── ─ ─ ─┤
	//   │   4 bytes   │   4 bytes   │   248 bytes         │
	//   ├─────────────┼─────────────┼─────────────── ─ ─ ─┤
	//   │  NumChunks  │  NumNodes   │  Folder path        │
	//   └─────────────┴─────────────┴─────────────── ─ ─ ─┘

	if len(tmp) > 248 {
		return xerrors.Errorf("folder path too big: %d", len(tmp))
	}

	buf.Write(nc)
	buf.Write(nn)

	buf.Write([]byte(tmp))

	conn.SetWriteDeadline(time.Now().Add(writeTimeout))

	dela.Logger.Info().Hex("output", buf.Bytes()).Msg("Data to Unikernel")

	_, err = buf.WriteTo(conn)
	if err != nil {
		return xerrors.Errorf("failed to write to unikernel: %v", err)
	}

	readRes := make([]byte, 256)

	conn.SetReadDeadline(time.Now().Add(readTimeout))

	n, err := conn.Read(readRes)
	if err != nil {
		return xerrors.Errorf("failed to read response from unikernel: %v", err)
	}

	readRes = readRes[:n]

	dela.Logger.Info().Hex("input", readRes).Msg("Data from Unikernel")

	if strings.HasPrefix(string(readRes), "ERROR") {
		return xerrors.Errorf("an error ocurred in the unikernel: %s", readRes)
	}

	rawBallots, err := importBallots(absPath, numChunks)
	if err != nil {
		return xerrors.Errorf("failed to import ballots: %v", err)
	}

	ballots := make([]types.Ballot, len(rawBallots))

	for i, rawBallot := range rawBallots {
		var ballot types.Ballot

		err = ballot.Unmarshal(string(rawBallot), election)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal ballot: %v", err)
		}

		ballots[i] = ballot
	}

	election.DecryptedBallots = ballots
	election.Status = types.ResultAvailable

	electionBuf, err := election.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionID, electionBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// cancelElection implements commands. It performs the CANCEL_ELECTION command
func (e evotingCommand) cancelElection(snap store.Snapshot, step execution.Step) error {

	msg, err := e.getTransaction(step.Current)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	tx, ok := msg.(types.CancelElection)
	if !ok {
		return xerrors.Errorf(errWrongTx, msg)
	}

	election, electionID, err := e.getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.AdminID != tx.UserID {
		return xerrors.Errorf("only the admin can cancel the election")
	}

	election.Status = types.Canceled

	electionBuf, err := election.Serialize(e.context)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionID, electionBuf)
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

// getElection gets the election from the snap. Returns the election ID NOT hex
// encoded.
func (e evotingCommand) getElection(electionIDHex string,
	snap store.Snapshot) (types.Election, []byte, error) {

	var election types.Election

	electionID, err := hex.DecodeString(electionIDHex)
	if err != nil {
		return election, nil, xerrors.Errorf("failed to decode electionIDHex: %v", err)
	}

	electionBuff, err := snap.Get(electionID)
	if err != nil {
		return election, nil, xerrors.Errorf("failed to get key %q: %v", electionID, err)
	}

	message, err := e.electionFac.Deserialize(e.context, electionBuff)
	if err != nil {
		return election, nil, xerrors.Errorf("failed to deserialize Election: %v", err)
	}

	election, ok := message.(types.Election)
	if !ok {
		return election, nil, xerrors.Errorf("wrong message type: %T", message)
	}

	if electionIDHex != election.ElectionID {
		return election, nil, xerrors.Errorf("electionID do not match: %q != %q",
			electionIDHex, election.ElectionID)
	}

	electionIDBuff, err := hex.DecodeString(electionIDHex)
	if err != nil {
		return election, nil, xerrors.Errorf("failed to get election id buff: %v", err)
	}

	return election, electionIDBuff, nil
}

// getTransaction extracts the argument from the transaction.
func (e evotingCommand) getTransaction(tx txn.Transaction) (serde.Message, error) {
	buff := tx.GetArg(ElectionArg)
	if len(buff) == 0 {
		return nil, xerrors.Errorf("%q not found in tx arg", ElectionArg)
	}

	message, err := e.transactionFac.Deserialize(e.context, buff)
	if err != nil {
		return nil, xerrors.Errorf("failed to deserialize transaction: %v", err)
	}

	return message, nil
}

// exportPubshares saves each ballot in a file. A file contains the pubshares,
// in order, of chunks. Each ballot will be saved as a "ballot_i" file.
func exportPubshares(folder string, units types.PubsharesUnits, numBallots, numChunks int) error {

	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return xerrors.Errorf("failed to make dir: %v", err)
	}

	// First we must sort the units by their node's indexes. This index is
	// provided by the DKG protocol.
	sortedUnits := make([]types.PubsharesUnit, len(units.Pubshares))

	for i, dkgIndex := range units.Indexes {
		sortedUnits[dkgIndex] = units.Pubshares[i]
	}

	fmt.Println("numBallots:", numBallots, "numChunks:", numChunks, "num pbshares:", len(units.Pubshares), "path:", folder)

	for ballotIndex := 0; ballotIndex < numBallots; ballotIndex++ {
		filepath := filepath.Join(folder, fmt.Sprintf("ballot_%d", ballotIndex))
		file, err := os.Create(filepath)
		if err != nil {
			return xerrors.Errorf("failed to create ballot file: %v", err)
		}

		for chunkIndex := 0; chunkIndex < numChunks; chunkIndex++ {
			for nodeIdex := range sortedUnits {
				p := sortedUnits[nodeIdex][ballotIndex][chunkIndex]

				_, err := p.MarshalTo(file)
				if err != nil {
					return xerrors.Errorf("failed to marshal to file: %v", err)
				}
			}
		}

		file.Close()
	}

	return nil
}

// importBallots parses all decrypted ballot files from a folder and returns a
// slice of decrypted results. Parses files that ends with ".decrypted". The
// decrypted file must contain chunks in order. A chunk is a kyber.Point that
// embeds the chunk's content.
func importBallots(folder string, numChunks int) ([][]byte, error) {
	result := [][]byte{}

	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, xerrors.Errorf("failed to read dir: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".decrypted") {
			continue
		}

		// a kyber.Point is 32 bytes long.
		expectedSize := 32 * numChunks

		if file.Size() != int64(expectedSize) {
			return nil, xerrors.Errorf("unexpected file size: %d != %d",
				file.Size(), expectedSize)
		}

		buf, err := os.ReadFile(filepath.Join(folder, file.Name()))
		if err != nil {
			return nil, xerrors.Errorf("failed to read file: %v", err)
		}

		point := suite.Point()
		decrypted := bytes.Buffer{}

		for i := 0; i < numChunks; i++ {
			err = point.UnmarshalBinary(buf[32*i : 32*i+32])
			if err != nil {
				return nil, xerrors.Errorf("failed to unmarshal point: %v", err)
			}

			data, err := point.Data()
			if err != nil {
				return nil, xerrors.Errorf("failed to read data: %v", err)
			}

			decrypted.Write(data)
		}

		result = append(result, decrypted.Bytes())
	}

	if err != nil {
		return nil, xerrors.Errorf("failed to parse files: %v", err)
	}

	return result, nil
}
