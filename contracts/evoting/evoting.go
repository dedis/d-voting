package evoting

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/cosi/threshold"
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

	minThreshold := threshold.ByzantineThreshold(len(createElectionTxn.Members))

	if createElectionTxn.ShuffleThreshold < minThreshold {
		return xerrors.Errorf("the shuffle threshold is too low: we require 3T >= 2N + 1, "+
			"found %d < %d", createElectionTxn.ShuffleThreshold, minThreshold)
	}

	election := types.Election{
		Title:            createElectionTxn.Title,
		ElectionID:       types.ID(createElectionTxn.ElectionID),
		AdminId:          createElectionTxn.AdminId,
		Status:           types.Open,
		Pubkey:           publicKeyBuf,
		EncryptedBallots: &types.EncryptedBallots{},
		ShuffledBallots:  [][][]byte{},
		ShuffledProofs:   [][]byte{},
		DecryptedBallots: []types.Ballot{},
		ShuffleThreshold: createElectionTxn.ShuffleThreshold,
		Members:          createElectionTxn.Members,
		Format:           createElectionTxn.Format,
	}

	electionJSON, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	electionIDBuff, err := hex.DecodeString(string(election.ElectionID))
	if err != nil {
		return xerrors.Errorf(errDecodeElectionID, err)
	}

	err = snap.Set(electionIDBuff, electionJSON)
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

	election.EncryptedBallots.CastVote(castVoteTransaction.UserId, castVoteTransaction.Ballot)

	// election.EncryptedBallots[castVoteTransaction.UserId] = castVoteTransaction.Ballot

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

	// Round starts at 1
	expectedRound := len(election.ShuffledBallots) + 1
	fmt.Println("expected round:", expectedRound)

	if shuffleBallotsTransaction.Round != expectedRound {
		return xerrors.Errorf("wrong shuffle round: expected round %d, "+
			"transaction is for round %s", expectedRound+1, shuffleBallotsTransaction.Round)
	}

	// Unmarshal the shuffled ballots from the transaction
	KsShuffled := make([]kyber.Point, len(shuffleBallotsTransaction.ShuffledBallots))
	CsShuffled := make([]kyber.Point, len(shuffleBallotsTransaction.ShuffledBallots))

	for i, shuffledBallot := range shuffleBallotsTransaction.ShuffledBallots {
		ciphertext := &types.Ciphertext{}
		err = json.Unmarshal(shuffledBallot, ciphertext)
		if err != nil {
			fmt.Println("ciphertext:", string(shuffledBallot))
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

		KsShuffled[i] = K
		CsShuffled[i] = C
	}

	// get the election public key
	pubKey := suite.Point()
	err = pubKey.UnmarshalBinary(election.Pubkey)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal public key: %v", err)
	}

	// unmarshal the current election shuffled ballots if any, or the encrypted
	// ballots.
	Ks := make([]kyber.Point, len(KsShuffled))
	Cs := make([]kyber.Point, len(CsShuffled))

	encryptedBallots := make([][]byte, len(election.EncryptedBallots.Ballots))

	if shuffleBallotsTransaction.Round == 1 {
		fmt.Println("this is the first round")
		for i, encryptedBallot := range election.EncryptedBallots.Ballots {
			encryptedBallots[i] = encryptedBallot
		}
	}

	if shuffleBallotsTransaction.Round > 1 {
		// get the election's last shuffled ballots
		fmt.Println("this is not the first round:", election.ShuffledBallots)
		encryptedBallots = election.ShuffledBallots[len(election.ShuffledBallots)-1]
	}

	for i, encryptedBallot := range encryptedBallots {
		ciphertext := &types.Ciphertext{}
		err = json.Unmarshal(encryptedBallot, ciphertext)
		if err != nil {
			fmt.Println("ciphertext2:", string(encryptedBallot))
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

		Ks[i] = K
		Cs[i] = C
	}

	// todo: add trusted nodes in election struct
	verifier := shuffle.Verifier(suite, nil, pubKey, Ks, Cs, KsShuffled, CsShuffled)

	err = e.prover(suite, protocolName, verifier, shuffleBallotsTransaction.Proof)
	if err != nil {
		return xerrors.Errorf("proof verification failed: %v", err)
	}

	// append the new shuffled ballots and the proof to the lists
	election.ShuffledBallots = append(election.ShuffledBallots, shuffleBallotsTransaction.ShuffledBallots)
	election.ShuffledProofs = append(election.ShuffledProofs, shuffleBallotsTransaction.Proof)

	// in case we have enough shuffled ballots, we update the status
	if len(election.ShuffledBallots) >= election.ShuffleThreshold {
		election.Status = types.ShuffledBallots
	}

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

	fmt.Println("election has new shuffled ballots, len=", len(election.ShuffledBallots))

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

	if len(election.EncryptedBallots.Ballots) <= 1 {
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
