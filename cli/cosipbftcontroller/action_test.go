package controller

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/core/txn/pool/mem"
	"go.dedis.ch/dela/core/validation"
	"go.dedis.ch/dela/cosi"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
)

func TestSetupAction_Execute(t *testing.T) {
	action := setupAction{}

	calls := &fake.Call{}
	ctx := prepContext(calls)
	ctx.Flags.(node.FlagSet)["member"] = []interface{}{"YQ==:YQ==", "YQ==:YQ=="}

	err := action.Execute(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, calls.Len())
	require.Equal(t, 2, calls.Get(0, 1).(mino.Players).Len())

	ctx.Flags.(node.FlagSet)["member"] = []interface{}{""}
	err = action.Execute(ctx)
	require.EqualError(t, err, "failed to read roster: failed to decode: invalid member base64 string")

	ctx.Flags = make(node.FlagSet)
	ctx.Injector = node.NewInjector()
	err = action.Execute(ctx)
	require.EqualError(t, err, "injector: couldn't find dependency for 'controller.Service'")

	ctx.Injector.Inject(fakeService{err: fake.GetError()})
	err = action.Execute(ctx)
	require.EqualError(t, err, fake.Err("failed to setup"))
}

func TestExportAction_Execute(t *testing.T) {
	action := exportAction{}

	ctx := prepContext(nil)

	buffer := new(bytes.Buffer)
	ctx.Out = buffer

	err := action.Execute(ctx)
	require.NoError(t, err)
	require.Equal(t, "AAAAAA==:UEs=", buffer.String())

	ctx.Injector = node.NewInjector()
	err = action.Execute(ctx)
	require.EqualError(t, err, "injector: couldn't find dependency for 'mino.Mino'")

	ctx.Injector.Inject(fake.NewBadMino())
	err = action.Execute(ctx)
	require.EqualError(t, err, fake.Err("failed to marshal address"))

	ctx.Injector.Inject(fake.Mino{})
	err = action.Execute(ctx)
	require.EqualError(t, err, "injector: couldn't find dependency for 'cosi.CollectiveSigning'")

	ctx.Injector.Inject(fakeCosi{err: true})
	err = action.Execute(ctx)
	require.EqualError(t, err, fake.Err("failed to marshal public key"))
}

func TestRosterAddAction_Execute(t *testing.T) {
	action := rosterAddAction{}

	ctx := prepContext(nil)
	ctx.Flags.(node.FlagSet)["member"] = "YQ==:YQ=="
	ctx.Flags.(node.FlagSet)["wait"] = float64(time.Second)

	err := action.Execute(ctx)
	require.NoError(t, err)

	ctx.Flags.(node.FlagSet)["wait"] = float64(0)

	err = action.Execute(ctx)
	require.NoError(t, err)

	var p pool.Pool
	require.NoError(t, ctx.Injector.Resolve(&p))
	require.Equal(t, 1, p.Len())

	ctx.Injector = node.NewInjector()
	err = action.Execute(ctx)
	require.EqualError(t, err, "injector: couldn't find dependency for 'controller.Service'")

	ctx.Injector.Inject(fakeService{err: fake.GetError()})
	err = action.Execute(ctx)
	require.EqualError(t, err, fake.Err("while preparing tx: failed to read roster"))

	ctx.Injector.Inject(fakeService{})
	err = action.Execute(ctx)
	require.EqualError(t, err,
		"while preparing tx: failed to decode member: injector: couldn't find dependency for 'mino.Mino'")

	ctx.Injector.Inject(fake.Mino{})
	ctx.Injector.Inject(fakeCosi{})
	err = action.Execute(ctx)
	require.EqualError(t, err,
		"while preparing tx: txn manager: injector: couldn't find dependency for 'txn.Manager'")

	ctx.Injector.Inject(fakeTxManager{errSync: fake.GetError()})
	err = action.Execute(ctx)
	require.EqualError(t, err, fake.Err("while preparing tx: txn manager: sync"))

	ctx.Injector.Inject(fakeTxManager{errMake: fake.GetError()})
	err = action.Execute(ctx)
	require.EqualError(t, err, fake.Err("while preparing tx: transaction: creating transaction"))

	ctx.Injector.Inject(fakeTxManager{})
	err = action.Execute(ctx)
	require.EqualError(t, err, "injector: couldn't find dependency for 'pool.Pool'")

	ctx.Injector.Inject(badPool{})
	err = action.Execute(ctx)
	require.EqualError(t, err, fake.Err("failed to add transaction"))

	events := []ordering.Event{
		{Transactions: []validation.TransactionResult{fakeResult{refused: true}}},
	}
	ctx = prepContext(nil)
	ctx.Flags.(node.FlagSet)["member"] = "YQ==:YQ=="
	ctx.Flags.(node.FlagSet)["wait"] = float64(time.Second)
	ctx.Injector.Inject(fakeService{events: events})
	err = action.Execute(ctx)
	require.EqualError(t, err, "transaction refused: message")

	ctx.Injector.Inject(fakeService{events: nil})
	err = action.Execute(ctx)
	require.EqualError(t, err, "transaction not found after timeout")
}

func TestDecodeMember(t *testing.T) {
	ctx := prepContext(nil)

	_, _, err := decodeMember(ctx, "a:a")
	require.EqualError(t, err, "base64 address: illegal base64 data at input byte 0")

	_, _, err = decodeMember(ctx, ":a")
	require.EqualError(t, err, "base64 public key: illegal base64 data at input byte 0")

	ctx.Injector = node.NewInjector()
	ctx.Injector.Inject(fake.Mino{})
	_, _, err = decodeMember(ctx, ":")
	require.EqualError(t, err, "injector: couldn't find dependency for 'cosi.CollectiveSigning'")

	ctx.Injector.Inject(fakeCosi{err: true})
	_, _, err = decodeMember(ctx, ":")
	require.EqualError(t, err, fake.Err("failed to decode public key"))
}

// -----------------------------------------------------------------------------
// Utility functions

func prepContext(calls *fake.Call) node.Context {
	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    make(node.FlagSet),
		Out:      io.Discard,
	}

	events := []ordering.Event{
		{Transactions: []validation.TransactionResult{fakeResult{}}},
	}

	ctx.Injector.Inject(fake.Mino{})
	ctx.Injector.Inject(fakeCosi{})
	ctx.Injector.Inject(fakeService{calls: calls, events: events})
	ctx.Injector.Inject(mem.NewPool())
	ctx.Injector.Inject(fakeTxManager{})

	return ctx
}

type fakeService struct {
	ordering.Service
	calls  *fake.Call
	events []ordering.Event
	err    error
}

func (s fakeService) GetRoster() (authority.Authority, error) {
	return authority.New(nil, nil), s.err
}

func (s fakeService) Setup(ctx context.Context, ca crypto.CollectiveAuthority) error {
	s.calls.Add(ctx, ca)
	return s.err
}

func (s fakeService) Watch(context.Context) <-chan ordering.Event {
	ch := make(chan ordering.Event, len(s.events))
	for _, evt := range s.events {
		ch <- evt
	}
	close(ch)

	return ch
}

func (s fakeService) Close() error {
	return s.err
}

type fakeCosi struct {
	cosi.CollectiveSigning
	err bool
}

func (c fakeCosi) GetPublicKeyFactory() crypto.PublicKeyFactory {
	if c.err {
		return fake.NewBadPublicKeyFactory()
	}

	return fake.NewPublicKeyFactory(fake.PublicKey{})
}

func (c fakeCosi) GetSigner() crypto.Signer {
	if c.err {
		return fake.NewSignerWithPublicKey(fake.NewBadPublicKey())
	}

	return fake.NewSigner()
}

type fakeTx struct {
	txn.Transaction
}

func (fakeTx) GetNonce() uint64 {
	return 0
}

func (fakeTx) GetIdentity() access.Identity {
	return fake.PublicKey{}
}

func (fakeTx) GetID() []byte {
	return []byte{0xaa}
}

type fakeResult struct {
	validation.TransactionResult
	refused bool
}

func (fakeResult) GetTransaction() txn.Transaction {
	return fakeTx{}
}

func (res fakeResult) GetStatus() (bool, string) {
	return !res.refused, "message"
}

type fakeTxManager struct {
	txn.Manager
	errMake error
	errSync error
}

func (mgr fakeTxManager) Make(args ...txn.Arg) (txn.Transaction, error) {
	return fakeTx{}, mgr.errMake
}

func (mgr fakeTxManager) Sync() error {
	return mgr.errSync
}

type badPool struct {
	pool.Pool
}

func (p badPool) Add(txn.Transaction) error {
	return fake.GetError()
}
