package controller

import (
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"go.dedis.ch/dela/core/txn/pool"
)

func TestController_SetCommands(t *testing.T) {
	c := NewController()

	call := &fake.Call{}
	c.SetCommands(fakeBuilder{call: call})

	require.Equal(t, 7, call.Len())
	require.Equal(t, "shuffle", call.Get(0, 0))
	require.Equal(t, "interact with the SHUFFLE service", call.Get(1, 0))
	require.Equal(t, "init", call.Get(2, 0))
	require.Len(t, call.Get(3, 0), 1)
	require.Equal(t, "initialize the SHUFFLE protocol", call.Get(4, 0))
	require.IsType(t, &initAction{}, call.Get(5, 0))
	require.Nil(t, call.Get(6, 0))

}

func TestController_OnStart(t *testing.T) {
	c := NewController()

	inj := node.NewInjector()

	err := c.OnStart(make(node.FlagSet), inj)
	require.EqualError(t, err,
		"failed to resolve mino.Mino: couldn't find dependency for 'mino.Mino'")

	inj.Inject(fake.Mino{})
	err = c.OnStart(make(node.FlagSet), inj)

	require.EqualError(t, err,
		"failed to resolve pool.Pool: couldn't find dependency for 'pool.Pool'")

	inj.Inject(fakePool{})
	err = c.OnStart(make(node.FlagSet), inj)
	require.EqualError(t, err,
		"failed to resolve ordering.Service: couldn't find dependency for 'ordering.Service'")

	inj.Inject(fakeService{})
	err = c.OnStart(make(node.FlagSet), inj)
	require.EqualError(t, err,
		"failed to resolve blockstore.InDisk: couldn't find dependency for '*blockstore.InDisk'")

	inj.Inject(&blockstore.InDisk{})
	err = c.OnStart(make(node.FlagSet), inj)
	require.EqualError(t, err,
		"failed to resolve authority.Factory")
}

func TestController_OnStop(t *testing.T) {
	err := NewController().OnStop(nil)
	require.Nil(t, err)
}

// -----------------------------------------------------------------------------
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

type fakePool struct {
	pool.Pool
}

type fakeService struct {
	ordering.Service
}
