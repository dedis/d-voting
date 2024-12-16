package controller

import (
	"encoding"
	"path/filepath"

	"go.dedis.ch/d-voting/contracts/evoting"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/bls"
	"go.dedis.ch/dela/crypto/loader"

	"go.dedis.ch/d-voting/services/dkg/pedersen"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/access/darc"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering/cosipbft"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/store/kv"
	"go.dedis.ch/dela/cosi/threshold"
	"go.dedis.ch/dela/mino"
	"golang.org/x/xerrors"

	etypes "go.dedis.ch/d-voting/contracts/evoting/types"
)

const privateKeyFile = "private.key"

// NewController returns a new controller initializer
func NewController() node.Initializer {
	return controller{}
}

// controller is an initializer with a set of commands.
//
// - implements node.Initializer
type controller struct{}

// SetCommands implements node.Initializer.
func (m controller) SetCommands(builder node.Builder) {

	formIDFlag := cli.StringFlag{
		Name:     "formID",
		Usage:    "the form ID, formatted in hexadecimal",
		Required: true,
	}

	cmd := builder.SetCommand("dkg")
	cmd.SetDescription("interact with the DKG service")

	// dvoting --config /tmp/node1 dkg init --formID formID
	sub := cmd.SetSubCommand("init")
	sub.SetDescription("initialize the DKG protocol for a given form")
	sub.SetFlags(formIDFlag)
	sub.SetAction(builder.MakeAction(&initAction{}))

	// dvoting --config /tmp/node1 dkg setup --formID formID
	sub = cmd.SetSubCommand("setup")
	sub.SetDescription("create the public distributed key and the private share on each node")
	sub.SetFlags(formIDFlag)
	sub.SetAction(builder.MakeAction(&setupAction{}))

	sub = cmd.SetSubCommand("export")
	sub.SetDescription("export the node address and public key")
	sub.SetAction(builder.MakeAction(&exportInfoAction{}))

	sub = cmd.SetSubCommand("getPublicKey")
	sub.SetDescription("print the distributed public key")
	sub.SetFlags(formIDFlag)
	sub.SetAction(builder.MakeAction(&getPublicKeyAction{}))

	sub = cmd.SetSubCommand("registerHandlers")
	sub.SetDescription("register the proxy handlers")
	sub.SetAction(builder.MakeAction(&RegisterHandlersAction{}))
}

// OnStart implements node.Initializer. It creates and registers a pedersen DKG.
func (m controller) OnStart(ctx cli.Flags, inj node.Injector) error {
	var no mino.Mino
	err := inj.Resolve(&no)
	if err != nil {
		return xerrors.Errorf("failed to resolve mino.Mino")
	}

	var exec *native.Service
	err = inj.Resolve(&exec)
	if err != nil {
		return xerrors.Errorf("failed to resolve *native.Service")
	}

	var access darc.Service
	err = inj.Resolve(&access)
	if err != nil {
		return xerrors.Errorf("failed to resolve darc.Service")
	}

	var cosi *threshold.Threshold
	err = inj.Resolve(&cosi)
	if err != nil {
		return xerrors.Errorf("failed to resolve *threshold.Threshold")
	}

	var p pool.Pool
	err = inj.Resolve(&p)
	if err != nil {
		return xerrors.Errorf("failed to resolve p.Pool: %v", err)
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

	if ctx == nil {
		return xerrors.Errorf("no flags")
	}
	if ctx.String("config") == "" {
		return xerrors.Errorf("no config path")
	}

	var db kv.DB

	err = inj.Resolve(&db)
	if err != nil {
		return xerrors.Errorf("failed to resolve db: %v", err)
	}

	signer, err := getSigner(ctx)
	if err != nil {
		return xerrors.Errorf("failed to get a signer for the pubShares: %v",
			err)
	}

	client, err := makeClient(inj)
	if err != nil {
		return xerrors.Errorf("failed to make client: %v", err)
	}

	formFac := etypes.NewFormFactory(etypes.CiphervoteFactory{}, rosterFac)

	dkg := pedersen.NewPedersen(no, srvc, db, p, formFac, signer)

	// Use dkgMap to fill the actors map
	err = dkg.ReadActors(signed.NewManager(signer, &client))
	if err != nil {
		return xerrors.Errorf("database read failed: %v", err)
	}

	inj.Inject(dkg)

	c := evoting.NewContract(access, dkg, rosterFac)
	evoting.RegisterContract(exec, c)

	return nil
}

// OnStop implements node.Initializer.
func (controller) OnStop(node.Injector) error {
	return nil
}

// getSigner creates a signer with the node's private key
func getSigner(flags cli.Flags) (crypto.AggregateSigner, error) {
	fileLoader := loader.NewFileLoader(filepath.Join(flags.Path("config"), privateKeyFile))

	signerData, err := fileLoader.LoadOrCreate(generator{newFn: blsSigner})
	if err != nil {
		return nil, xerrors.Errorf("while loading: %v", err)
	}

	signer, err := bls.NewSignerFromBytes(signerData)
	if err != nil {
		return nil, xerrors.Errorf("while unmarshaling: %v", err)
	}

	return signer, nil
}

// generator is an implementation to generate a private key.
//
// - implements loader.Generator
type generator struct {
	newFn func() encoding.BinaryMarshaler
}

// Generate implements loader.Generator. It returns the marshaled data of a
// private key.
func (g generator) Generate() ([]byte, error) {
	signer := g.newFn()

	data, err := signer.MarshalBinary()
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal signer: %v", err)
	}

	return data, nil
}

// blsSigner is a wrapper to use a signer with the primitives to use a BLS
// signature
func blsSigner() encoding.BinaryMarshaler {
	return bls.NewSigner()
}
