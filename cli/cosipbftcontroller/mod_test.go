package controller

import (
	"encoding"
	"os"
	"path/filepath"
	"testing"

	"github.com/dedis/d-voting/services/dkg"
	"go.dedis.ch/dela/core/ordering"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/store/kv"
	"go.dedis.ch/dela/core/txn/pool"
)

func TestMinimal_SetCommands(t *testing.T) {
	m := NewController()

	b := node.NewBuilder()
	m.SetCommands(b)
}

func TestMinimal_OnStart(t *testing.T) {
	flags, dir, clean := makeFlags(t)
	defer clean()

	db, err := kv.New(filepath.Join(dir, "test.db"))
	require.NoError(t, err)

	m := NewController().(miniController)

	inj := node.NewInjector()
	inj.Inject(fake.Mino{})
	inj.Inject(db)
	inj.Inject(fakeDKG{})

	err = m.OnStart(flags, inj)
	require.NoError(t, err)
}

func TestMinimal_MissingMino_OnStart(t *testing.T) {
	m := NewController()

	err := m.OnStart(make(node.FlagSet), node.NewInjector())
	require.EqualError(t, err,
		"injector: couldn't find dependency for 'mino.Mino'")
}

func TestMinimal_FailLoadKey_OnStart(t *testing.T) {
	flags, _, clean := makeFlags(t)
	defer clean()

	m := NewController().(miniController)

	inj := node.NewInjector()
	inj.Inject(fake.Mino{})
	inj.Inject(fake.NewInMemoryDB())
	inj.Inject(fakeDKG{})

	m.signerFn = badFn

	err := m.OnStart(flags, inj)
	require.EqualError(t, err,
		fake.Err("signer: while loading: generator failed: failed to marshal signer"))
}

func TestMinimal_MissingDB_OnStart(t *testing.T) {
	flags, _, clean := makeFlags(t)
	defer clean()

	m := NewController().(miniController)

	inj := node.NewInjector()
	inj.Inject(fake.Mino{})
	inj.Inject(fakeDKG{})

	err := m.OnStart(flags, inj)
	require.EqualError(t, err, "injector: couldn't find dependency for 'kv.DB'")
}

func TestMinimal_MalformedKey_OnStart(t *testing.T) {
	flags, dir, clean := makeFlags(t)
	defer clean()

	m := NewController().(miniController)

	inj := node.NewInjector()
	inj.Inject(fake.Mino{})
	inj.Inject(fake.NewInMemoryDB())
	inj.Inject(fakeDKG{})

	file, err := os.Create(filepath.Join(dir, privateKeyFile))
	require.NoError(t, err)

	file.Close()

	err = m.OnStart(flags, inj)
	require.Error(t, err)
	require.Contains(t, err.Error(), "signer: while unmarshaling: ")
}

func TestMinimal_OnStop(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "dela-test-")
	require.NoError(t, err)

	db, err := kv.New(filepath.Join(dir, "test.db"))
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	m := NewController()

	fset := make(node.FlagSet)
	fset["config"] = dir

	inj := node.NewInjector()
	inj.Inject(fake.Mino{})
	inj.Inject(db)
	inj.Inject(fakeDKG{})

	err = m.OnStart(fset, inj)
	require.NoError(t, err)

	err = m.OnStop(inj)
	require.NoError(t, err)

	inj = node.NewInjector()
	err = m.OnStop(inj)
	require.EqualError(t, err,
		"injector: couldn't find dependency for 'ordering.Service'")

	inj.Inject(fakeService{err: fake.GetError()})
	err = m.OnStop(inj)
	require.EqualError(t, err, fake.Err("while closing service"))

	inj.Inject(fakeService{})
	err = m.OnStop(inj)
	require.EqualError(t, err,
		"injector: couldn't find dependency for 'pool.Pool'")

	inj.Inject(fakePool{err: fake.GetError()})
	err = m.OnStop(inj)
	require.EqualError(t, err, fake.Err("while closing pool"))
}

// -----------------------------------------------------------------------------
// Utility functions

func makeFlags(t *testing.T) (cli.Flags, string, func()) {
	dir, err := os.MkdirTemp(os.TempDir(), "dela-")
	require.NoError(t, err)

	fset := make(node.FlagSet)
	fset["config"] = dir

	return fset, dir, func() { os.RemoveAll(dir) }
}

func badFn() encoding.BinaryMarshaler {
	return fake.NewBadHash()
}

type fakeDKG struct {
	err error
}

func (f fakeDKG) Listen() (dkg.Actor, error) {
	return nil, f.err
}

func (f fakeDKG) GetLastActor() (dkg.Actor, error) {
	return nil, f.err
}

func (f fakeDKG) SetService(service ordering.Service) {
}

type fakePool struct {
	pool.Pool

	err error
}

func (p fakePool) Close() error {
	return p.err
}
