package controller

import (
	"encoding/hex"
	"go.dedis.ch/dela/core/txn/signed"
	"io/ioutil"
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/store/kv"
)

func TestInitAction_Execute(t *testing.T) {

	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    make(node.FlagSet),
		Out:      ioutil.Discard,
	}

	action := initAction{}

	// Try without a DKG in the system
	err := action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve DKG: couldn't find dependency for 'dkg.DKG'")

	//// Try without a signer TODO: unwillingly uses the signer created in mod_test.go
	//bp := fake.BadPedersen{Err: xerrors.Errorf("fake error")}
	//ctx.Injector.Inject(bp)
	//err = action.Execute(ctx)
	//require.EqualError(t, err, "failed to get signer: failed to load signer:" +
	//	" while opening file: open : The system cannot find the file specified.")

	//// Try with a bogus DKG in the system
	//bp = fake.BadPedersen{Err: xerrors.Errorf("fake error")}
	//ctx.Injector.Inject(bp)
	//err = action.Execute(ctx)
	//require.EqualError(t, err, "failed to start the RPC: fake error")
	//
	//ctx.Injector = node.NewInjector()
	//
	//// Try with a DKG but no DKGMap in the system
	//p := fake.Pedersen{Actors: make(map[string]dkg.Actor)}
	//ctx.Injector.Inject(p)
	//err = action.Execute(ctx)
	//require.EqualError(t, err, "failed to update DKG store: failed to resolve db: "+
	//	"couldn't find dependency for 'kv.DB'")
	//
	//ctx.Injector = node.NewInjector()
	//
	//// Try with a DKG and a DKGMap in the system
	//p.Actors = make(map[string]dkg.Actor)
	//ctx.Injector.Inject(p)
	//db := fake.NewInMemoryDB()
	//ctx.Injector.Inject(db)
	//
	//err = action.Execute(ctx)
	//require.NoError(t, err)
}

func TestSetupAction_Execute(t *testing.T) {
	action := setupAction{}

	flags := fakeFlags{strings: make(map[string]string)}
	inj := node.NewInjector()

	ctx := node.Context{
		Injector: inj,
		Out:      ioutil.Discard,
	}

	electionID := "deadbeef"

	flags.strings["member"] = "badAddress"
	flags.strings["electionID"] = electionID
	ctx.Flags = flags

	// No DKG
	err := action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve DKG: couldn't find dependency for 'dkg.DKG'")

	// DKG but not DKGMap

	// This implementation makes trivial Actors that
	// already have a public key
	p := fake.Pedersen{Actors: make(map[string]dkg.Actor)}

	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	a, err := p.Listen(electionIDBuf, signed.NewManager(fake.NewSigner(), &client{}))
	require.NoError(t, err)

	inj.Inject(p)

	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to update DKG store: failed to resolve db: "+
		"couldn't find dependency for 'kv.DB'")

	// DKG and DKGMap
	db := fake.NewInMemoryDB()
	ctx.Injector.Inject(db)

	err = action.Execute(ctx)
	require.NoError(t, err)

	// Check that the map contains the actor
	err = db.View(func(tx kv.ReadableTx) error {
		bucket := tx.GetBucket([]byte(BucketName))
		require.NotNil(t, bucket)

		pubKeyBuf := bucket.Get(electionIDBuf)
		pubKeyRes := suite.Point()
		err = pubKeyRes.UnmarshalBinary(pubKeyBuf)
		require.NoError(t, err)

		pubKey := a.(fake.DKGActor).PubKey

		require.True(t, pubKeyRes.Equal(pubKey))

		return nil
	})
	require.NoError(t, err)
}

func TestExportInfoAction_Execute(t *testing.T) {

	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    make(node.FlagSet),
		Out:      ioutil.Discard,
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
