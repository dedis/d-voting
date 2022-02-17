package controller

import (
	"go.dedis.ch/dela/crypto/bls"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/access/darc"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering/cosipbft"
	"go.dedis.ch/dela/cosi/threshold"
)

func TestMinimal_OnStart(t *testing.T) {
	c := NewController()

	flags := fakeFlags{strings: make(map[string]string)}

	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    flags,
		Out:      ioutil.Discard,
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

	dir, err := ioutil.TempDir(os.TempDir(), "memcoin1")
	require.NoError(t, err)
	flags.strings["config"] = dir

	signerFilePath := filepath.Join(dir, privateKeyFile)
	file, err := os.Create(signerFilePath)
	require.NoError(t, err)

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
	// TODO: Inject validation service

	err = os.Remove(signerFilePath)
	require.NoError(t, err)
}

func TestMinimal_OnStop(t *testing.T) {
	c := NewController()

	err := c.OnStop(node.NewInjector())
	require.NoError(t, err)
}
