package evoting

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	shuffleKyber "go.dedis.ch/kyber/v3/shuffle"

	// shuffleKyber "go.dedis.ch/kyber/v3/shuffle"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

const protocolName = "PairShuffle"
const messageOnlyOneShufflePerRound = "shuffle is already happening in this round"
const ElectionsMetadataKey = "ElectionsMetadataKey"

var suite = suites.MustFind("Ed25519")

// TODO : the smart contract should create its own dkg Actor

// commands defines the commands of the evoting contract.
type commands interface {
	createElection(snap store.Snapshot, step execution.Step, dkgActor dkg.Actor) error
	castVote(snap store.Snapshot, step execution.Step) error
	closeElection(snap store.Snapshot, step execution.Step) error
	shuffleBallots(snap store.Snapshot, step execution.Step) error
	decryptBallots(snap store.Snapshot, step execution.Step) error
	cancelElection(snap store.Snapshot, step execution.Step) error
}

const (
	// ContractName is the name of the contract.
	ContractName = "go.dedis.ch/dela.Evoting"

	// CmdArg is the argument's name to indicate the kind of command we want to
	// run on the contract. Should be one of the Command type.
	CmdArg = "evoting:command"

	CreateElectionArg = "evoting:createElectionArgs"

	CastVoteArg = "evoting:castVoteArgs"

	CancelElectionArg = "evoting:cancelElectionArgs"

	CloseElectionArg = "evoting:closeElectionArgs"

	ShuffleBallotsArg = "evoting:shuffleBallotsArgs"

	DecryptBallotsArg = "evoting:decryptBallotsArgs"

	// credentialAllCommand defines the credential command that is allowed to
	// perform all commands.
	credentialAllCommand = "all"
)

// Command defines a type of command for the value contract
type Command string

const (
	CmdCreateElection Command = "CREATE_ELECTION"

	CmdCastVote Command = "CAST_VOTE"

	CmdCloseElection Command = "CLOSE_ELECTION"

	CmdShuffleBallots Command = "SHUFFLE_BALLOTS"

	CmdDecryptBallots Command = "DECRYPT_BALLOTS"

	CmdCancelElection Command = "CANCEL_ELECTION"
)

// NewCreds creates new credentials for a evoting contract execution. We might
// want to use in the future a separate credential for each command.
func NewCreds(id []byte) access.Credential {
	return access.NewContractCreds(id, ContractName, credentialAllCommand)
}

// RegisterContract registers the value contract to the given execution service.
func RegisterContract(exec *native.Service, c Contract) {
	exec.Set(ContractName, c)
}

// Contract is a smart contract that allows one to execute evoting commands
//
// - implements native.Contract
type Contract struct {

	// access is the access control service managing this smart contract
	access access.Service

	// accessKey is the access identifier allowed to use this smart contract
	accessKey []byte

	// cmd provides the commands that can be executed by this smart contract
	cmd commands

	pedersen dkg.DKG
}

// NewContract creates a new Value contract
func NewContract(aKey []byte, srvc access.Service, pedersen dkg.DKG) Contract {
	contract := Contract{
		// indexElection:     map[string]struct{}{},
		access:    srvc,
		accessKey: aKey,
		pedersen:  pedersen,
	}

	contract.cmd = evotingCommand{Contract: &contract}
	return contract
}

func (c Contract) Execute(snap store.Snapshot, step execution.Step) error {
	creds := NewCreds(c.accessKey)

	err := c.access.Match(snap, creds, step.Current.GetIdentity())
	if err != nil {
		return xerrors.Errorf("identity not authorized: %v (%v)",
			step.Current.GetIdentity(), err)
	}

	cmd := step.Current.GetArg(CmdArg)
	if len(cmd) == 0 {
		return xerrors.Errorf("'%s' not found in tx arg", CmdArg)
	}

	switch Command(cmd) {
	case CmdCreateElection:
		dkgActor, err := c.pedersen.GetLastActor()
		if err != nil {
			return xerrors.Errorf("failed to get dkgActor: %v", err)
		}
		err = c.cmd.createElection(snap, step, dkgActor)
		if err != nil {
			return xerrors.Errorf("failed to create election: %v", err)
		}
	case CmdCastVote:
		err := c.cmd.castVote(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to cast vote: %v", err)
		}
	case CmdCloseElection:
		err := c.cmd.closeElection(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to close election: %v", err)
		}
	case CmdShuffleBallots:
		err := c.cmd.shuffleBallots(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to shuffle ballots: %v", err)
		}
	case CmdDecryptBallots:
		err := c.cmd.decryptBallots(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to decrypt ballots: %v", err)
		}
	case CmdCancelElection:
		err := c.cmd.cancelElection(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to cancel election: %v", err)
		}
	default:
		return xerrors.Errorf("unknown command: %s", cmd)
	}

	return nil
}

// evotingCommand implements the commands of the evoting contract
//
// - implements commands
type evotingCommand struct {
	*Contract
}

// createElection implements commands. It performs the CREATE_ELECTION command
func (e evotingCommand) createElection(snap store.Snapshot, step execution.Step, dkgActor dkg.Actor) error {
	createElectionArg := step.Current.GetArg(CreateElectionArg)
	if len(createElectionArg) == 0 {
		return xerrors.Errorf("'%s' not found in tx arg", CreateElectionArg)
	}

	createElectionTransaction := new(types.CreateElectionTransaction)
	err := json.NewDecoder(bytes.NewBuffer(createElectionArg)).Decode(createElectionTransaction)
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

	if float64(createElectionTransaction.ShuffleThreshold) < (float64(2) / 3 * float64(len(createElectionTransaction.Members))) {
		return xerrors.Errorf("the threshold is too low, it should be at least 2/3 of the length of the roster")
	}

	election := types.Election{
		Title:            createElectionTransaction.Title,
		ElectionID:       types.ID(createElectionTransaction.ElectionID),
		AdminId:          createElectionTransaction.AdminId,
		Status:           types.Open,
		Pubkey:           publicKeyBuf,
		EncryptedBallots: map[string][]byte{},
		ShuffledBallots:  map[int][][]byte{},
		Proofs:           map[int][]byte{},
		DecryptedBallots: []types.Ballot{},
		ShuffleThreshold: createElectionTransaction.ShuffleThreshold,
		Members:          createElectionTransaction.Members,
		Format:           createElectionTransaction.Format,
	}

	js, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	electionIDBuff, err := hex.DecodeString(string(election.ElectionID))
	if err != nil {
		return xerrors.Errorf("failed to decode ElectionID : %v", err)
	}

	err = snap.Set(electionIDBuff, js)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	electionsMetadataBuff, err := snap.Get([]byte(ElectionsMetadataKey))
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionsMetadataBuff, err)
	}

	electionsMetadata := new(types.ElectionsMetadata)

	if len(electionsMetadataBuff) == 0 {
		electionsMetadata.ElectionsIds = []string{}
	} else {
		err = json.NewDecoder(bytes.NewBuffer(electionsMetadataBuff)).Decode(electionsMetadata)
		if err != nil {
			return xerrors.Errorf("failed to unmarshal ElectionsMetadata: %v", err)
		}
	}

	electionsMetadata.ElectionsIds = append(electionsMetadata.ElectionsIds, createElectionTransaction.ElectionID)

	js, err = json.Marshal(electionsMetadata)
	if err != nil {
		return xerrors.Errorf("failed to marshal ElectionsMetadata: %v", err)
	}

	err = snap.Set([]byte(ElectionsMetadataKey), js)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// castVote implements commands. It performs the CAST_VOTE command
func (e evotingCommand) castVote(snap store.Snapshot, step execution.Step) error {
	castVoteArg := step.Current.GetArg(CastVoteArg)
	if len(castVoteArg) == 0 {
		return xerrors.Errorf("'%s' not found in tx arg", CastVoteArg)
	}

	castVoteTransaction := new(types.CastVoteTransaction)
	err := json.NewDecoder(bytes.NewBuffer(castVoteArg)).Decode(castVoteTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal CastVoteTransaction: %v", err)
	}

	electionTxIDBuff, _ := hex.DecodeString(castVoteTransaction.ElectionID)

	electionMarshalled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionTxIDBuff, err)
	}

	election := new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(electionMarshalled)).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election: %v", err)
	}

	if election.Status != types.Open {
		// todo : send status ?
		return xerrors.Errorf("the election is not open")
	}

	election.EncryptedBallots[castVoteTransaction.UserId] = castVoteTransaction.Ballot

	js, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election : %v", err)
	}

	electionIDBuff, _ := hex.DecodeString(string(election.ElectionID))

	err = snap.Set(electionIDBuff, js)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil

}

// shuffleBallots implements commands. It performs the SHUFFLE_BALLOTS command
func (e evotingCommand) shuffleBallots(snap store.Snapshot, step execution.Step) error {

	shuffleBallotsArg := step.Current.GetArg(ShuffleBallotsArg)
	if len(shuffleBallotsArg) == 0 {
		return xerrors.Errorf("'%s' not found in tx arg", ShuffleBallotsArg)
	}

	shuffleBallotsTransaction := new(types.ShuffleBallotsTransaction)
	err := json.NewDecoder(bytes.NewBuffer(shuffleBallotsArg)).Decode(shuffleBallotsTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal ShuffleBallotsTransaction: %v", err)
	}

	err = checkPreviousTransactions(step, shuffleBallotsTransaction.Round)
	if err != nil {
		return xerrors.Errorf("check previous transactions failed: %v", err)
	}

	electionTxIDBuff, _ := hex.DecodeString(shuffleBallotsTransaction.ElectionID)

	electionMarshalled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionTxIDBuff, err)
	}

	election := new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(electionMarshalled)).Decode(election)
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
	for _, v := range shuffleBallotsTransaction.ShuffledBallots {
		ciphertext := new(types.Ciphertext)
		err = json.NewDecoder(bytes.NewBuffer(v)).Decode(ciphertext)
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
		for _, value := range encryptedBallotsMap {
			encryptedBallots = append(encryptedBallots, value)
		}
	}

	if shuffleBallotsTransaction.Round > 1 {
		encryptedBallots = election.ShuffledBallots[shuffleBallotsTransaction.Round-1]
	}

	for _, v := range encryptedBallots {
		ciphertext := new(types.Ciphertext)
		err = json.NewDecoder(bytes.NewBuffer(v)).Decode(ciphertext)
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
	verifier := shuffleKyber.Verifier(suite, nil, pubKey, Ks, Cs, KsShuffled, CsShuffled)

	/*
		fmt.Printf(" KS : %v", Ks)
		fmt.Printf(" CS : %v", Cs)
		fmt.Printf(" KsShuffled : %v", KsShuffled)
		fmt.Printf(" CsShuffled : %v", CsShuffled)
	*/

	err = proof.HashVerify(suite, protocolName, verifier, shuffleBallotsTransaction.Proof)
	if err != nil {
		dela.Logger.Info().Msg("proof failed !!!!!!!!" + err.Error())
		// return xerrors.Errorf("proof verification failed: %v", err)
	}

	// todo : threshold should be part of election struct
	if shuffleBallotsTransaction.Round == election.ShuffleThreshold {
		election.Status = types.ShuffledBallots
	}

	election.ShuffledBallots[shuffleBallotsTransaction.Round] = shuffleBallotsTransaction.ShuffledBallots
	election.Proofs[shuffleBallotsTransaction.Round] = shuffleBallotsTransaction.Proof

	js, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshall Election : %v", err)
	}

	electionIDBuff, _ := hex.DecodeString(string(election.ElectionID))

	err = snap.Set(electionIDBuff, js)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// Checks if a ShuffleBallotsTransaction has already been accepted and executed
// for a specific round
func checkPreviousTransactions(step execution.Step, round int) error {
	for _, tx := range step.Previous {

		if string(tx.GetArg(native.ContractArg)) == ContractName {

			if string(tx.GetArg(CmdArg)) == ShuffleBallotsArg {

				shuffleBallotsArg := tx.GetArg(ShuffleBallotsArg)
				shuffleBallotsTransaction := new(types.ShuffleBallotsTransaction)

				err := json.NewDecoder(bytes.NewBuffer(shuffleBallotsArg)).Decode(shuffleBallotsTransaction)
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
	closeElectionArg := step.Current.GetArg(CloseElectionArg)
	if len(closeElectionArg) == 0 {
		return xerrors.Errorf("'%s' not found in tx arg", CloseElectionArg)
	}

	closeElectionTransaction := new(types.CloseElectionTransaction)
	err := json.NewDecoder(bytes.NewBuffer(closeElectionArg)).Decode(closeElectionTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal CloseElectionTransaction: %v", err)
	}

	electionTxIDBuff, _ := hex.DecodeString(closeElectionTransaction.ElectionID)

	electionMarshalled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionTxIDBuff, err)
	}

	election := new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(electionMarshalled)).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election: %v", err)
	}

	if election.AdminId != closeElectionTransaction.UserId {
		return xerrors.Errorf("only the admin can close the election")
	}

	if election.Status != types.Open {
		// todo : send status ?
		return xerrors.Errorf("the election is not open")
	}

	if len(election.EncryptedBallots) <= 1 {
		return xerrors.Errorf("at least two ballots are required")
	}

	election.Status = types.Closed

	js, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshal Election: %v", err)
	}

	electionIDBuff, _ := hex.DecodeString(string(election.ElectionID))

	err = snap.Set(electionIDBuff, js)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// decryptBallots implements commands. It performs the DECRYPT_BALLOTS command
func (e evotingCommand) decryptBallots(snap store.Snapshot, step execution.Step) error {
	decryptBallotsArg := step.Current.GetArg(DecryptBallotsArg)
	if len(decryptBallotsArg) == 0 {
		return xerrors.Errorf("'%s' not found in tx arg", DecryptBallotsArg)
	}

	decryptBallotsTransaction := new(types.DecryptBallotsTransaction)
	err := json.NewDecoder(bytes.NewBuffer(decryptBallotsArg)).Decode(decryptBallotsTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal DecryptBallotsTransaction: %v", err)
	}

	electionTxIDBuff, _ := hex.DecodeString(decryptBallotsTransaction.ElectionID)

	electionMarshalled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionTxIDBuff, err)
	}

	election := new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(electionMarshalled)).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election : %v", err)
	}

	if election.AdminId != decryptBallotsTransaction.UserId {
		return xerrors.Errorf("only the admin can decrypt the ballots")
	}

	if election.Status != types.ShuffledBallots {
		// todo : send status ?
		return xerrors.Errorf("the ballots are not shuffled")
	}

	election.Status = types.ResultAvailable
	election.DecryptedBallots = decryptBallotsTransaction.DecryptedBallots

	js, err := json.Marshal(election)
	if err != nil {
		return xerrors.Errorf("failed to marshall Election : %v", err)
	}

	electionIDBuff, _ := hex.DecodeString(string(election.ElectionID))

	err = snap.Set(electionIDBuff, js)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}

// cancelElection implements commands. It performs the CANCEL_ELECTION command
func (e evotingCommand) cancelElection(snap store.Snapshot, step execution.Step) error {
	cancelElectionArg := step.Current.GetArg(CancelElectionArg)
	if len(cancelElectionArg) == 0 {
		return xerrors.Errorf("'%s' not found in tx arg", CancelElectionArg)
	}

	cancelElectionTransaction := new(types.CancelElectionTransaction)
	err := json.NewDecoder(bytes.NewBuffer(cancelElectionArg)).Decode(cancelElectionTransaction)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal CancelElectionTransaction: %v", err)
	}

	electionTxIDBuff, _ := hex.DecodeString(cancelElectionTransaction.ElectionID)

	electionMarshalled, err := snap.Get(electionTxIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to get key '%s': %v", electionTxIDBuff, err)
	}

	election := new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(electionMarshalled)).Decode(election)
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

	electionIDBuff, _ := hex.DecodeString(string(election.ElectionID))

	err = snap.Set(electionIDBuff, js)
	if err != nil {
		return xerrors.Errorf("failed to set value: %v", err)
	}

	return nil
}
