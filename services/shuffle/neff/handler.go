package neff

import (
	"bytes"
	"context"
	"crypto/sha256"
	"math/rand"
	"time"

	"go.dedis.ch/kyber/v3"

	"github.com/dedis/d-voting/contracts/evoting"
	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/shuffle/neff/types"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3/proof"
	shuffleKyber "go.dedis.ch/kyber/v3/shuffle"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

const watchTimeout = time.Second * 6

var suite = suites.MustFind("Ed25519")

// Handler represents the RPC executed on each node
//
// - implements mino.Handler
type Handler struct {
	mino.UnsupportedHandler
	me            mino.Address
	service       ordering.Service
	p             pool.Pool
	txmngr        txn.Manager
	shuffleSigner crypto.Signer
	context       serde.Context
	electionFac   serde.Factory
}

// NewHandler creates a new handler
func NewHandler(me mino.Address, service ordering.Service, p pool.Pool,
	txmngr txn.Manager, shuffleSigner crypto.Signer, ctx serde.Context,
	electionFac serde.Factory) *Handler {

	return &Handler{
		me:            me,
		service:       service,
		p:             p,
		txmngr:        txmngr,
		shuffleSigner: shuffleSigner,
		context:       ctx,
		electionFac:   electionFac,
	}
}

// Stream implements mino.Handler. It allows one to stream messages to the
// players.
func (h *Handler) Stream(out mino.Sender, in mino.Receiver) error {

	from, msg, err := in.Recv(context.Background())
	if err != nil {
		return xerrors.Errorf("failed to receive: %v", err)
	}

	dela.Logger.Trace().Msgf("message received from: %v", from)

	switch msg := msg.(type) {
	case types.StartShuffle:
		err := h.handleStartShuffle(msg.GetElectionId())
		if err != nil {
			return xerrors.Errorf("failed to handle StartShuffle message: %v", err)
		}
	default:
		return xerrors.Errorf("expected StartShuffle message, got: %T", msg)
	}

	return nil
}

func (h *Handler) handleStartShuffle(electionID string) error {
	dela.Logger.Info().Msg("Starting the neff shuffle protocol ...")

	err := h.txmngr.Sync()
	if err != nil {
		return xerrors.Errorf("failed to sync manager: %v", err.Error())
	}

	// loop until the threshold is reached or our transaction has been accepted
	for {
		election, err := getElection(h.electionFac, h.context, electionID, h.service)
		if err != nil {
			return xerrors.Errorf("failed to get election: %v", err)
		}

		round := len(election.ShuffleInstances)

		// check if the threshold is reached
		if round >= election.ShuffleThreshold {
			dela.Logger.Info().Msgf("shuffle done with round n°%d", round)
			return nil
		}

		if election.Status != etypes.Closed {
			return xerrors.Errorf("the election must be closed: (%v)", election.Status)
		}

		tx, err := makeTx(h.context, &election, h.txmngr, h.shuffleSigner)
		if err != nil {
			return xerrors.Errorf("failed to make tx: %v", err)
		}

		watchCtx, cancel := context.WithTimeout(context.Background(), watchTimeout)
		defer cancel()

		events := h.service.Watch(watchCtx)

		err = h.p.Add(tx)
		if err != nil {
			// it is possible that an error is returned in case the nonce is not
			// synced. In that case we sync and retry.
			err = h.txmngr.Sync()
			if err != nil {
				dela.Logger.Warn().Err(err).Msgf("failed to add tx, syncing nonce")
				return xerrors.Errorf("failed to sync manager: %v", err.Error())
			}
		}

		accepted, msg := watchTx(events, tx.GetID())

		if !accepted {
			err = h.txmngr.Sync()
			if err != nil {
				return xerrors.Errorf("failed to sync manager: %v", err.Error())
			}
		}

		if accepted {
			dela.Logger.Info().Msg("our shuffling contribution has " +
				"been accepted, we are exiting the process")

			return nil
		}

		dela.Logger.Info().Msg("shuffling contribution denied : " + msg)

		cancel()
	}
}

func makeTx(ctx serde.Context, election *etypes.Election, manager txn.Manager,
	shuffleSigner crypto.Signer) (txn.Transaction, error) {

	shuffledBallots, getProver, err := getShuffledBallots(election)
	if err != nil {
		return nil, xerrors.Errorf("failed to get shuffled ballots: %v", err)
	}

	shuffleBallots := etypes.ShuffleBallots{
		ElectionID:      election.ElectionID,
		Round:           len(election.ShuffleInstances),
		ShuffledBallots: shuffledBallots,
	}

	h := sha256.New()

	err = shuffleBallots.Fingerprint(h)
	if err != nil {
		return nil, xerrors.Errorf("failed to get fingerprint: %v", err)
	}

	hash := h.Sum(nil)

	// Generate random vector and proof
	semiRandomStream, err := evoting.NewSemiRandomStream(hash)
	if err != nil {
		return nil, xerrors.Errorf("could not create semi-random stream: %v", err)
	}

	e := make([]kyber.Scalar, election.ChunksPerBallot())

	for i := 0; i < election.ChunksPerBallot(); i++ {
		v := suite.Scalar().Pick(semiRandomStream)
		e[i] = v
	}

	prover, err := getProver(e)
	if err != nil {
		return nil, xerrors.Errorf("could not get prover for shuffle : %v", err)
	}

	shuffleProof, err := proof.HashProve(suite, protocolName, prover)
	if err != nil {
		return nil, xerrors.Errorf("shuffle proof failed: %v", err)
	}

	shuffleBallots.Proof = shuffleProof
	shuffleBallots.RandomVector = etypes.RandomVector{}

	err = shuffleBallots.RandomVector.LoadFromScalars(e)
	if err != nil {
		return nil, xerrors.Errorf("could not marshal shuffle random vector")
	}

	// Sign the shuffle:
	signature, err := shuffleSigner.Sign(hash)
	if err != nil {
		return nil, xerrors.Errorf("could not sign the shuffle : %v", err)
	}

	encodedSignature, err := signature.Serialize(ctx)
	if err != nil {
		return nil, xerrors.Errorf("could not encode signature as []byte : %v ", err)
	}

	publicKey, err := shuffleSigner.GetPublicKey().MarshalBinary()
	if err != nil {
		return nil, xerrors.Errorf("could not unmarshal public key from nodeSigner: %v", err)
	}

	// Complete transaction:
	shuffleBallots.PublicKey = publicKey
	shuffleBallots.Signature = encodedSignature

	data, err := shuffleBallots.Serialize(ctx)
	if err != nil {
		return nil, xerrors.Errorf("failed to serialize shuffle ballots: %v", err)
	}

	args := make([]txn.Arg, 3)
	args[0] = txn.Arg{
		Key:   native.ContractArg,
		Value: []byte(evoting.ContractName),
	}
	args[1] = txn.Arg{
		Key:   evoting.CmdArg,
		Value: []byte(evoting.CmdShuffleBallots),
	}
	args[2] = txn.Arg{
		Key:   evoting.ElectionArg,
		Value: data,
	}
	rand.Seed(0)
	timeWait := rand.Intn(10)
	time.Sleep(time.Second * time.Duration(timeWait))

	tx, err := manager.Make(args...)
	if err != nil {
		if err != nil {
			return nil, xerrors.Errorf("failed to use manager: %v", err.Error())
		}
	}

	return tx, nil
}

// getShuffledBallots returns the shuffled ballots with the shuffling proof.
func getShuffledBallots(election *etypes.Election) ([]etypes.Ciphervote,
	func(e []kyber.Scalar) (proof.Prover, error), error) {

	round := len(election.ShuffleInstances)

	var ciphervotes []etypes.Ciphervote

	if round == 0 {
		ciphervotes = election.Suffragia.Ciphervotes
	} else {
		ciphervotes = election.ShuffleInstances[round-1].ShuffledBallots
	}

	seqSize := len(ciphervotes[0])

	X := make([][]kyber.Point, seqSize)
	Y := make([][]kyber.Point, seqSize)

	for _, ciphervote := range ciphervotes {

		x, y := ciphervote.GetElGPairs()

		for i := 0; i < seqSize; i++ {
			X[i] = append(X[i], x[i])
			Y[i] = append(Y[i], y[i])
		}
	}

	// shuffle sequences
	XX, YY, getProver := shuffleKyber.SequencesShuffle(suite, nil, election.Pubkey,
		X, Y, suite.RandomStream())

	ciphervotes, err := etypes.CiphervotesFromPairs(XX, YY)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to get ciphervotes: %v", err)
	}

	return ciphervotes, getProver, nil
}

// watchTx checks the transaction to find one that match txID. Return if the
// transaction has been accepted or not. Will also return false if/when the
// events chan is closed, which is expected to happen.
func watchTx(events <-chan ordering.Event, txID []byte) (bool, string) {
	for event := range events {
		for _, res := range event.Transactions {
			if !bytes.Equal(res.GetTransaction().GetID(), txID) {
				continue
			}

			dela.Logger.Info().Hex("id", txID).Msg("transaction included in the block")

			accepted, msg := res.GetStatus()
			if accepted {
				return true, ""
			}

			return false, msg
		}
	}

	return false, "watch timeout"
}
