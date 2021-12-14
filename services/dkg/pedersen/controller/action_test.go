package controller

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"golang.org/x/xerrors"
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

	// Try with a bogus DKG in the system
	p := fake.Pedersen{Err: xerrors.Errorf("fake error")}
	ctx.Injector.Inject(p)
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to start the RPC: fake error")

	// Try with a DKG but no DKGMap in the system
	p.Err = nil
	ctx.Injector.Inject(p)
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve dkgMap: couldn't find dependency for 'pedersen.Store'")

	// Try with a DKG and a DKGMap in the system
	var dkgMap fake.PedersenStore
	ctx.Injector.Inject(dkgMap)

	err = action.Execute(ctx)
	require.NoError(t, err)
}

func TestSetupAction_Execute(t *testing.T) {
	action := setupAction{}

	flags := fakeFlags{strings: make(map[string][]string)}
	inj := node.NewInjector()

	ctx := node.Context{
		Injector: inj,
		Out:      ioutil.Discard,
	}

	flags.strings["member"] = []string{"badAddress"}
	flags.strings["electionID"] = []string{"deadbeef"}
	ctx.Flags = flags

	err := action.Execute(ctx)
	require.EqualError(t, err, "failed to read roster: failed to decode: invalid member base64 string 'badAddress'")

	flags.strings["member"] = []string{"badAddress:badKey"}
	ctx.Flags = flags
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to read roster: failed to decode: injector: couldn't find dependency for 'mino.Mino'")

	inj.Inject(fake.Mino{})
	ctx.Injector = inj
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to read roster: failed to decode: base64 address: illegal base64 data at input byte 8")

	addrBuf, _ := fake.Mino{}.GetAddress().MarshalText()
	base64address := base64.StdEncoding.EncodeToString(addrBuf)
	flags.strings["member"] = []string{base64address + ":badKey"}
	ctx.Flags = flags
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to read roster: failed to decode: base64 public key: illegal base64 data at input "+
		"byte 4")

	badKeyBuf := []byte("badKey")
	base64key := base64.StdEncoding.EncodeToString(badKeyBuf)
	flags.strings["member"] = []string{base64address + ":" + base64key}
	ctx.Flags = flags
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to read roster: failed to decode: failed to decode public key: failed to unmarshal "+
		"the key: couldn't unmarshal point: invalid Ed25519 curve point")

	fakePubkey := suite.Point()
	fakeKeyBuf, _ := fakePubkey.MarshalBinary()
	base64key = base64.StdEncoding.EncodeToString(fakeKeyBuf)
	flags.strings["member"] = []string{base64address + ":" + base64key}
	ctx.Flags = flags
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve actor: couldn't find dependency for 'dkg.Actor'")

	fakeErr := xerrors.Errorf("fake error")
	actor := fake.DKGActor{Err: fakeErr}
	ctx.Injector.Inject(actor)

	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to setup DKG: fake error")

	actor = fake.DKGActor{PubKey: fakePubkey}
	ctx.Injector.Inject(actor)

	err = action.Execute(ctx)
	require.NoError(t, err)
}

func TestExportInfoAction_Execute(t *testing.T) {
	inj := node.NewInjector()

	ctx := node.Context{
		Injector: inj,
		Flags:    make(node.FlagSet),
		Out:      ioutil.Discard,
	}

	action := exportInfoAction{}

	err := action.Execute(ctx)
	require.EqualError(t, err, "injector: couldn't find dependency for 'mino.Mino'")

	inj.Inject(fake.NewBadMino())
	ctx.Injector = inj
	err = action.Execute(ctx)
	require.EqualError(t, err, fake.Err("failed to marshal address"))

	inj.Inject(fake.Mino{})
	ctx.Injector = inj
	err = action.Execute(ctx)
	require.EqualError(t, err, "injector: couldn't find dependency for 'kyber.Point'")

	pubKey := suite.Point()
	inj.Inject(pubKey)
	ctx.Injector = inj
	err = action.Execute(ctx)
	require.NoError(t, err)

	// todo : check context writer
}

// -----------------------------------------------------------------------------
// Utility functions

type fakeFlags struct {
	cli.Flags

	strings map[string][]string
}

func (f fakeFlags) StringSlice(name string) []string {
	return f.strings[name]
}
