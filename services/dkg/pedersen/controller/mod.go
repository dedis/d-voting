package controller

import (
	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/services/dkg/pedersen"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/access/darc"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering/cosipbft"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/cosi/threshold"
	"go.dedis.ch/dela/mino"
	"golang.org/x/xerrors"
)

// evotingAccessKey is the access key used for the evoting contract.
var evotingAccessKey = [32]byte{3}

// NewController returns a new controller initializer
func NewController() node.Initializer {
	return controller{}
}

// controller is an initializer with a set of commands.
//
// - implements node.Initializer
type controller struct{}

// Build implements node.Initializer.
func (m controller) SetCommands(builder node.Builder) {

	cmd := builder.SetCommand("dkg")
	cmd.SetDescription("interact with the DKG service")

	sub := cmd.SetSubCommand("init")
	sub.SetDescription("initialize the DKG protocol")
	sub.SetAction(builder.MakeAction(&initAction{}))

	// memcoin --config /tmp/node1 dkg setup --member $(memcoin --config
	// /tmp/node1 dkg export) --member $(memcoin --config /tmp/node2 dkg export)
	sub = cmd.SetSubCommand("setup")
	sub.SetDescription("creates the public distributed key and the private share on each node")
	sub.SetFlags(cli.StringFlag{
		Name:     "electionID",
		Usage:    "the election ID, formatted in hexadecimal",
		Required: true,
	})
	sub.SetAction(builder.MakeAction(&setupAction{}))

	sub = cmd.SetSubCommand("export")
	sub.SetDescription("export the node address and public key")
	sub.SetAction(builder.MakeAction(&exportInfoAction{}))

	sub = cmd.SetSubCommand("getPublicKey")
	sub.SetDescription("prints the distributed public Key")
	sub.SetAction(builder.MakeAction(&getPublicKeyAction{}))

	sub = cmd.SetSubCommand("registerHandlers")
	sub.SetDescription("register the proxy handlers")
	sub.SetAction(builder.MakeAction(&registerHandlersAction{}))
}

// OnStart implements node.Initializer. It creates and registers a pedersen DKG.
func (m controller) OnStart(ctx cli.Flags, inj node.Injector) error {
	var no mino.Mino
	err := inj.Resolve(&no)
	if err != nil {
		return xerrors.Errorf("failed to resolve mino: %v", err)
	}

	var exec *native.Service
	err = inj.Resolve(&exec)
	if err != nil {
		return xerrors.Errorf("failed to resolve *native.Service")
	}

	var access darc.Service
	err = inj.Resolve(&access)
	if err != nil {
		return xerrors.Errorf("failed to resolve dac.Service")
	}

	var cosi *threshold.Threshold
	err = inj.Resolve(&cosi)
	if err != nil {
		return xerrors.Errorf("failed to resolve *threshold.Threshold")
	}

	var rosterFac authority.Factory
	err = inj.Resolve(&rosterFac)
	if err != nil {
		return xerrors.Errorf("failed to resolve authority.Factory")
	}

	var srvc *cosipbft.Service
	err = inj.Resolve(&srvc)
	if err != nil {
		return xerrors.Errorf("failed to resolve *cosipbft.Service")
	}

	dkg, pubkey := pedersen.NewPedersen(no, true, srvc, rosterFac)

	pubkeyBuf, err := pubkey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to encode pubkey: %v", err)
	}

	inj.Inject(dkg)
	inj.Inject(pubkey)

	dela.Logger.Info().
		Hex("public key", pubkeyBuf).
		Msg("perdersen public key")

	rosterKey := [32]byte{}
	evoting.RegisterContract(exec, evoting.NewContract(evotingAccessKey[:], rosterKey[:], access, dkg, rosterFac))

	return nil
}

// OnStop implements node.Initializer.
func (controller) OnStop(node.Injector) error {
	return nil
}
