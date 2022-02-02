package evoting

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math/rand"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/cosi/threshold"
	_ "go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/bls"
	_ "go.dedis.ch/dela/crypto/bls/json"
	"go.dedis.ch/kyber/v3/proof"
	"go.dedis.ch/kyber/v3/shuffle"
	"golang.org/x/xerrors"
)

const (
	shufflingProtocolName = "PairShuffle"
	errGetTransaction     = "failed to get transaction: %v"
	errGetElection        = "failed to get election: %v"
)

// evotingCommand implements the commands of the Evoting contract.
//
// - implements commands
type evotingCommand struct {
	*Contract

	prover prover
}

type prover func(suite proof.Suite, protocolName string, verifier proof.Verifier, proof []byte) error

// createElection implements commands. It performs the CREATE_ELECTION command
func (e evotingCommand) createElection(snap store.Snapshot, step execution.Step) error {

	var tx types.CreateElectionTransaction

	err := getTransaction(step.Current, &tx)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
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

	election := types.Election{
		ElectionID:    hex.EncodeToString(electionIDBuf),
		Configuration: tx.Configuration,
		AdminID:       tx.AdminID,
		Status:        types.Initial,
		// Pubkey is set by the opening command
		BallotSize:          tx.Configuration.MaxBallotSize(),
		PublicBulletinBoard: types.PublicBulletinBoard{},
		ShuffleInstances:    []types.ShuffleInstance{},
		DecryptedBallots:    []types.Ballot{},
		// We set the participant in the e-voting once for all. If it happens
		// that 1/3 of the participants go away, the election will never end.
		RosterBuf:        append([]byte{}, rosterBuf...),
		ShuffleThreshold: threshold.ByzantineThreshold(roster.Len()),
	}

	electionJSON, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionIDBuf, electionJSON)
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

	var tx types.OpenElectionTransaction

	err := getTransaction(step.Current, &tx)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	election, electionID, err := getElection(tx.ElectionID, snap)
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

	pubkeyBuf, err := pubkey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to marshal pubkey: %v", err)
	}

	election.Pubkey = pubkeyBuf

	electionMarshaled, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionID, electionMarshaled)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// castVote implements commands. It performs the CAST_VOTE command
func (e evotingCommand) castVote(snap store.Snapshot, step execution.Step) error {

	var tx types.CastVoteTransaction

	err := getTransaction(step.Current, &tx)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	election, electionID, err := getElection(tx.ElectionID, snap)
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

	for _, ciphertext := range tx.Ballot {
		if len(ciphertext.K) == 0 || len(ciphertext.C) == 0 {
			return xerrors.Errorf("part of the casted ballot has empty El Gamal pairs")
		}
		_, _, err = ciphertext.GetPoints()
		if err != nil {
			return xerrors.Errorf("casted ballot has invalid El Gamal pairs: %v", err)
		}
	}

	election.PublicBulletinBoard.CastVote(tx.UserID, tx.Ballot)

	electionMarshaled, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionID, electionMarshaled)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// shuffleBallots implements commands. It performs the SHUFFLE_BALLOTS command
func (e evotingCommand) shuffleBallots(snap store.Snapshot, step execution.Step) error {

	var tx types.ShuffleBallotsTransaction

	err := getTransaction(step.Current, &tx)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	err = checkPreviousTransactions(step, tx.Round)
	if err != nil {
		return xerrors.Errorf("check previous transactions failed: %v", err)
	}

	election, electionID, err := getElection(tx.ElectionID, snap)
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

	// Check the shuffler is a valid member of the roster
	roster, err := e.rosterFac.AuthorityOf(e.context, election.RosterBuf)
	if err != nil {
		return xerrors.Errorf("failed to deserialize roster: %v", err)
	}

	err = isMemberOf(roster, shufflerPublicKey)
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

	shuffleHash, err := tx.HashShuffle(electionID)
	if err != nil {
		return xerrors.Errorf("could not hash shuffle : %v", err)
	}

	// Check the signature matches the shuffle using the shuffler's public key
	err = signerPubKey.Verify(shuffleHash, signature)
	if err != nil {
		return xerrors.Errorf("signature does not match the Shuffle : %v ", err)
	}

	// Retrieve the random vector (ie the Scalar vector)
	randomVector, err := tx.RandomVector.Unmarshal()
	if err != nil {
		return xerrors.Errorf("failed to unmarshal random vector: %v", err)
	}

	// Check that the random vector is correct
	semiRandomStream, err := NewSemiRandomStream(shuffleHash)
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

	XX, YY, err := tx.ShuffledBallots.GetElGPairs()
	if err != nil {
		return xerrors.Errorf("failed to get X, Y: %v", err)
	}

	// get the election public key
	pubKey := suite.Point()

	err = pubKey.UnmarshalBinary(election.Pubkey)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal public key: %v", err)
	}

	var encryptedBallots types.EncryptedBallots

	if tx.Round == 0 {
		encryptedBallots = election.PublicBulletinBoard.Ballots
	} else {
		// get the election's last shuffled ballots
		encryptedBallots = election.ShuffleInstances[len(election.ShuffleInstances)-1].ShuffledBallots
	}

	X, Y, err := encryptedBallots.GetElGPairs()
	if err != nil {
		return xerrors.Errorf("failed to get X, Y: %v", err)
	}

	XXUp, YYUp, XXDown, YYDown := shuffle.GetSequenceVerifiable(suite, X, Y, XX,
		YY, randomVector)

	verifier := shuffle.Verifier(suite, nil, pubKey, XXUp, YYUp, XXDown, YYDown)

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

	electionBuf, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshall Election : %v", err)
	}

	err = snap.Set(electionID, electionBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// checkPreviousTransactions checks if a ShuffleBallotsTransaction has already
// been accepted and executed for a specific round.
func checkPreviousTransactions(step execution.Step, round int) error {
	for _, tx := range step.Previous {

		if string(tx.GetArg(native.ContractArg)) == ContractName {

			if string(tx.GetArg(CmdArg)) == ElectionArg {

				shuffledBallotsBuf := tx.GetArg(ElectionArg)
				var shuffleBallotsTransaction types.ShuffleBallotsTransaction

				err := json.Unmarshal(shuffledBallotsBuf, &shuffleBallotsTransaction)
				if err != nil {
					return xerrors.Errorf("failed to unmarshall ShuffleBallotsTransaction : %v", err)
				}

				if shuffleBallotsTransaction.Round == round {
					return xerrors.Errorf("shuffle is already happening in this round")
				}
			}
		}
	}
	return nil
}

// closeElection implements commands. It performs the CLOSE_ELECTION command
func (e evotingCommand) closeElection(snap store.Snapshot, step execution.Step) error {

	var tx types.CloseElectionTransaction

	err := getTransaction(step.Current, &tx)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	election, electionID, err := getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.AdminID != tx.UserID {
		return xerrors.Errorf("only the admin can close the election")
	}

	if election.Status != types.Open {
		return xerrors.Errorf("the election is not open, current status: %d", election.Status)
	}

	if len(election.PublicBulletinBoard.Ballots) <= 1 {
		return xerrors.Errorf("at least two ballots are required")
	}

	election.Status = types.Closed

	electionMarshaled, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election: %v", err)
	}

	err = snap.Set(electionID, electionMarshaled)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// decryptBallots implements commands. It performs the DECRYPT_BALLOTS command
func (e evotingCommand) decryptBallots(snap store.Snapshot, step execution.Step) error {

	var tx types.DecryptBallotsTransaction

	err := getTransaction(step.Current, &tx)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	election, electionID, err := getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.AdminID != tx.UserID {
		return xerrors.Errorf("only the admin can decrypt the ballots")
	}

	if election.Status != types.ShuffledBallots {
		return xerrors.Errorf("the ballots are not shuffled, current status: %d", election.Status)
	}

	election.Status = types.ResultAvailable
	election.DecryptedBallots = tx.DecryptedBallots

	electionMarshaled, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshall Election : %v", err)
	}

	err = snap.Set(electionID, electionMarshaled)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// cancelElection implements commands. It performs the CANCEL_ELECTION command
func (e evotingCommand) cancelElection(snap store.Snapshot, step execution.Step) error {

	var tx types.CancelElectionTransaction

	err := getTransaction(step.Current, &tx)
	if err != nil {
		return xerrors.Errorf(errGetTransaction, err)
	}

	election, electionID, err := getElection(tx.ElectionID, snap)
	if err != nil {
		return xerrors.Errorf(errGetElection, err)
	}

	if election.AdminID != tx.UserID {
		return xerrors.Errorf("only the admin can cancel the election")
	}

	election.Status = types.Canceled

	js, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionID, js)
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
func getElection(electionIDHex string, snap store.Snapshot) (types.Election, []byte, error) {
	var election types.Election

	electionID, err := hex.DecodeString(electionIDHex)
	if err != nil {
		return election, nil, xerrors.Errorf("failed to decode electionIDHex: %v", err)
	}

	electionBuff, err := snap.Get(electionID)
	if err != nil {
		return election, nil, xerrors.Errorf("failed to get key %q: %v", electionID, err)
	}

	err = json.Unmarshal(electionBuff, &election)
	if err != nil {
		return election, nil, xerrors.Errorf("failed to unmarshal Election: %v", err)
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

// getTransaction extracts the argument from the transaction and unmarshals it
// to e. e MUST be a pointer.
func getTransaction(tx txn.Transaction, e interface{}) error {
	buff := tx.GetArg(ElectionArg)
	if len(buff) == 0 {
		return xerrors.Errorf("%q not found in tx arg", ElectionArg)
	}

	err := json.Unmarshal(buff, e)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal e: %v", err)
	}

	return nil
}
