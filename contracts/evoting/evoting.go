package evoting

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/cosi/threshold"
	_ "go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/bls"
	_ "go.dedis.ch/dela/crypto/bls/json"
	"go.dedis.ch/kyber/v3/proof"
	"go.dedis.ch/kyber/v3/shuffle"
	"golang.org/x/xerrors"
	"math/rand"
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
	createElectionBuf := step.Current.GetArg(CreateElectionArg)
	if len(createElectionBuf) == 0 {
		return xerrors.Errorf(errArgNotFound, CreateElectionArg)
	}

	createElectionTxn := &types.CreateElectionTransaction{}
	err := json.Unmarshal(createElectionBuf, createElectionTxn)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal CreateElectionTransaction : %v", err)
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
	electionIDBuff := h.Sum(nil)

	if !createElectionTxn.Configuration.IsValid() {
		return xerrors.Errorf("configuration of election is incoherent or has duplicated IDs")
	}

	election := types.Election{
		Configuration: createElectionTxn.Configuration,
		ElectionID:    hex.EncodeToString(electionIDBuff),
		AdminID:       createElectionTxn.AdminID,
		Status:        types.Open,
		// Pubkey is set by the opening command
		BallotSize:          createElectionTxn.Configuration.MaxBallotSize(),
		PublicBulletinBoard: types.PublicBulletinBoard{},
		ShuffleInstances:    []types.ShuffleInstance{},
		DecryptedBallots:    []types.Ballot{},
		RosterBuf:           append([]byte{}, rosterBuf...),
		ShuffleThreshold:    threshold.ByzantineThreshold(roster.Len()),
	}

	electionJSON, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	err = snap.Set(electionIDBuff, electionJSON)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	// Update the election metadata store

	electionsMetadataBuff, err := snap.Get([]byte(ElectionsMetadataKey))
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionsMetadataBuff, err)
	}

	electionsMetadata := &types.ElectionsMetadata{}

	if len(electionsMetadataBuff) == 0 {
		electionsMetadata.ElectionsIDs = types.ElectionIDs{}
	} else {
		err := json.Unmarshal(electionsMetadataBuff, electionsMetadata)
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
// from the DKG actor. It works only if DKG is setup.
func (e evotingCommand) openElection(snap store.Snapshot, step execution.Step, dkgActor dkg.Actor) error {
	pubkey, err := dkgActor.GetPublicKey()
	if err != nil {
		return xerrors.Errorf("failed to get pubkey: %v", err)
	}

	openElecBuf := step.Current.GetArg(OpenElectionArg)
	if len(openElecBuf) == 0 {
		return xerrors.Errorf(errArgNotFound, OpenElectionArg)
	}

	openElectTransaction := &types.OpenElectionTransaction{}
	err = json.Unmarshal(openElecBuf, openElectTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal OpenElectionTransaction: %v", err)
	}

	electionTxIDBuff, err := hex.DecodeString(openElectTransaction.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	electionMarshaled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key %q: %v", electionTxIDBuff, err)
	}

	election := &types.Election{}
	err = json.Unmarshal(electionMarshaled, election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election: %v", err)
	}

	if election.Pubkey != nil {
		return xerrors.Errorf("pubkey is already set: %s", election.Pubkey)
	}

	if election.Status != types.Open {
		return xerrors.Errorf("the election is not open, current status: %d", election.Status)
	}

	pubkeyBuf, err := pubkey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to marshal pubkey: %v", err)
	}

	election.Pubkey = pubkeyBuf

	electionMarshaled, err = json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	electionIDBuff, err := hex.DecodeString(election.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	err = snap.Set(electionIDBuff, electionMarshaled)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// castVote implements commands. It performs the CAST_VOTE command
func (e evotingCommand) castVote(snap store.Snapshot, step execution.Step) error {
	castVoteBuf := step.Current.GetArg(CastVoteArg)
	if len(castVoteBuf) == 0 {
		return xerrors.Errorf(errArgNotFound, CastVoteArg)
	}

	castVoteTransaction := &types.CastVoteTransaction{}
	err := json.Unmarshal(castVoteBuf, castVoteTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal CastVoteTransaction: %v", err)
	}

	electionTxIDBuff, err := hex.DecodeString(castVoteTransaction.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	electionMarshaled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key %q: %v", electionTxIDBuff, err)
	}

	election := &types.Election{}
	err = json.Unmarshal(electionMarshaled, election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election: %v", err)
	}

	if election.Status != types.Open {
		return xerrors.Errorf("the election is not open, current status: %d", election.Status)
	}

	// TODO: check that castVoteTransaction.Ballot is a well formatted
	election.PublicBulletinBoard.CastVote(castVoteTransaction.UserID, castVoteTransaction.Ballot)

	electionMarshaled, err = json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	electionIDBuff, err := hex.DecodeString(election.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	err = snap.Set(electionIDBuff, electionMarshaled)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil

}

// shuffleBallots implements commands. It performs the SHUFFLE_BALLOTS command
func (e evotingCommand) shuffleBallots(snap store.Snapshot, step execution.Step) error {
	shuffledBallotsBuf := step.Current.GetArg(ShuffleBallotsArg)
	if len(shuffledBallotsBuf) == 0 {
		return xerrors.Errorf(errArgNotFound, ShuffleBallotsArg)
	}

	shuffleBallotsTransaction := &types.ShuffleBallotsTransaction{}
	err := json.Unmarshal(shuffledBallotsBuf, shuffleBallotsTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal ShuffleBallotsTransaction: %v", err)
	}

	err = checkPreviousTransactions(step, shuffleBallotsTransaction.Round)
	if err != nil {
		return xerrors.Errorf("check previous transactions failed: %v", err)
	}

	electionTxIDBuff, err := hex.DecodeString(shuffleBallotsTransaction.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	electionMarshaled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionTxIDBuff, err)
	}

	election := &types.Election{}
	err = json.Unmarshal(electionMarshaled, election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election : %v", err)
	}

	if election.Status != types.Closed {
		return xerrors.Errorf("the election is not closed")
	}

	// Round starts at 0
	expectedRound := len(election.ShuffleInstances)

	if shuffleBallotsTransaction.Round != expectedRound {
		return xerrors.Errorf("wrong shuffle round: expected round '%d', "+
			"transaction is for round '%d'", expectedRound, shuffleBallotsTransaction.Round)
	}

	shufflerPublicKey := shuffleBallotsTransaction.PublicKey

	// Check the shuffler is a valid member of the roster:
	roster, err := e.rosterFac.AuthorityOf(e.Contract.context, election.RosterBuf)
	if err != nil {
		return xerrors.Errorf("failed to deserialize roster: %v", err)
	}

	err = verifyIdentity(roster, shufflerPublicKey)
	if err != nil {
		return xerrors.Errorf("could not verify identity of shuffler : %v", err)
	}

	// Chek the node who submitted the shuffle did not already submit an accepted shuffle
	for i, shuffleInstance := range election.ShuffleInstances {
		if bytes.Equal(shufflerPublicKey, shuffleInstance.ShufflerPublicKey) {
			return xerrors.Errorf("a node already submitted a shuffle that"+
				" has been accepted in round %v", i)
		}
	}

	// Check the shuffler indeed signed the transaction:
	signerPubKey, err := bls.NewPublicKey(shuffleBallotsTransaction.PublicKey)
	if err != nil {
		return xerrors.Errorf("could not decode public key of signer : %v ", err)
	}

	txSignature := shuffleBallotsTransaction.Signature
	signature, err := bls.NewSignatureFactory().SignatureOf(e.context, txSignature)
	if err != nil {
		return xerrors.Errorf("could node deserialize shuffle signature : %v", err)
	}

	shuffleHash, err := shuffleBallotsTransaction.HashShuffle(election.ElectionID)
	if err != nil {
		return xerrors.Errorf("could not hash shuffle : %v", err)
	}

	// Check the signature matches the shuffle using the shuffler's public key:
	err = signerPubKey.Verify(shuffleHash, signature)
	if err != nil {
		return xerrors.Errorf("signature does not match the Shuffle : %v ", err)
	}

	// Retrieve the random vector (ie the Scalar vector)
	randomVector, err := shuffleBallotsTransaction.RandomVector.UnMarshal()
	if err != nil {
		return err
	}

	// Check the random vector is correct :
	semiRandomStream, err := NewSemiRandomStream(shuffleHash)
	if err != nil {
		return xerrors.Errorf("could not create semi-random stream")
	}

	lenRandomVector := len(election.PublicBulletinBoard.Ballots)
	for i := 0; i < lenRandomVector; i++ {
		v := suite.Scalar().Pick(semiRandomStream)
		if !randomVector[i].Equal(v) {
			return xerrors.Errorf("random vector from shuffle transaction is" +
				"different than expected random vector")
		}
	}

	XX, YY, err := shuffleBallotsTransaction.ShuffledBallots.GetElGPairs()
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

	if shuffleBallotsTransaction.Round == 0 {
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

	err = e.prover(suite, protocolName, verifier, shuffleBallotsTransaction.Proof)
	if err != nil {
		return xerrors.Errorf("proof verification failed: %v", err)
	}

	// append the new shuffled ballots and the proof to the lists
	currentShuffleInstance := types.ShuffleInstance{
		ShuffledBallots:   shuffleBallotsTransaction.ShuffledBallots,
		ShuffleProofs:     shuffleBallotsTransaction.Proof,
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

	electionIDBuff, err := hex.DecodeString(election.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	err = snap.Set(electionIDBuff, electionBuf)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// checkPreviousTransactions checks if a ShuffleBallotsTransaction has
// already been accepted and executed for a specific round.
func checkPreviousTransactions(step execution.Step, round int) error {
	for _, tx := range step.Previous {

		if string(tx.GetArg(native.ContractArg)) == ContractName {

			if string(tx.GetArg(CmdArg)) == ShuffleBallotsArg {

				shuffledBallotsBuf := tx.GetArg(ShuffleBallotsArg)
				shuffleBallotsTransaction := &types.ShuffleBallotsTransaction{}

				err := json.Unmarshal(shuffledBallotsBuf, shuffleBallotsTransaction)
				if err != nil {
					return xerrors.Errorf("failed to unmarshall ShuffleBallotsTransaction : %v", err)
				}

				if shuffleBallotsTransaction.Round == round {
					return xerrors.Errorf(messageOnlyOneShufflePerRound)
				}
			}
		}
	}
	return nil
}

// closeElection implements commands. It performs the CLOSE_ELECTION command
func (e evotingCommand) closeElection(snap store.Snapshot, step execution.Step) error {
	closeElectionBuf := step.Current.GetArg(CloseElectionArg)
	if len(closeElectionBuf) == 0 {
		return xerrors.Errorf(errArgNotFound, CloseElectionArg)
	}
	closeElectionTransaction := &types.CloseElectionTransaction{}
	err := json.Unmarshal(closeElectionBuf, closeElectionTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal CloseElectionTransaction: %v", err)
	}
	electionTxIDBuff, err := hex.DecodeString(closeElectionTransaction.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}
	electionMarshaled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionTxIDBuff, err)
	}
	election := &types.Election{}
	err = json.Unmarshal(electionMarshaled, election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election: %v", err)
	}
	if election.AdminID != closeElectionTransaction.UserID {
		return xerrors.Errorf("only the admin can close the election")
	}
	if election.Status != types.Open {
		return xerrors.Errorf("the election is not open, current status: %d", election.Status)
	}

	if len(election.PublicBulletinBoard.Ballots) <= 1 {
		return xerrors.Errorf("at least two ballots are required")
	}

	election.Status = types.Closed
	electionMarshaled, err = json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election: %v", err)
	}
	electionIDBuff, err := hex.DecodeString(string(election.ElectionID))
	if err != nil {
		return xerrors.Errorf("failed to marshal Election ID: %v", err)
	}
	err = snap.Set(electionIDBuff, electionMarshaled)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// decryptBallots implements commands. It performs the DECRYPT_BALLOTS command
func (e evotingCommand) decryptBallots(snap store.Snapshot, step execution.Step) error {
	decryptBallotsBuf := step.Current.GetArg(DecryptBallotsArg)
	if len(decryptBallotsBuf) == 0 {
		return xerrors.Errorf(errArgNotFound, DecryptBallotsArg)
	}

	decryptBallotsTransaction := &types.DecryptBallotsTransaction{}
	err := json.Unmarshal(decryptBallotsBuf, decryptBallotsTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal DecryptBallotsTransaction: %v", err)
	}

	electionTxIDBuff, err := hex.DecodeString(decryptBallotsTransaction.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	electionMarshaled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionTxIDBuff, err)
	}

	election := &types.Election{}
	err = json.Unmarshal(electionMarshaled, election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election : %v", err)
	}

	if election.AdminID != decryptBallotsTransaction.UserID {
		return xerrors.Errorf("only the admin can decrypt the ballots")
	}

	if election.Status != types.ShuffledBallots {
		return xerrors.Errorf("the ballots are not shuffled, current status: %d", election.Status)
	}

	election.Status = types.ResultAvailable
	election.DecryptedBallots = decryptBallotsTransaction.DecryptedBallots

	electionMarshaled, err = json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshall Election : %v", err)
	}

	electionIDBuff, err := hex.DecodeString(string(election.ElectionID))
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	err = snap.Set(electionIDBuff, electionMarshaled)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// cancelElection implements commands. It performs the CANCEL_ELECTION command
func (e evotingCommand) cancelElection(snap store.Snapshot, step execution.Step) error {
	cancelElectionBuf := step.Current.GetArg(CancelElectionArg)
	if len(cancelElectionBuf) == 0 {
		return xerrors.Errorf(errArgNotFound, CancelElectionArg)
	}

	cancelElectionTransaction := new(types.CancelElectionTransaction)
	err := json.Unmarshal(cancelElectionBuf, cancelElectionTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal CancelElectionTransaction: %v", err)
	}

	electionTxIDBuff, err := hex.DecodeString(cancelElectionTransaction.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	electionMarshaled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionTxIDBuff, err)
	}

	election := &types.Election{}
	err = json.Unmarshal(electionMarshaled, election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election : %v", err)
	}

	if election.AdminID != cancelElectionTransaction.UserID {
		return xerrors.Errorf("only the admin can cancel the election")
	}

	election.Status = types.Canceled

	js, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	electionIDBuff, err := hex.DecodeString(election.ElectionID)
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	err = snap.Set(electionIDBuff, js)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// verifyIdentity is a utility function to verify if a public key is associated
// to a member of the roster or not. Returns no error if it's the case.
func verifyIdentity(roster authority.Authority, publicKey []byte) error {
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

func (s SemiRandomStream) XORKeyStream(dst, src []byte) {
	key := make([]byte, len(src))

	_, err := s.stream.Read(key)
	if err != nil {
		panic("error reading into semi random stream :" + err.Error())
	}

	xof := suite.XOF(key)
	xof.XORKeyStream(dst, src)
}
