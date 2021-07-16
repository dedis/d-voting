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

func TestController_SetCommands(t *testing.T) {
	c := NewController()

	call := &fake.Call{}
	c.SetCommands(fakeBuilder{call: call})

	require.Equal(t, 19, call.Len())
	require.Equal(t, "dkg", call.Get(0, 0))
	require.Equal(t, "interact with the DKG service", call.Get(1, 0))

	require.Equal(t, "init", call.Get(2, 0))
	require.Equal(t, "initialize the DKG protocol", call.Get(3, 0))
	require.IsType(t, &initAction{}, call.Get(4, 0))
	require.Nil(t, call.Get(5, 0))

	require.Equal(t, "setup", call.Get(6, 0))
	require.Equal(t, "creates the public distributed key and the private share on each node", call.Get(7, 0))
	require.Len(t, call.Get(8, 0), 1)
	require.IsType(t, &setupAction{}, call.Get(9, 0))
	require.Nil(t, call.Get(10, 0))

	require.Equal(t, "export", call.Get(11, 0))
	require.Equal(t, "export the node address and public key", call.Get(12, 0))
	require.IsType(t, &exportInfoAction{}, call.Get(13, 0))
	require.Nil(t, call.Get(14, 0))

	require.Equal(t, "getPublicKey", call.Get(15, 0))
	require.Equal(t, "prints the distributed public Key", call.Get(16, 0))
	require.IsType(t, &getPublicKeyAction{}, call.Get(17, 0))
	require.Nil(t, call.Get(18, 0))
}

func TestMinimal_OnStart(t *testing.T) {
	c := NewController()
	inj := newInjector(nil)

	err := c.OnStart(nil, inj)
	require.EqualError(t, err, fake.Err("failed to resolve mino"))

	inj = newInjector(fake.Mino{})
	err = c.OnStart(nil, inj)
	require.NoError(t, err)
	require.Len(t, inj.(*fakeInjector).history, 2)
	require.IsType(t, &pedersen.Pedersen{}, inj.(*fakeInjector).history[0])
	pubkey := suite.Point()
	require.IsType(t, pubkey, inj.(*fakeInjector).history[1])
}

func TestMinimal_OnStop(t *testing.T) {
	c := NewController()

	err := c.OnStop(node.NewInjector())
	require.NoError(t, err)
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

func newInjector(mino mino.Mino) node.Injector {
	return &fakeInjector{
		mino: mino,
	}
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
		return xerrors.Errorf("unkown message '%T", msg)
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
