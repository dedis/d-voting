package evoting

import (
	dvoting "github.com/dedis/d-voting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/prometheus/client_golang/prometheus"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/execution"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/dela/serde/json"

	"go.dedis.ch/kyber/v3/proof"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"

	// Register the JSON format for the election
	_ "github.com/dedis/d-voting/contracts/evoting/json"
)

var (
	PromElectionStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "evoting_status",
		Help: "status of election",
	},
		[]string{"election"},
	)

	PromElectionBallots = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "evoting_ballots_total",
		Help: "number of cast ballots",
	},
		[]string{"election"},
	)

	PromElectionShufflingInstances = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "evoting_shufflings_total",
		Help: "number of shuffling instances",
	},
		[]string{"election"},
	)

	PromElectionPubShares = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "evoting_pubshares_total",
		Help: "published public shares",
	},
		[]string{"election"},
	)

	PromDkgStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "evoting_dkg_status",
		Help: "status of DKG",
	},
		[]string{"election"},
	)
)

const (
	// ElectionsMetadataKey is the key at which election metadata are saved in
	// the storage.
	ElectionsMetadataKey = "ElectionsMetadataKey"
)

var suite = suites.MustFind("Ed25519")

const (
	// ContractName is the name of the contract.
	ContractName = "go.dedis.ch/dela.Evoting"

	// CmdArg is the argument's name to indicate the kind of command we want to
	// run on the contract. Should be one of the Command type.
	CmdArg = "evoting:command"

	// ElectionArg is the key at which the election argument is stored in the
	// transaction. The content is defined by the type of command.
	ElectionArg = "evoting:arg"

	// credentialAllCommand defines the credential command that is allowed to
	// perform all commands.
	credentialAllCommand = "all"
)

// commands defines the commands of the evoting contract. Using an interface
// helps in testing.
type commands interface {
	createElection(snap store.Snapshot, step execution.Step) error
	openElection(snap store.Snapshot, step execution.Step) error
	castVote(snap store.Snapshot, step execution.Step) error
	closeElection(snap store.Snapshot, step execution.Step) error
	shuffleBallots(snap store.Snapshot, step execution.Step) error
	registerPubshares(snap store.Snapshot, step execution.Step) error
	combineShares(snap store.Snapshot, step execution.Step) error
	cancelElection(snap store.Snapshot, step execution.Step) error
}

// Command defines a type of command for the value contract
type Command string

const (
	// CmdCreateElection is the command to create an election
	CmdCreateElection Command = "CREATE_ELECTION"
	// CmdOpenElection is the command to open an election
	CmdOpenElection Command = "OPEN_ELECTION"
	// CmdCastVote is the command to cast a vote
	CmdCastVote Command = "CAST_VOTE"
	// CmdCloseElection is the command to close an election
	CmdCloseElection Command = "CLOSE_ELECTION"
	// CmdShuffleBallots is the command to shuffle ballots
	CmdShuffleBallots Command = "SHUFFLE_BALLOTS"

	CmdRegisterPubShares Command = "REGISTER_PUB_SHARES"

	// CmdCombineShares is the command to decrypt ballots
	CmdCombineShares Command = "COMBINE_SHARES"
	// CmdCancelElection is the command to cancel an election
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

	cmd commands

	pedersen dkg.DKG

	rosterKey []byte

	context serde.Context

	electionFac    serde.Factory
	rosterFac      authority.Factory
	transactionFac serde.Factory
}

// NewContract creates a new Value contract
func NewContract(accessKey, rosterKey []byte, srvc access.Service,
	pedersen dkg.DKG, rosterFac authority.Factory) Contract {

	ctx := json.NewContext()

	ciphervoteFac := types.CiphervoteFactory{}
	electionFac := types.NewElectionFactory(ciphervoteFac, rosterFac)
	transactionFac := types.NewTransactionFactory(ciphervoteFac)

	contract := Contract{
		access:    srvc,
		accessKey: accessKey,
		pedersen:  pedersen,

		rosterKey: rosterKey,

		context: ctx,

		electionFac:    electionFac,
		rosterFac:      rosterFac,
		transactionFac: transactionFac,
	}

	contract.cmd = evotingCommand{Contract: &contract, prover: proof.HashVerify}

	return contract
}

// Execute implements native.Contract
func (c Contract) Execute(snap store.Snapshot, step execution.Step) error {
	creds := NewCreds(c.accessKey)

	err := c.access.Match(snap, creds, step.Current.GetIdentity())
	if err != nil {
		return xerrors.Errorf("identity not authorized: %v (%v)",
			step.Current.GetIdentity(), err)
	}

	cmd := step.Current.GetArg(CmdArg)
	if len(cmd) == 0 {
		return xerrors.Errorf("%q not found in tx arg", CmdArg)
	}

	switch Command(cmd) {
	case CmdCreateElection:
		err = c.cmd.createElection(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to create election: %v", err)
		}
	case CmdOpenElection:
		err := c.cmd.openElection(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to open election: %v", err)
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
	case CmdRegisterPubShares:
		err := c.cmd.registerPubshares(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to register the pubShares: %v", err)
		}
	case CmdCombineShares:
		err := c.cmd.combineShares(snap, step)
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

func init() {
	dvoting.PromCollectors = append(dvoting.PromCollectors,
		PromElectionStatus,
		PromElectionBallots,
		PromElectionShufflingInstances,
		PromElectionPubShares,
		PromDkgStatus)
}
