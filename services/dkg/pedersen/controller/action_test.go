package controller

import (
	"encoding/hex"
	"io"
	"testing"

	"golang.org/x/xerrors"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/d-voting/internal/testing/fake"
	"go.dedis.ch/d-voting/services/dkg"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
)

func TestInitAction_Execute(t *testing.T) {

	flags := fakeFlags{strings: make(map[string]string)}

	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    flags,
		Out:      io.Discard,
	}

	action := initAction{}

	// Try without a DKG in the system
	err := action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve DKG: couldn't find dependency for 'dkg.DKG'")

	bp := fake.BadPedersen{Err: xerrors.Errorf("fake error")}
	ctx.Injector.Inject(bp)

	// Try without ordering service
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to make client: failed to resolve ordering.Service:"+
		" couldn't find dependency for 'ordering.Service'")

	// Try without validation service
	service := fake.Service{}
	ctx.Injector.Inject(&service)

	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to make client: failed to resolve validation.Service: couldn't find dependency for 'validation.Service'")

	// Try with a bogus DKG in the system
	bp = fake.BadPedersen{Err: xerrors.Errorf("fake error")}
	ctx.Injector.Inject(bp)

	valService := fake.ValidationService{}
	ctx.Injector.Inject(valService)

	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to start the RPC: fake error")

	ctx.Injector = node.NewInjector()
	ctx.Injector.Inject(&service)
	ctx.Injector.Inject(valService)
}

func TestSetupAction_Execute(t *testing.T) {
	action := setupAction{}

	flags := fakeFlags{strings: make(map[string]string)}
	inj := node.NewInjector()

	ctx := node.Context{
		Injector: inj,
		Out:      io.Discard,
	}

	formID := "deadbeef"

	flags.strings["member"] = "badAddress"
	flags.strings["formID"] = formID
	ctx.Flags = flags

	// No DKG
	err := action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve DKG: couldn't find dependency for 'dkg.DKG'")

	// DKG but not DKGMap

	// This implementation makes trivial Actors that
	// already have a public key
	p := fake.Pedersen{Actors: make(map[string]dkg.Actor)}

	formIDBuf, err := hex.DecodeString(formID)
	require.NoError(t, err)

	_, err = p.Listen(formIDBuf, fake.Manager{})
	require.NoError(t, err)

	inj.Inject(p)
}

func TestExportInfoAction_Execute(t *testing.T) {

	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    make(node.FlagSet),
		Out:      io.Discard,
	}

	action := exportInfoAction{}

	err := action.Execute(ctx)
	require.EqualError(t, err, "injector: couldn't find dependency for 'mino.Mino'")

	ctx.Injector.Inject(fake.NewBadMino())
	err = action.Execute(ctx)
	require.EqualError(t, err, fake.Err("failed to marshal address"))

	ctx.Injector.Inject(fake.Mino{})
	err = action.Execute(ctx)
	require.EqualError(t, err, "injector: couldn't find dependency for 'kv.DB'")

	db := fake.NewInMemoryDB()
	ctx.Injector.Inject(db)
	err = action.Execute(ctx)
	require.NoError(t, err)
}

// -----------------------------------------------------------------------------
// Utility functions

type fakeFlags struct {
	cli.Flags

	strings map[string]string
}

func (f fakeFlags) String(name string) string {
	return f.strings[name]
}

func (f fakeFlags) Path(name string) string {
	return f.String(name)
}
