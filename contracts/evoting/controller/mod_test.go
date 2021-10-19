package controller

import (
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
)

func TestController_SetCommands(t *testing.T) {
	c := NewController()

	call := &fake.Call{}
	c.SetCommands(fakeBuilder{call: call})

	require.Equal(t, 13, call.Len())
	require.Equal(t, "e-voting", call.Get(0, 0))
	require.Equal(t, "interact with the evoting service", call.Get(1, 0))

	require.Equal(t, "serve", call.Get(2, 0))
	require.Equal(t, "Serve the HTTP server", call.Get(3, 0))
	require.Len(t, call.Get(4, 0), 1)
	require.IsType(t, &registerAction{}, call.Get(6, 0))
	require.Nil(t, call.Get(7, 0))

	require.Equal(t, "scenarioTest", call.Get(8, 0))
	require.Equal(t, "evoting scenario test", call.Get(9, 0))
	require.Len(t, call.Get(10, 0), 1)
	require.IsType(t, &scenarioTestAction{}, call.Get(11, 0))
	require.Nil(t, call.Get(12, 0))
}

func TestController_OnStart(t *testing.T) {
	err := NewController().OnStart(node.FlagSet{}, nil)
	require.Nil(t, err)
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
