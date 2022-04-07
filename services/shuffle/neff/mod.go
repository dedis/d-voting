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
	mino        mino.Mino
	factory     serde.Factory
	service     ordering.Service
	p           pool.Pool
	blocks      *blockstore.InDisk
	context     serde.Context
	nodeSigner  crypto.Signer
	electionFac serde.Factory
}

// NewNeffShuffle returns a new NeffShuffle factory.
func NewNeffShuffle(m mino.Mino, s ordering.Service, p pool.Pool,
	blocks *blockstore.InDisk, electionFac serde.Factory, signer crypto.Signer) *NeffShuffle {

	factory := types.NewMessageFactory(m.GetAddressFactory())

	ctx := json.NewContext()

	return &NeffShuffle{
		mino:        m,
		factory:     factory,
		service:     s,
		p:           p,
		blocks:      blocks,
		context:     ctx,
		nodeSigner:  signer,
		electionFac: electionFac,
	}
}

// Listen implements shuffle.SHUFFLE. It must be called on each node that
// participates in the SHUFFLE. Creates the RPC.
func (n NeffShuffle) Listen(txmngr txn.Manager) (shuffle.Actor, error) {
	// We are expecting the manager to be exclusive for the service, with no
	// other use than us.
	err := txmngr.Sync()
	if err != nil {
		return nil, xerrors.Errorf("failed to sync manager: %v", err)
	}

	h := NewHandler(n.mino.GetAddress(), n.service, n.p, txmngr, n.nodeSigner,
		n.context, n.electionFac)

	a := &Actor{
		rpc:         mino.MustCreateRPC(n.mino, "shuffle", h, n.factory),
		factory:     n.factory,
		mino:        n.mino,
		service:     n.service,
		context:     n.context,
		electionFac: n.electionFac,
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

	context     serde.Context
	electionFac serde.Factory
}

// Shuffle must be called by ONE of the actor to shuffle the list of ElGamal
// pairs.
// Each node represented by a player must first execute Listen().
func (a *Actor) Shuffle(electionID []byte) error {
	a.Lock()
	defer a.Unlock()

	electionIDHex := hex.EncodeToString(electionID)

	election, err := getElection(a.electionFac, a.context, electionIDHex, a.service)
	if err != nil {
		return xerrors.Errorf("failed to get election: %v", err)
	}

	if election.Roster.Len() == 0 {
		return xerrors.Errorf("the roster is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), shuffleTimeout)
	defer cancel()

	sender, _, err := a.rpc.Stream(ctx, election.Roster)
	if err != nil {
		return xerrors.Errorf("failed to stream: %v", err)
	}

	addrs := make([]mino.Address, 0, election.Roster.Len())
	addrs = append(addrs, a.mino.GetAddress())
	addrIter := election.Roster.AddressIterator()
	for addrIter.HasNext() {
		addr := addrIter.GetNext()
		if !addr.Equal(a.mino.GetAddress()) {
			addrs = append(addrs, addr)
		}
	}

	dela.Logger.Info().Msgf("sending start shuffle to: %v", addrs)

	message := types.NewStartShuffle(electionIDHex, addrs)

	errs := sender.Send(message, addrs...)
	err = <-errs
	if err != nil {
		return xerrors.Errorf("failed to start shuffle: %v", err)
	}

	err = a.waitAndCheckShuffling(message.GetElectionId(), election.Roster.Len())
	if err != nil {
		return xerrors.Errorf("failed to wait and check shuffling: %v", err)
	}

	return nil
}

// waitAndCheckShuffling periodically checks the state of the election. It
// returns an error if the shuffling is not done after a while. The retry and
// waiting time depends on the rosterLen. electionID is Hex-encoded.
func (a *Actor) waitAndCheckShuffling(electionID string, rosterLen int) error {
	var election etypes.Election
	var err error

	for i := 0; i < rosterLen*10; i++ {
		election, err = getElection(a.electionFac, a.context, electionID, a.service)
		if err != nil {
			return xerrors.Errorf("failed to get election: %v", err)
		}

		round := len(election.ShuffleInstances)
		dela.Logger.Info().Msgf("SHUFFLE / ROUND : %d", round)

		// if the threshold is reached that means we have enough shuffling.
		if round >= election.ShuffleThreshold {
			dela.Logger.Info().Msgf("shuffle done with round n°%d", round)
			return nil
		}

		dela.Logger.Info().Msgf("waiting a while before checking election: %d", i)
		time.Sleep(time.Second * time.Duration(rosterLen/2+1))
	}

	return xerrors.Errorf("threshold of shuffling not reached: %d < %d",
		len(election.ShuffleInstances), election.ShuffleThreshold)
}

// getElection gets the election from the service.
func getElection(electionFac serde.Factory, ctx serde.Context,
	electionIDHex string, srv ordering.Service) (etypes.Election, error) {

	var election etypes.Election

	electionID, err := hex.DecodeString(electionIDHex)
	if err != nil {
		return election, xerrors.Errorf("failed to decode electionIDHex: %v", err)
	}

	proof, err := srv.GetProof(electionID)
	if err != nil {
		return election, xerrors.Errorf("failed to get proof: %v", err)
	}

	if string(proof.GetValue()) == "" {
		return election, xerrors.Errorf("election does not exist")
	}

	message, err := electionFac.Deserialize(ctx, proof.GetValue())
	if err != nil {
		return election, xerrors.Errorf("failed to deserialize Election: %v", err)
	}

	election, ok := message.(etypes.Election)
	if !ok {
		return election, xerrors.Errorf("wrong message type: %T", message)
	}

	if electionIDHex != election.ElectionID {
		return election, xerrors.Errorf("electionID do not match: %q != %q",
			electionIDHex, election.ElectionID)
	}

	return election, nil
}
