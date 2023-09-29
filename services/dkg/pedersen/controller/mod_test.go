package controller

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/c4dt/dela/core/txn/signed"
	"github.com/c4dt/dela/core/validation/simple"
	"github.com/c4dt/dela/crypto/bls"

	"github.com/c4dt/d-voting/internal/testing/fake"
	"github.com/c4dt/dela/cli/node"
	"github.com/c4dt/dela/core/access/darc"
	"github.com/c4dt/dela/core/execution/native"
	"github.com/c4dt/dela/core/ordering/cosipbft"
	"github.com/c4dt/dela/cosi/threshold"
	"github.com/stretchr/testify/require"
)

func TestMinimal_OnStart(t *testing.T) {
	c := NewController()

	flags := fakeFlags{strings: make(map[string]string)}

	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    flags,
		Out:      io.Discard,
	}

	// Should miss mino.Mino
	err := c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve mino.Mino")

	m := fake.Mino{}
	ctx.Injector.Inject(m)

	// Should miss *native.Service
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve *native.Service")

	exec := native.NewExecution()
	ctx.Injector.Inject(exec)

	// Should miss darc.Service
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve darc.Service")

	// ds := darc.NewService(json.NewContext())
	ctx.Injector.Inject(darc.Service{})

	// Should miss *threshold.Threshold
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve *threshold.Threshold")

	//th := threshold.NewThreshold(fake.Mino{}, fake.NewAggregateSigner())
	ctx.Injector.Inject(&threshold.Threshold{})

	// Should miss pool.Pool
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve p.Pool: couldn't find "+
		"dependency for 'pool.Pool'")

	ctx.Injector.Inject(&fake.Pool{})

	// Should miss authority.Factory
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve authority.Factory")

	ctx.Injector.Inject(fake.RosterFac{})

	// Should miss *cosipbft.Service
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "failed to resolve *cosipbft.Service")

	ctx.Injector.Inject(&cosipbft.Service{})

	ctx.Injector.Inject(&fake.InMemoryDB{})

	// Should miss flags
	err = c.OnStart(nil, ctx.Injector)
	require.EqualError(t, err, "no flags")

	dir, err := os.MkdirTemp(os.TempDir(), "dvoting1")
	require.NoError(t, err)
	flags.strings["config"] = dir

	signerFilePath := filepath.Join(dir, privateKeyFile)
	file, err := os.Create(signerFilePath)
	require.NoError(t, err)
	defer os.RemoveAll(signerFilePath)

	signer, err := bls.NewSigner().MarshalBinary()
	require.NoError(t, err)

	_, err = file.Write(signer)

	require.NoError(t, err)
	require.NoError(t, file.Close())

	// Should miss validation service to make client
	err = c.OnStart(flags, ctx.Injector)
	require.EqualError(t, err, "failed to make client: failed to resolve"+
		" validation.Service: couldn't find dependency for 'validation.Service'")

	// Should work now
	valService := simple.NewService(native.NewExecution(), signed.NewTransactionFactory())
	ctx.Injector.Inject(valService)

	err = c.OnStart(flags, ctx.Injector)

	require.NoError(t, err)
}

func TestMinimal_OnStop(t *testing.T) {
	c := NewController()

	err := c.OnStop(node.NewInjector())
	require.NoError(t, err)
}
