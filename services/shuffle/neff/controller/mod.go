package controller

import (
	"encoding"
	"path/filepath"

	etypes "github.com/c4dt/d-voting/contracts/evoting/types"
	"github.com/c4dt/d-voting/services/shuffle/neff"
	"github.com/c4dt/dela"
	"github.com/c4dt/dela/cli"
	"github.com/c4dt/dela/cli/node"
	"github.com/c4dt/dela/core/ordering"
	"github.com/c4dt/dela/core/ordering/cosipbft/authority"
	"github.com/c4dt/dela/core/ordering/cosipbft/blockstore"
	"github.com/c4dt/dela/core/txn/pool"
	"github.com/c4dt/dela/crypto"
	"github.com/c4dt/dela/crypto/bls"
	"github.com/c4dt/dela/crypto/loader"
	"github.com/c4dt/dela/mino"
	"golang.org/x/xerrors"
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

// Build implements node.Initializer.
func (m controller) SetCommands(builder node.Builder) {

	cmd := builder.SetCommand("shuffle")
	cmd.SetDescription("interact with the SHUFFLE service")

	sub := cmd.SetSubCommand("init")
	sub.SetFlags(cli.StringFlag{
		Name:     "signer",
		Usage:    "path to the private key",
		Required: true,
	})
	sub.SetDescription("initialize the SHUFFLE protocol")
	sub.SetAction(builder.MakeAction(&InitAction{}))

	sub = cmd.SetSubCommand("registerHandlers")
	sub.SetDescription("register the proxy handlers")
	sub.SetAction(builder.MakeAction(&RegisterHandlersAction{}))
}

// OnStart implements node.Initializer. It creates and registers a neff
// Shuffling.
func (m controller) OnStart(ctx cli.Flags, inj node.Injector) error {
	var no mino.Mino
	err := inj.Resolve(&no)
	if err != nil {
		return xerrors.Errorf("failed to resolve mino.Mino: %v", err)
	}

	var p pool.Pool
	err = inj.Resolve(&p)
	if err != nil {
		return xerrors.Errorf("failed to resolve pool.Pool: %v", err)
	}

	var service ordering.Service
	err = inj.Resolve(&service)
	if err != nil {
		return xerrors.Errorf("failed to resolve ordering.Service: %v", err)
	}

	var blocks *blockstore.InDisk
	err = inj.Resolve(&blocks)
	if err != nil {
		return xerrors.Errorf("failed to resolve blockstore.InDisk: %v", err)
	}

	var rosterFac authority.Factory
	err = inj.Resolve(&rosterFac)
	if err != nil {
		return xerrors.Errorf("failed to resolve authority.Factory")
	}

	signer, err := getNodeSigner(ctx)
	if err != nil {
		return xerrors.Errorf("failed to get Signer for the shuffle : %v", err)
	}

	neffShuffle := neff.NewNeffShuffle(no, service, p, blocks,
		etypes.NewFormFactory(etypes.CiphervoteFactory{}, rosterFac), signer)

	inj.Inject(neffShuffle)

	return nil
}

// OnStop implements node.Initializer.
func (controller) OnStop(node.Injector) error {
	return nil
}

// getSigner creates a signer from a file.
func getSigner(filePath string) (crypto.Signer, error) {
	l := loader.NewFileLoader(filePath)

	dela.Logger.Info().Msgf("loading private key from %q", filePath)

	signerData, err := l.Load()
	if err != nil {
		return nil, xerrors.Errorf("failed to load signer: %v", err)
	}

	signer, err := bls.NewSignerFromBytes(signerData)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal signer: %v", err)
	}

	return signer, nil
}

// getNodeSigner creates a signer with the node's private key
func getNodeSigner(flags cli.Flags) (crypto.AggregateSigner, error) {
	loader := loader.NewFileLoader(filepath.Join(flags.Path("config"), privateKeyFile))

	signerData, err := loader.LoadOrCreate(generator{newFn: blsSigner})
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
