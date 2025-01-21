package neff

import (
	"encoding/hex"
	"sync"
	"time"

	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/shuffle"
	"github.com/dedis/d-voting/services/shuffle/neff/types"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	"golang.org/x/net/context"
	"golang.org/x/xerrors"

	"go.dedis.ch/dela/serde/json"
)

const (
	shuffleTimeout = time.Second * 30
	protocolName   = "PairShuffle"
)

// NeffShuffle allows one to initialize a new SHUFFLE protocol.
//
// - implements shuffle.SHUFFLE
type NeffShuffle struct {
	mino       mino.Mino
	factory    serde.Factory
	service    ordering.Service
	p          pool.Pool
	blocks     *blockstore.InDisk
	context    serde.Context
	nodeSigner crypto.Signer
	formFac    serde.Factory
}

// NewNeffShuffle returns a new NeffShuffle factory.
func NewNeffShuffle(m mino.Mino, s ordering.Service, p pool.Pool,
	blocks *blockstore.InDisk, formFac serde.Factory, signer crypto.Signer) *NeffShuffle {

	factory := types.NewMessageFactory(m.GetAddressFactory())

	ctx := json.NewContext()

	return &NeffShuffle{
		mino:       m,
		factory:    factory,
		service:    s,
		p:          p,
		blocks:     blocks,
		context:    ctx,
		nodeSigner: signer,
		formFac:    formFac,
	}
}

// Listen implements shuffle.SHUFFLE. It must be called on each node that
// participates in the SHUFFLE. Creates the RPC.
func (n NeffShuffle) Listen(txmngr txn.Manager) (shuffle.Actor, error) {
	h := NewHandler(n.mino.GetAddress(), n.service, n.p, txmngr, n.nodeSigner,
		n.context, n.formFac)

	a := &Actor{
		rpc:     mino.MustCreateRPC(n.mino, "shuffle", h, n.factory),
		factory: n.factory,
		mino:    n.mino,
		service: n.service,
		context: n.context,
		formFac: n.formFac,
	}

	return a, nil
}

// Actor allows one to perform SHUFFLE operations like shuffling a list of
// ElGamal pairs and verify a shuffle
//
// - implements shuffle.Actor
type Actor struct {
	sync.Mutex
	rpc     mino.RPC
	mino    mino.Mino
	factory serde.Factory
	// startRes *state
	service ordering.Service

	context serde.Context
	formFac serde.Factory
}

// Shuffle must be called by ONE of the actors to shuffle the list of ElGamal
// pairs.
// Each node represented by a player must first execute Listen().
func (a *Actor) Shuffle(formID []byte, userID string) error {
	a.Lock()
	defer a.Unlock()

	formIDHex := hex.EncodeToString(formID)

	form, err := etypes.FormFromStore(a.context, a.formFac, formIDHex, a.service.GetStore())
	if err != nil {
		return xerrors.Errorf("failed to get form: %v", err)
	}

	if form.Roster.Len() == 0 {
		return xerrors.Errorf("the roster is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), shuffleTimeout)
	defer cancel()

	sender, _, err := a.rpc.Stream(ctx, form.Roster)
	if err != nil {
		return xerrors.Errorf("failed to stream: %v", err)
	}

	addrs := make([]mino.Address, 0, form.Roster.Len())
	addrs = append(addrs, a.mino.GetAddress())
	addrIter := form.Roster.AddressIterator()
	for addrIter.HasNext() {
		addr := addrIter.GetNext()
		if !addr.Equal(a.mino.GetAddress()) {
			addrs = append(addrs, addr)
		}
	}

	dela.Logger.Info().Msgf("sending start shuffle to: %v", addrs)

	message := types.NewStartShuffle(formIDHex, userID, addrs)

	errs := sender.Send(message, addrs...)
	err = <-errs
	if err != nil {
		dela.Logger.Warn().Msgf("failed to start shuffle: %v", err)
		//return xerrors.Errorf("failed to start shuffle: %v", err)
	}

	err = a.waitAndCheckShuffling(message.GetFormID(), form.Roster.Len())
	if err != nil {
		return xerrors.Errorf("failed to wait and check shuffling: %v", err)
	}

	return nil
}

// waitAndCheckShuffling periodically checks the state of the form. It
// returns an error if the shuffling is not done after a while. The retry and
// waiting time depends on the rosterLen. formID is Hex-encoded.
func (a *Actor) waitAndCheckShuffling(formID string, rosterLen int) error {
	var form etypes.Form
	var err error

	for i := 0; ; i++ {
		form, err = etypes.FormFromStore(a.context, a.formFac, formID, a.service.GetStore())
		if err != nil {
			return xerrors.Errorf("failed to get form: %v", err)
		}

		round := len(form.ShuffleInstances)
		dela.Logger.Info().Msgf("SHUFFLE / ROUND : %d", round)

		// if the threshold is reached that means we have enough shuffling.
		if round >= form.ShuffleThreshold {
			dela.Logger.Info().Msgf("shuffle done with round n°%d", round)
			return nil
		}

		dela.Logger.Info().Msgf("waiting a while before checking form: %d", i)
		sleepTime := rosterLen / 2
		time.Sleep(time.Duration(sleepTime) * time.Second)
		if i >= form.ShuffleThreshold*((int)(form.BallotCount)/16+1) {
			break
		}
		dela.Logger.Info().Msgf("WaitingRounds is : %d", form.ShuffleThreshold*((int)(form.BallotCount)/10+1))
	}

	return xerrors.Errorf("threshold of shuffling not reached: %d < %d",
		len(form.ShuffleInstances), form.ShuffleThreshold)
}
