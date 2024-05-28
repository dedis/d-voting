package neff

import (
	"bytes"
	"context"
	"crypto/sha256"
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
	formFac       serde.Factory
}

// NewHandler creates a new handler
func NewHandler(me mino.Address, service ordering.Service, p pool.Pool,
	txmngr txn.Manager, shuffleSigner crypto.Signer, ctx serde.Context,
	formFac serde.Factory) *Handler {

	return &Handler{
		me:            me,
		service:       service,
		p:             p,
		txmngr:        txmngr,
		shuffleSigner: shuffleSigner,
		context:       ctx,
		formFac:       formFac,
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
		err := h.handleStartShuffle(msg.GetFormID(), msg.GetUserID())
		if err != nil {
			return xerrors.Errorf("failed to handle StartShuffle message: %v", err)
		}
	default:
		return xerrors.Errorf("expected StartShuffle message, got: %T", msg)
	}

	return nil
}

func (h *Handler) handleStartShuffle(formID string, userID string) error {
	dela.Logger.Info().Msg("Starting the neff shuffle protocol ...")

	err := h.txmngr.Sync()
	if err != nil {
		return xerrors.Errorf("failed to sync manager: %v", err.Error())
	}

	// loop until the threshold is reached or our transaction has been accepted
	for {
		form, err := etypes.FormFromStore(h.context, h.formFac, formID, h.service.GetStore())
		if err != nil {
			return xerrors.Errorf("failed to get form: %v", err)
		}

		round := len(form.ShuffleInstances)

		// check if the threshold is reached
		if round >= form.ShuffleThreshold {
			dela.Logger.Info().Msgf("shuffle done with round n°%d", round)
			return nil
		}

		if form.Status != etypes.Closed {
			return xerrors.Errorf("the form must be closed: (%v)", form.Status)
		}

		tx, err := h.makeTx(&form, userID)
		if err != nil {
			return xerrors.Errorf("failed to make tx: %v", err)
		}
		watchTimeout := 4 + form.ShuffleThreshold/2
		watchCtx, cancel := context.WithTimeout(context.Background(), time.Duration(watchTimeout)*time.Second)
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

		if accepted {
			dela.Logger.Info().Msg("our shuffling contribution has " +
				"been accepted, we are exiting the process")

			return nil
		}

		err = h.txmngr.Sync()
		if err != nil {
			return xerrors.Errorf("failed to sync manager: %v", err.Error())
		}

		dela.Logger.Info().Msg("shuffling contribution denied : " + msg)

		cancel()
	}
}

func (h *Handler) makeTx(form *etypes.Form, userID string) (txn.Transaction, error) {

	shuffledBallots, getProver, err := h.getShuffledBallots(form)
	if err != nil {
		return nil, xerrors.Errorf("failed to get shuffled ballots: %v", err)
	}

	shuffleBallots := etypes.ShuffleBallots{
		FormID:          form.FormID,
		UserID:          userID,
		Round:           len(form.ShuffleInstances),
		ShuffledBallots: shuffledBallots,
	}

	hash := sha256.New()

	err = shuffleBallots.Fingerprint(hash)
	if err != nil {
		return nil, xerrors.Errorf("failed to get fingerprint: %v", err)
	}

	seed := hash.Sum(nil)

	// Generate random vector and proof
	semiRandomStream, err := evoting.NewSemiRandomStream(seed)
	if err != nil {
		return nil, xerrors.Errorf("could not create semi-random stream: %v", err)
	}

	e := make([]kyber.Scalar, form.ChunksPerBallot())

	for i := 0; i < form.ChunksPerBallot(); i++ {
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
	signature, err := h.shuffleSigner.Sign(seed)
	if err != nil {
		return nil, xerrors.Errorf("could not sign the shuffle : %v", err)
	}

	encodedSignature, err := signature.Serialize(h.context)
	if err != nil {
		return nil, xerrors.Errorf("could not encode signature as []byte : %v ", err)
	}

	publicKey, err := h.shuffleSigner.GetPublicKey().MarshalBinary()
	if err != nil {
		return nil, xerrors.Errorf("could not unmarshal public key from nodeSigner: %v", err)
	}

	// Complete transaction:
	shuffleBallots.PublicKey = publicKey
	shuffleBallots.Signature = encodedSignature

	data, err := shuffleBallots.Serialize(h.context)
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
		Key:   evoting.FormArg,
		Value: data,
	}

	tx, err := h.txmngr.Make(args...)
	if err != nil {
		if err != nil {
			return nil, xerrors.Errorf("failed to use manager: %v", err.Error())
		}
	}

	return tx, nil
}

// getShuffledBallots returns the shuffled ballots with the shuffling proof.
func (h *Handler) getShuffledBallots(form *etypes.Form) ([]etypes.Ciphervote,
	func(e []kyber.Scalar) (proof.Prover, error), error) {

	round := len(form.ShuffleInstances)

	var ciphervotes []etypes.Ciphervote

	if round == 0 {
		suff, err := form.Suffragia(h.context, h.service.GetStore())
		if err != nil {
			return nil, nil, xerrors.Errorf("couldn't get ballots: %v", err)
		}
		ciphervotes = suff.Ciphervotes
	} else {
		ciphervotes = form.ShuffleInstances[round-1].ShuffledBallots
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
	XX, YY, getProver := shuffleKyber.SequencesShuffle(suite, nil, form.Pubkey,
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
