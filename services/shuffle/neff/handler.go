package neff

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	evotingController "github.com/dedis/d-voting/contracts/evoting/controller"
	electionTypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/shuffle/neff/types"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	jsondela "go.dedis.ch/dela/serde/json"
	"go.dedis.ch/kyber/v3/proof"
	shuffleKyber "go.dedis.ch/kyber/v3/shuffle"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

const watchTimeout = time.Second * 2

// const endShuffleTimeout = time.Second * 50

var suite = suites.MustFind("Ed25519")

// Handler represents the RPC executed on each node
//
// - implements mino.Handler
type Handler struct {
	mino.UnsupportedHandler
	me            mino.Address
	service       ordering.Service
	p             pool.Pool
	signer        crypto.Signer
	client        *evotingController.Client
	shuffleSigner crypto.Signer
}

// NewHandler creates a new handler
func NewHandler(me mino.Address, service ordering.Service, p pool.Pool,
	signer crypto.Signer, client *evotingController.Client, shuffleSigner crypto.Signer) *Handler {
	return &Handler{
		me:            me,
		service:       service,
		p:             p,
		signer:        signer,
		client:        client,
		shuffleSigner: shuffleSigner,
	}
}

// Stream implements mino.Handler. It allows one to stream messages to the
// players.
func (h *Handler) Stream(out mino.Sender, in mino.Receiver) error {

	from, msg, err := in.Recv(context.Background())
	if err != nil {
		return xerrors.Errorf("failed to receive: %v", err)
	}

	dela.Logger.Info().Msgf("message received from: %v", from)

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
	manager := getManager(h.signer, h.client)

	// loop until the threshold is reached or our transaction has been accepted
	for {
		election, err := getElection(h.service, electionID)
		if err != nil {
			return xerrors.Errorf("failed to get election: %v", err)
		}

		round := len(election.ShuffleInstances)

		// check if the threshold is reached
		if round >= election.ShuffleThreshold {
			dela.Logger.Info().Msgf("shuffle done with round n°%d", round)
			return nil
		}

		if election.Status != electionTypes.Closed {
			return xerrors.Errorf("the election must be closed: %s", election.Status)
		}

		tx, err := makeTx(election, manager, h.shuffleSigner)
		if err != nil {
			return xerrors.Errorf("failed to make tx: %v", err)
		}

		watchCtx, cancel := context.WithTimeout(context.Background(), watchTimeout)
		defer cancel()

		events := h.service.Watch(watchCtx)

		dela.Logger.Info().Msgf("sending shuffling tx with nonce %d", tx.GetNonce())

		err = h.p.Add(tx)
		if err != nil {
			return xerrors.Errorf("failed to add transaction to the pool: %v", err.Error())
		}

		accepted, msg := watchTx(events, tx.GetID())
		if accepted {
			dela.Logger.Info().Msg("our shuffling contribution has " +
				"been accepted, we are exitting the process")
			return nil
		}

		dela.Logger.Info().Msg("shuffling contribution denied : " + msg)
	}
}

func makeTx(election *electionTypes.Election, manager txn.Manager, shuffleSigner crypto.Signer) (txn.Transaction, error) {
	shuffledBallots, shuffleProof, err := getShuffledBallots(election)
	if err != nil {
		return nil, xerrors.Errorf("failed to get shuffled ballots: %v", err)
	}

	shuffleBallotsTransaction := electionTypes.ShuffleBallotsTransaction{
		ElectionID:      string(election.ElectionID),
		Round:           len(election.ShuffleInstances),
		ShuffledBallots: shuffledBallots,
		Proof:           shuffleProof,
	}

	//Sign the shuffle:
	shuffleHash, err := electionTypes.HashShuffle(shuffleBallotsTransaction, election.ElectionID)
	if err != nil {
		return nil, xerrors.Errorf("Could not hash the shuffle while creating transaction: %v", err)
	}

	signature, err := shuffleSigner.Sign(shuffleHash)
	if err != nil {
		return nil, xerrors.Errorf("Could not sign the shuffle : %v", err)
	}

	encodedSignature, err := signature.Serialize(jsondela.NewContext())
	if err != nil {
		return nil, xerrors.Errorf("Could not encode signature as []byte : %v ", err)
	}

	publicKey, err := shuffleSigner.GetPublicKey().MarshalBinary()
	if err != nil {
		return nil, xerrors.Errorf("Could not unmarshal public key from nodeSigner: %v", err)
	}

	//Complete transaction:
	shuffleBallotsTransaction.PublicKey = publicKey
	shuffleBallotsTransaction.Signature = encodedSignature

	js, err := json.Marshal(shuffleBallotsTransaction)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal "+
			"ShuffleBallotsTransaction: %v", err.Error())
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
		Key:   evoting.ShuffleBallotsArg,
		Value: js,
	}

	err = manager.Sync()
	if err != nil {
		return nil, xerrors.Errorf("failed to sync manager: %v", err.Error())
	}

	tx, err := manager.Make(args...)
	if err != nil {
		return nil, xerrors.Errorf("failed to make transaction: %v", err.Error())
	}

	return tx, nil
}

// getShuffledBallots returns the shuffled ballots with the shuffling proof.
func getShuffledBallots(election *electionTypes.Election) (electionTypes.Ciphertexts, []byte, error) {
	round := len(election.ShuffleInstances)

	var encryptedBallots electionTypes.Ciphertexts

	if round == 0 {
		encryptedBallots = election.EncryptedBallots.Ballots
		fmt.Println("encrypted ballots:", election.EncryptedBallots)
	} else {
		encryptedBallots = election.ShuffleInstances[round-1].ShuffledBallots
	}

	ks, cs, err := encryptedBallots.GetKsCs()
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to get ks, cs: %v", err)
	}

	pubKey := suite.Point()

	err = pubKey.UnmarshalBinary(election.Pubkey)
	if err != nil {
		return nil, nil, xerrors.Errorf("couldn't unmarshal public key: %v", err)
	}

	rand := suite.RandomStream()
	Kbar, Cbar, prover := shuffleKyber.Shuffle(suite, nil, pubKey, ks, cs, rand)

	shuffleProof, err := proof.HashProve(suite, protocolName, prover)
	if err != nil {
		return nil, nil, xerrors.Errorf("Shuffle proof failed: %v", err.Error())
	}

	var shuffledBallots electionTypes.Ciphertexts

	err = shuffledBallots.InitFromKsCs(Kbar, Cbar)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to init ciphertexts: %v", err)
	}

	return shuffledBallots, shuffleProof, nil
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

// getElection returns the election state from the global state.
func getElection(service ordering.Service, electionID string) (*electionTypes.Election, error) {
	electionIDBuff, err := hex.DecodeString(electionID)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode election id: %v", err)
	}

	prf, err := service.GetProof(electionIDBuff)
	if err != nil {
		return nil, xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election := new(electionTypes.Election)

	err = json.NewDecoder(bytes.NewBuffer(prf.GetValue())).Decode(election)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal Election: %v", err)
	}
	return election, nil
}

// getManager is the function called when we need a transaction manager. It
// allows us to use a different manager for the tests.
var getManager = func(signer crypto.Signer, s signed.Client) txn.Manager {
	return signed.NewManager(signer, s)
}
