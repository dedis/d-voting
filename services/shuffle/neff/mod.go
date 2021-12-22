package neff

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	evotingController "github.com/dedis/d-voting/contracts/evoting/controller"
	electionTypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/shuffle"
	"github.com/dedis/d-voting/services/shuffle/neff/types"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	shuffleKyber "go.dedis.ch/kyber/v3/shuffle"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/net/context"
	"golang.org/x/xerrors"

	jsonserde "go.dedis.ch/dela/serde/json"
)

const (
	shuffleTimeout = time.Second * 30
	protocolName   = "PairShuffle"

	// waitN is the number of time we check if the shuffling is done. Must be at
	// least one.
	waitN = 20
	// waitT is the time we wait before retrying to check if the shuffling is
	// done.
	waitT = time.Second * 3
)

// NeffShuffle allows one to initialize a new SHUFFLE protocol.
//
// - implements shuffle.SHUFFLE
type NeffShuffle struct {
	mino          mino.Mino
	factory       serde.Factory
	service       ordering.Service
	p             pool.Pool
	blocks        *blockstore.InDisk
	rosterFactory authority.Factory
	context       serde.Context
	nodeSigner    crypto.Signer
}

// NewNeffShuffle returns a new NeffShuffle factory.
func NewNeffShuffle(m mino.Mino, s ordering.Service, p pool.Pool,
	blocks *blockstore.InDisk, rosterFac authority.Factory, signer crypto.Signer) *NeffShuffle {

	factory := types.NewMessageFactory(m.GetAddressFactory())

	return &NeffShuffle{
		mino:          m,
		factory:       factory,
		service:       s,
		p:             p,
		blocks:        blocks,
		rosterFactory: rosterFac,
		context:       jsonserde.NewContext(),
		nodeSigner:    signer,
	}
}

// Listen implements shuffle.SHUFFLE. It must be called on each node that
// participates in the SHUFFLE. Creates the RPC.
func (n NeffShuffle) Listen(signer crypto.Signer) (shuffle.Actor, error) {
	client := &evotingController.Client{
		Nonce:  0,
		Blocks: n.blocks,
	}
	h := NewHandler(n.mino.GetAddress(), n.service, n.p, signer, client, n.nodeSigner)

	a := &Actor{
		rpc:       mino.MustCreateRPC(n.mino, "shuffle", h, n.factory),
		factory:   n.factory,
		mino:      n.mino,
		service:   n.service,
		rosterFac: n.rosterFactory,
		context:   n.context,
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

	rosterFac authority.Factory
	context   serde.Context
}

// Shuffle must be called by ONE of the actor to shuffle the list of ElGamal
// pairs.
// Each node represented by a player must first execute Listen().
func (a *Actor) Shuffle(electionID []byte) error {
	a.Lock()
	defer a.Unlock()

	proof, exists := electionExists(a.service, electionID)
	if !exists {
		return xerrors.Errorf("election %s was not found", electionID)
	}

	election := new(electionTypes.Election)
	err := json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal Election: %v", err)
	}

	roster, err := a.rosterFac.AuthorityOf(a.context, election.RosterBuf)
	if err != nil {
		return xerrors.Errorf("failed to deserialize roster: %v", err)
	}

	if roster.Len() == 0 {
		return xerrors.Errorf("the roster is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), shuffleTimeout)
	defer cancel()

	sender, _, err := a.rpc.Stream(ctx, roster)
	if err != nil {
		return xerrors.Errorf("failed to stream: %v", err)
	}

	addrs := make([]mino.Address, 0, roster.Len())
	addrs = append(addrs, a.mino.GetAddress())

	addrIter := roster.AddressIterator()
	for addrIter.HasNext() {
		addr := addrIter.GetNext()
		if !addr.Equal(a.mino.GetAddress()) {
			addrs = append(addrs, addr)
		}
	}

	message := types.NewStartShuffle(hex.EncodeToString(electionID), addrs)

	errs := sender.Send(message, addrs...)
	err = <-errs
	if err != nil {
		return xerrors.Errorf("failed to start shuffle: %v", err)
	}

	err = a.waitAndCheckShuffling(message.GetElectionId())
	if err != nil {
		return xerrors.Errorf("failed to wait and check shuffling: %v", err)
	}

	return nil
}

// waitAndCheckShuffling periodically checks the state of the election. It
// returns an error if the shuffling is not done after a while.
func (a *Actor) waitAndCheckShuffling(electionID string) error {
	var election *electionTypes.Election
	var err error

	for i := 0; i < waitN; i++ {
		election, err = getElection(a.service, electionID)
		if err != nil {
			return xerrors.Errorf("failed to get election: %v", err)
		}

		round := len(election.ShuffleInstances)
		dela.Logger.Info().Msgf("SHUFFLE / ROUND : %d", round)

		// if the threshold is reached that means we have enough
		// shuffling.
		if round >= election.ShuffleThreshold {
			dela.Logger.Info().Msgf("shuffle done with round n°%d", round)
			return nil
		}

		dela.Logger.Info().Msgf("waiting a while before checking election: %d", i)
		time.Sleep(waitT)
	}

	return xerrors.Errorf("threshold of shuffling not reached: %d < %d",
		len(election.ShuffleInstances), election.ShuffleThreshold)
}

// Todo : this is useless in the new implementation, maybe remove ?

// Verify allows to verify a Shuffle
func (a *Actor) Verify(suiteName string, Ks []kyber.Point, Cs []kyber.Point,
	pubKey kyber.Point, KsShuffled []kyber.Point, CsShuffled []kyber.Point, prf []byte) (err error) {

	suite := suites.MustFind(suiteName)

	verifier := shuffleKyber.Verifier(suite, nil, pubKey, Ks, Cs, KsShuffled, CsShuffled)
	return proof.HashVerify(suite, protocolName, verifier, prf)
}

func electionExists(service ordering.Service, electionIDBuf []byte) (ordering.Proof, bool) {
	proof, err := service.GetProof(electionIDBuf)
	if err != nil {
		return proof, false
	}

	// this is proof of absence
	if string(proof.GetValue()) == "" {
		return proof, false
	}

	return proof, true
}
