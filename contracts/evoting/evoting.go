package evoting

import (
	"encoding/hex"
	"encoding/json"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	"go.dedis.ch/kyber/v3/shuffle"
	"golang.org/x/xerrors"
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
func (e evotingCommand) createElection(snap store.Snapshot, step execution.Step, dkgActor dkg.Actor) error {
	createElectionBuf := step.Current.GetArg(CreateElectionArg)
	if len(createElectionBuf) == 0 {
		return xerrors.Errorf(errArgNotFound, CreateElectionArg)
	}

	createElectionTxn := &types.CreateElectionTransaction{}
	err := json.Unmarshal(createElectionBuf, createElectionTxn)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal CreateElectionTransaction : %v", err)
	}

	publicKey, err := dkgActor.GetPublicKey()
	if err != nil {
		return xerrors.Errorf("failed to get dkg public key : %v", err)
	}

	publicKeyBuf, err := publicKey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to marshall dkg public key : %v", err)
	}

	if 3*createElectionTxn.ShuffleThreshold < 2*len(createElectionTxn.Members)+1 {
		return xerrors.Errorf("the shuffle threshold is too low: we require 3T >= 2N + 1")
	}

	election := types.Election{
		Title:            createElectionTxn.Title,
		ElectionID:       types.ID(createElectionTxn.ElectionID),
		AdminId:          createElectionTxn.AdminId,
		Status:           types.Open,
		Pubkey:           publicKeyBuf,
		EncryptedBallots: map[string][]byte{},
		ShuffledBallots:  map[int][][]byte{},
		ShuffleProofs:    map[int][]byte{},
		DecryptedBallots: []types.Ballot{},
		ShuffleThreshold: createElectionTxn.ShuffleThreshold,
		Members:          createElectionTxn.Members,
		Format:           createElectionTxn.Format,
	}

	electionJson, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	electionIDBuff, err := hex.DecodeString(string(election.ElectionID))
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	err = snap.Set(electionIDBuff, electionJson)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	electionsMetadataBuff, err := snap.Get([]byte(ElectionsMetadataKey))
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionsMetadataBuff, err)
	}

	electionsMetadata := &types.ElectionsMetadata{}

	if len(electionsMetadataBuff) == 0 {
		electionsMetadata.ElectionsIds = []string{}
	} else {
		err := json.Unmarshal(electionsMetadataBuff, electionsMetadata)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal ElectionsMetadata: %v", err)
		}
	}

	electionsMetadata.ElectionsIds = append(electionsMetadata.ElectionsIds, createElectionTxn.ElectionID)

	electionMetadataJson, err := json.Marshal(electionsMetadata)
	if err != nil {
		return xerrors.Errorf("failed to marshal ElectionsMetadata: %v", err)
	}

	err = snap.Set([]byte(ElectionsMetadataKey), electionMetadataJson)
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

	election.EncryptedBallots[castVoteTransaction.UserId] = castVoteTransaction.Ballot

	electionMarshaled, err = json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
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
		// todo : send status ?
		return xerrors.Errorf("the election is not closed")
	}

	if len(election.ShuffledBallots) != shuffleBallotsTransaction.Round-1 {
		return xerrors.Errorf("wrong number of shuffled ballots: expected '%d', got '%d'",
			shuffleBallotsTransaction.Round-1, len(election.ShuffledBallots))
	}

	KsShuffled := make([]kyber.Point, 0, len(shuffleBallotsTransaction.ShuffledBallots))
	CsShuffled := make([]kyber.Point, 0, len(shuffleBallotsTransaction.ShuffledBallots))
	for _, shuffledBallot := range shuffleBallotsTransaction.ShuffledBallots {
		ciphertext := &types.Ciphertext{}
		err = json.Unmarshal(shuffledBallot, ciphertext)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal Ciphertext: %v", err)
		}

		K := suite.Point()
		err = K.UnmarshalBinary(ciphertext.K)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal K kyber.Point: %v", err)
		}

		C := suite.Point()
		err = C.UnmarshalBinary(ciphertext.C)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal C kyber.Point: %v", err)
		}

		KsShuffled = append(KsShuffled, K)
		CsShuffled = append(CsShuffled, C)
	}

	pubKey := suite.Point()
	err = pubKey.UnmarshalBinary(election.Pubkey)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal public key: %v", err)
	}

	Ks := make([]kyber.Point, 0, len(KsShuffled))
	Cs := make([]kyber.Point, 0, len(CsShuffled))

	encryptedBallotsMap := election.EncryptedBallots

	encryptedBallots := make([][]byte, 0, len(encryptedBallotsMap))

	if shuffleBallotsTransaction.Round == 1 {
		for _, encryptedBallot := range encryptedBallotsMap {
			encryptedBallots = append(encryptedBallots, encryptedBallot)
		}
	}

	if shuffleBallotsTransaction.Round > 1 {
		encryptedBallots = election.ShuffledBallots[shuffleBallotsTransaction.Round-1]
	}

	for _, encryptedBallot := range encryptedBallots {
		ciphertext := &types.Ciphertext{}
		err = json.Unmarshal(encryptedBallot, ciphertext)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal Ciphertext: %v", err)
		}

		K := suite.Point()
		err = K.UnmarshalBinary(ciphertext.K)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal K kyber.Point: %v", err)
		}

		C := suite.Point()
		err = C.UnmarshalBinary(ciphertext.C)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal C kyber.Point: %v", err)
		}

		Ks = append(Ks, K)
		Cs = append(Cs, C)
	}

	// todo: add trusted nodes in election struct
	verifier := shuffle.Verifier(suite, nil, pubKey, Ks, Cs, KsShuffled, CsShuffled)

	err = e.prover(suite, protocolName, verifier, shuffleBallotsTransaction.Proof)
	if err != nil {
		return xerrors.Errorf("proof verification failed: %v", err)
	}

	if shuffleBallotsTransaction.Round == election.ShuffleThreshold {
		election.Status = types.ShuffledBallots
	}

	election.ShuffledBallots[shuffleBallotsTransaction.Round] = shuffleBallotsTransaction.ShuffledBallots
	election.ShuffleProofs[shuffleBallotsTransaction.Round] = shuffleBallotsTransaction.Proof

	electionBuf, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshall Election : %v", err)
	}

	electionIDBuff, err := hex.DecodeString(string(election.ElectionID))
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

	if election.AdminId != closeElectionTransaction.UserId {
		return xerrors.Errorf("only the admin can close the election")
	}

	if election.Status != types.Open {
		return xerrors.Errorf("the election is not open, current status: %d", election.Status)
	}

	if len(election.EncryptedBallots) <= 1 {
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

	if election.AdminId != decryptBallotsTransaction.UserId {
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

	if election.AdminId != cancelElectionTransaction.UserId {
		return xerrors.Errorf("only the admin can cancel the election")
	}

	election.Status = types.Canceled

	js, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	electionIDBuff, err := hex.DecodeString(string(election.ElectionID))
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	err = snap.Set(electionIDBuff, js)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}
