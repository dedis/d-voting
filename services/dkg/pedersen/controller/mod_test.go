package controller

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/access/darc"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/cosi/threshold"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"golang.org/x/xerrors"
)

func TestMinimal_OnStart(t *testing.T) {
	c := NewController()

	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    make(node.FlagSet),
		Out:      ioutil.Discard,
	}

	// Should miss mino.Mino
	err := c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve mino.Mino")

	ctx.Injector.Inject(fake.Mino{})

	// Should miss *native.Service
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve *native.Service")

	ctx.Injector.Inject(native.NewExecution())

	// Should miss darc.Service
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve darc.Service")

	ctx.Injector.Inject(darc.Service{})

	// Should miss *threshold.Threshold
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve *threshold.Threshold")

	th := threshold.NewThreshold(fake.Mino{}, fake.NewAggregateSigner())
	ctx.Injector.Inject(th)

	// Should miss authority.Factory
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve authority.Factory")

	ctx.Injector.Inject(fake.RosterFac{})

	// Should miss *cosipbft.Service
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve *cosipbft.Service")

	rosterLen := 2
	roster := authority.FromAuthority(fake.NewAuthority(rosterLen, fake.NewSigner))
	ctx.Injector.Inject(fakeService{roster: roster})

	// Should work
	err = c.OnStart(nil, ctx.Injector)
	require.NoError(t, err)

	// pubkey := suite.Point()
	// require.IsType(t, pubkey, inj.(*fakeInjector).history[1])
}

func TestMinimal_OnStop(t *testing.T) {
	c := NewController()

	err := c.OnStop(node.NewInjector())
	require.NoError(t, err)
}

// Utility functions

type fakeCommandBuilder struct {
	call *fake.Call
}

func (b fakeCommandBuilder) SetSubCommand(name string) cli.CommandBuilder {
	b.call.Add(name)
	return b
}

func (b fakeCommandBuilder) SetDescription(value string) {
	b.call.Add(value)
}

func (b fakeCommandBuilder) SetFlags(flags ...cli.Flag) {
	b.call.Add(flags)
}

func (b fakeCommandBuilder) SetAction(a cli.Action) {
	b.call.Add(a)
}

type fakeBuilder struct {
	call *fake.Call
}

func (b fakeBuilder) SetCommand(name string) cli.CommandBuilder {
	b.call.Add(name)
	return fakeCommandBuilder(b)
}

func (b fakeBuilder) SetStartFlags(flags ...cli.Flag) {
	b.call.Add(flags)
}

func (b fakeBuilder) MakeAction(tmpl node.ActionTemplate) cli.Action {
	b.call.Add(tmpl)
	return nil
}

// fakeInjector is a fake injector
//
// - implements node.Injector
type fakeInjector struct {
	isBad   bool
	mino    mino.Mino
	history []interface{}
}

// Resolve implements node.Injector
func (i fakeInjector) Resolve(el interface{}) error {
	if i.isBad {
		return fake.GetError()
	}

	switch msg := el.(type) {
	case *mino.Mino:
		if i.mino == nil {
			return fake.GetError()
		}
		*msg = i.mino
	default:
		return xerrors.Errorf("unkown message: %T", msg)
	}

	return nil
}

// Inject implements node.Injector
func (i *fakeInjector) Inject(v interface{}) {
	if i.history == nil {
		i.history = make([]interface{}, 0)
	}
	i.history = append(i.history, v)
}

func newInjector(args ...interface{}) node.Injector {
	f := &fakeInjector{}
	for i := range args {
		f.Inject(i)
	}
	return f
}

func newInjectorFromMino(mino mino.Mino) node.Injector {
	return &fakeInjector{mino: mino}
}

// fakeService is a fake service
//
// - implements cosipbft.Service
type fakeService struct {
	ordering.Service

	roster authority.Roster
}

func (f fakeService) GetRoster() (authority.Authority, error) {
	return f.roster, nil
}

func (f fakeService) Setup(ctx context.Context, ca crypto.CollectiveAuthority) error {
	return nil
}
