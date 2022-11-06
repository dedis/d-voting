package neff

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/dela/serde/json"

	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/shuffle/neff/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var serdecontext serde.Context

var formFac serde.Factory
var transactionFac serde.Factory

func init() {
	ciphervoteFac := etypes.CiphervoteFactory{}
	formFac = etypes.NewFormFactory(ciphervoteFac, fakeAuthorityFactory{})
	transactionFac = etypes.NewTransactionFactory(ciphervoteFac)

	serdecontext = json.NewContext()
}

func TestNeffShuffle_Listen(t *testing.T) {

	NeffShuffle := NewNeffShuffle(fake.Mino{}, &fake.Service{}, &fake.Pool{}, nil, fakeAuthorityFactory{}, fake.NewSigner())

	actor, err := NeffShuffle.Listen(fake.Manager{})
	require.NoError(t, err)

	require.NotNil(t, actor)
}

func TestNeffShuffle_Shuffle(t *testing.T) {

	formID := "deadbeef"
	formIDBuf, err := hex.DecodeString(formID)
	require.NoError(t, err)

	rosterLen := 2
	roster := authority.FromAuthority(fake.NewAuthority(rosterLen, fake.NewSigner))

	form := fake.NewForm(formID)
	form.Roster = roster

	shuffledBallots := append([]etypes.Ciphervote{}, form.Suffragia.Ciphervotes...)
	form.ShuffleInstances = append(form.ShuffleInstances, etypes.ShuffleInstance{ShuffledBallots: shuffledBallots})

	form.ShuffleThreshold = 1

	service := fake.NewService(formID, form, serdecontext)

	actor := Actor{
		rpc:         fake.NewBadRPC(),
		mino:        fake.Mino{},
		service:     &service,
		context:     serdecontext,
		formFac: etypes.NewFormFactory(etypes.CiphervoteFactory{}, fake.NewRosterFac(roster)),
	}

	err = actor.Shuffle(formIDBuf)
	require.EqualError(t, err, fake.Err("failed to stream"))

	rpc := fake.NewStreamRPC(fake.NewReceiver(), fake.NewBadSender())
	actor.rpc = rpc

	oldLog := dela.Logger
	defer func() {
		dela.Logger = oldLog
	}()

	out := new(bytes.Buffer)
	dela.Logger = zerolog.New(out)

	// should only output a warning
	err = actor.Shuffle(formIDBuf)
	require.NoError(t, err)
	require.True(t, strings.Contains(out.String(), "failed to start shuffle"), out.String())

	rpc = fake.NewStreamRPC(fake.NewBadReceiver(), fake.Sender{})
	actor.rpc = rpc

	// we no longer use the receiver:
	err = actor.Shuffle(formIDBuf)
	require.NoError(t, err)

	recv := fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), types.NewEndShuffle()))

	rpc = fake.NewStreamRPC(recv, fake.Sender{})
	actor.rpc = rpc

	err = actor.Shuffle(formIDBuf)
	require.NoError(t, err)
}

// -----------------------------------------------------------------------------
// Utility functions

type fakeAuthorityFactory struct {
	serde.Factory
}

func (f fakeAuthorityFactory) AuthorityOf(ctx serde.Context, rosterBuf []byte) (authority.Authority, error) {
	fakeAuthority := &fakeAuthority{}
	return fakeAuthority, nil
}

type fakeAuthority struct {
	serde.Message
	serde.Fingerprinter
	crypto.CollectiveAuthority

	len int
}

func (f fakeAuthority) Apply(c authority.ChangeSet) authority.Authority {
	return nil
}

// Diff should return the change set to apply to get the given authority.
func (f fakeAuthority) Diff(a authority.Authority) authority.ChangeSet {
	return nil
}

func (f fakeAuthority) PublicKeyIterator() crypto.PublicKeyIterator {
	signers := make([]crypto.Signer, f.len)
	signers[0] = fake.NewSigner()

	return fake.NewPublicKeyIterator(signers)
}

func (f fakeAuthority) AddressIterator() mino.AddressIterator {
	addrs := make([]mino.Address, f.Len())
	for i := 0; i < f.Len(); i++ {
		addrs[i] = fake.NewAddress(i)
	}
	return fake.NewAddressIterator(addrs)
}

func (f fakeAuthority) Len() int {
	return f.len
}
