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

	// Register the JSON format for the form
	_ "github.com/dedis/d-voting/contracts/evoting/json"
)

var (
	PromFormStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dvoting_status",
		Help: "status of form",
	},
		[]string{"form"},
	)

	PromFormBallots = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dvoting_ballots_total",
		Help: "number of cast ballots",
	},
		[]string{"form"},
	)

	PromFormShufflingInstances = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dvoting_shufflings_total",
		Help: "number of shuffling instances",
	},
		[]string{"form"},
	)

	PromFormPubShares = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dvoting_pubshares_total",
		Help: "published public shares",
	},
		[]string{"form"},
	)

	PromFormDkgStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dvoting_dkg_status",
		Help: "status of distributed key generator",
	},
		[]string{"form"},
	)
)

const (
	// FormsMetadataKey is the key at which form metadata are saved in
	// the storage.
	FormsMetadataKey = "FormsMetadataKey"
)

var suite = suites.MustFind("Ed25519")

const (
	// ContractUID is the UID of the contract
	ContractUID = "EVOT"

	// ContractName is the name of the contract.
	ContractName = "go.dedis.ch/dela.Evoting"

	// CmdArg is the argument's name to indicate the kind of command we want to
	// run on the contract. Should be one of the Command type.
	CmdArg = "evoting:command"

	// FormArg is the key at which the form argument is stored in the
	// transaction. The content is defined by the type of command.
	FormArg = "evoting:arg"

	// credentialAllCommand defines the credential command that is allowed to
	// perform all commands.
	credentialAllCommand = "all"
)

// commands defines the commands of the evoting contract. Using an interface
// helps in testing.
type commands interface {
	createForm(snap store.Snapshot, step execution.Step) error
	openForm(snap store.Snapshot, step execution.Step) error
	castVote(snap store.Snapshot, step execution.Step) error
	closeForm(snap store.Snapshot, step execution.Step) error
	shuffleBallots(snap store.Snapshot, step execution.Step) error
	registerPubshares(snap store.Snapshot, step execution.Step) error
	combineShares(snap store.Snapshot, step execution.Step) error
	cancelForm(snap store.Snapshot, step execution.Step) error
	deleteForm(snap store.Snapshot, step execution.Step) error
}

// Command defines a type of command for the value contract
type Command string

const (
	// CmdCreateForm is the command to create a form
	CmdCreateForm Command = "CREATE_FORM"
	// CmdOpenForm is the command to open a form
	CmdOpenForm Command = "OPEN_FORM"
	// CmdCastVote is the command to cast a vote
	CmdCastVote Command = "CAST_VOTE"
	// CmdCloseForm is the command to close a form
	CmdCloseForm Command = "CLOSE_FORM"
	// CmdShuffleBallots is the command to shuffle ballots
	CmdShuffleBallots Command = "SHUFFLE_BALLOTS"

	// CmdRegisterPubShares is the command to register the pubshares
	CmdRegisterPubShares Command = "REGISTER_PUB_SHARES"

	// CmdCombineShares is the command to decrypt ballots
	CmdCombineShares Command = "COMBINE_SHARES"
	// CmdCancelForm is the command to cancel a form
	CmdCancelForm Command = "CANCEL_FORM"

	// CmdDeleteForm is the command to delete a form
	CmdDeleteForm Command = "DELETE_FORM"
)

// NewCreds creates new credentials for a evoting contract execution. We might
// want to use in the future a separate credential for each command.
func NewCreds() access.Credential {
	return access.NewContractCreds([]byte(ContractUID), ContractName, credentialAllCommand)
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

	cmd commands

	pedersen dkg.DKG

	context serde.Context

	formFac        serde.Factory
	rosterFac      authority.Factory
	transactionFac serde.Factory
}

// NewContract creates a new Value contract
func NewContract(srvc access.Service,
	pedersen dkg.DKG, rosterFac authority.Factory) Contract {

	ctx := json.NewContext()

	ciphervoteFac := types.CiphervoteFactory{}
	formFac := types.NewFormFactory(ciphervoteFac, rosterFac)
	transactionFac := types.NewTransactionFactory(ciphervoteFac)

	contract := Contract{
		access:   srvc,
		pedersen: pedersen,

		context: ctx,

		formFac:        formFac,
		rosterFac:      rosterFac,
		transactionFac: transactionFac,
	}

	contract.cmd = evotingCommand{Contract: &contract, prover: proof.HashVerify}

	return contract
}

// Execute implements native.Contract
func (c Contract) Execute(snap store.Snapshot, step execution.Step) error {
	creds := NewCreds()

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
	case CmdCreateForm:
		err = c.cmd.createForm(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to create form: %v", err)
		}
	case CmdOpenForm:
		err := c.cmd.openForm(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to open form: %v", err)
		}
	case CmdCastVote:
		err := c.cmd.castVote(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to cast vote: %v", err)
		}
	case CmdCloseForm:
		err := c.cmd.closeForm(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to close form: %v", err)
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
	case CmdCancelForm:
		err := c.cmd.cancelForm(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to cancel form: %v", err)
		}
	case CmdDeleteForm:
		err := c.cmd.deleteForm(snap, step)
		if err != nil {
			return xerrors.Errorf("failed to delete form: %v", err)
		}
	default:
		return xerrors.Errorf("unknown command: %s", cmd)
	}

	return nil
}

// UID returns the unique 4-bytes contract identifier.
//
// - implements native.Contract
func (c Contract) UID() string {
	return ContractUID
}

func init() {
	dvoting.PromCollectors = append(dvoting.PromCollectors,
		PromFormStatus,
		PromFormBallots,
		PromFormShufflingInstances,
		PromFormPubShares)
}
