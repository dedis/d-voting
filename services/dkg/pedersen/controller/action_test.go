package controller

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

func TestInitAction_Execute(t *testing.T) {

	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    make(node.FlagSet),
		Out:      ioutil.Discard,
	}

	action := initAction{}

	err := action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve dkg: couldn't find dependency for 'dkg.DKG'")

	fakeErr := xerrors.Errorf("fake error")
	dkgPedersen := fakePedersen{err: fakeErr}
	ctx.Injector.Inject(dkgPedersen)
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to start the RPC: fake error")

	dkgPedersen.err = nil
	ctx.Injector.Inject(dkgPedersen)
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
	actor := fakeActor{err: fakeErr}
	ctx.Injector.Inject(actor)

	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to setup DKG: fake error")

	actor = fakeActor{pubKey: fakePubkey}
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

type fakePedersen struct {
	err error
}

func (f fakePedersen) Listen() (dkg.Actor, error) {
	return nil, f.err
}

func (f fakePedersen) GetLastActor() (dkg.Actor, error) {
	return nil, f.err
}

func (f fakePedersen) SetService(service ordering.Service) {
}

type fakeActor struct {
	err    error
	pubKey kyber.Point
}

func (f fakeActor) Setup(co crypto.CollectiveAuthority, threshold int) (pubKey kyber.Point, err error) {
	return f.pubKey, f.err
}

func (f fakeActor) GetPublicKey() (kyber.Point, error) {
	return f.pubKey, f.err
}

func (f fakeActor) Encrypt(message []byte) (K, C kyber.Point, remainder []byte, err error) {
	return nil, nil, nil, f.err
}

func (f fakeActor) Decrypt(K, C kyber.Point, electionId string) ([]byte, error) {
	return nil, f.err
}

func (f fakeActor) Reshare() error {
	return f.err
}

type fakeFlags struct {
	cli.Flags

	strings map[string][]string
}

func (f fakeFlags) StringSlice(name string) []string {
	return f.strings[name]
}
