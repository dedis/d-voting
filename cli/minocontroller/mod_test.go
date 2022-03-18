package minocontroller

import (
	"crypto/elliptic"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/store/kv"
	"go.dedis.ch/dela/internal/testing/fake"
	"go.dedis.ch/dela/mino/minogrpc"
)

func TestMiniController_Build(t *testing.T) {
	ctrl := NewController()

	call := &fake.Call{}
	ctrl.SetCommands(fakeBuilder{call: call})

	require.Equal(t, 17, call.Len())
}

func TestMiniController_OnStart(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "minogrpc")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	db, err := kv.New(filepath.Join(dir, "test.db"))
	require.NoError(t, err)

	ctrl := NewController().(miniController)

	injector := node.NewInjector()
	injector.Inject(db)

	err = ctrl.OnStart(fakeContext{path: dir}, injector)
	require.NoError(t, err)

	var m *minogrpc.Minogrpc
	err = injector.Resolve(&m)
	require.NoError(t, err)
	require.NoError(t, m.GracefulStop())
}

func TestMiniController_InvalidPort_OnStart(t *testing.T) {
	ctrl := NewController()

	err := ctrl.OnStart(fakeContext{num: 100000}, node.NewInjector())
	require.EqualError(t, err, "invalid port value 100000")
}

func TestMiniController_MissingDB_OnStart(t *testing.T) {
	ctrl := NewController()

	err := ctrl.OnStart(fakeContext{}, node.NewInjector())
	require.EqualError(t, err, "injector: couldn't find dependency for 'kv.DB'")
}

func TestMiniController_FailGenerateKey_OnStart(t *testing.T) {
	ctrl := NewController().(miniController)
	ctrl.random = badReader{}

	inj := node.NewInjector()
	inj.Inject(fake.NewInMemoryDB())

	err := ctrl.OnStart(fakeContext{}, inj)
	require.EqualError(t, err,
		fake.Err("cert private key: while loading: generator failed: ecdsa"))
}

func TestMiniController_FailMarshalKey_OnStart(t *testing.T) {
	ctrl := NewController().(miniController)
	ctrl.curve = badCurve{Curve: elliptic.P224()}

	inj := node.NewInjector()
	inj.Inject(fake.NewInMemoryDB())

	err := ctrl.OnStart(fakeContext{}, inj)
	require.EqualError(t, err,
		"cert private key: while loading: generator failed: while marshaling: x509: unknown elliptic curve")
}

func TestMiniController_FailParseKey_OnStart(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "dela")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	ctrl := NewController().(miniController)

	inj := node.NewInjector()
	inj.Inject(fake.NewInMemoryDB())

	file, err := os.Create(filepath.Join(dir, certKeyName))
	require.NoError(t, err)

	defer file.Close()

	err = ctrl.OnStart(fakeContext{path: dir}, inj)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cert private key: while parsing: x509: ")
}

func TestMiniController_OnStop(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "minogrpc")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	db, err := kv.New(filepath.Join(dir, "test.db"))
	require.NoError(t, err)

	ctrl := NewController()

	injector := node.NewInjector()
	injector.Inject(db)

	err = ctrl.OnStart(fakeContext{path: dir}, injector)
	require.NoError(t, err)

	err = ctrl.OnStop(injector)
	require.NoError(t, err)
}

func TestMiniController_MissingMino_OnStop(t *testing.T) {
	ctrl := NewController()

	err := ctrl.OnStop(node.NewInjector())
	require.EqualError(t, err, "injector: couldn't find dependency for 'controller.StoppableMino'")
}

func TestMiniController_FailStopMino_OnStop(t *testing.T) {
	ctrl := NewController()

	inj := node.NewInjector()
	inj.Inject(badMino{})

	err := ctrl.OnStop(inj)
	require.EqualError(t, err, fake.Err("while stopping mino"))
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

type badMino struct {
	StoppableMino
}

func (badMino) GracefulStop() error {
	return fake.GetError()
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) {
	return 0, fake.GetError()
}

type badCurve struct {
	elliptic.Curve
}
