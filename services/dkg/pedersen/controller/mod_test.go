package controller

import (
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg/pedersen"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/mino"
	"golang.org/x/xerrors"
)

func TestMinimal_OnStart(t *testing.T) {
	c := NewController()
	inj := newInjector(nil)

	err := c.OnStart(nil, inj)
	require.EqualError(t, err, fake.Err("failed to resolve mino"))

	inj = newInjector(fake.Mino{})
	err = c.OnStart(nil, inj)
	require.NoError(t, err)

	// var exec *native.Service

	// var access darc.Service

	// var cosi *threshold.Threshold

	// var rosterFac authority.Factory

	// var srvc *cosipbft.Service

	require.Len(t, inj.(*fakeInjector).history, 2)
	require.IsType(t, &pedersen.Pedersen{}, inj.(*fakeInjector).history[0])

	// Weird, one line is enough?
	pubkey := suite.Point()
	require.IsType(t, pubkey, inj.(*fakeInjector).history[1])
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
