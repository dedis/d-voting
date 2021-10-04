package neff

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
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
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/proof"
	shuffleKyber "go.dedis.ch/kyber/v3/shuffle"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

const shuffleTransactionTimeout = time.Second * 1

// const endShuffleTimeout = time.Second * 50

var suite = suites.MustFind("Ed25519")

// Handler represents the RPC executed on each node
//
// - implements mino.Handler
type Handler struct {
	mino.UnsupportedHandler
	me      mino.Address
	service ordering.Service
	p       pool.Pool
	signer  crypto.Signer
	client  *evotingController.Client
}

// NewHandler creates a new handler
func NewHandler(me mino.Address, service ordering.Service, p pool.Pool,
	signer crypto.Signer, client *evotingController.Client) *Handler {
	return &Handler{
		me:      me,
		service: service,
		p:       p,
		signer:  signer,
		client:  client,
	}
}

// Stream implements mino.Handler. It allows one to stream messages to the
// players.
func (h *Handler) Stream(out mino.Sender, in mino.Receiver) error {

	from, msg, err := in.Recv(context.Background())
	if err != nil {
		return xerrors.Errorf("failed to receive: %v", err)
	}

	switch msg := msg.(type) {

	case types.StartShuffle:
		err := h.HandleStartShuffleMessage(msg, from, out, in)
		if err != nil {
			return xerrors.Errorf("failed to handle StartShuffle message: %v", err)
		}

	default:
		return xerrors.Errorf("expected StartShuffle message, got: %T", msg)
	}

	return nil
}

func (h *Handler) HandleStartShuffleMessage(startShuffleMessage types.StartShuffle, from mino.Address, out mino.Sender,
	in mino.Receiver) error {

	dela.Logger.Info().Msg("Starting the neff shuffle protocol ...")

	dela.Logger.Info().Msg("SHUFFLE / RECEIVED FROM  : " + from.String())

	election, err := h.getElection(startShuffleMessage)
	if err != nil {
		return xerrors.Errorf("failed to get election: %v", err)
	}

	for round := 1; round <= election.ShuffleThreshold; round++ {
		dela.Logger.Info().Msgf("SHUFFLE / ROUND : %d", round)

		election, err := h.getElection(startShuffleMessage)
		if err != nil {
			return xerrors.Errorf("failed to get election: %v", err)
		}

		if election.Status != electionTypes.Closed {
			return xerrors.Errorf("the election must be closed")
		}

		encryptedBallots := make([][]byte, 0, len(election.EncryptedBallots.Ballots))

		if round == 1 {
			encryptedBallots = append([][]byte{}, election.EncryptedBallots.Ballots...)
			// for _, value := range encryptedBallotsMap {
			// 	encryptedBallots = append(encryptedBallots, value)
			// }
		} else {
			if len(election.ShuffledBallots) != round-1 {
				return xerrors.Errorf("number of shuffled ballots must equal the round number: %d != %d", len(election.ShuffledBallots), round-1)
			}

			if len(election.ShuffledBallots[round-2]) != len(election.EncryptedBallots.Ballots) {
				return xerrors.Errorf("the election must be closed")
			}
			encryptedBallots = election.ShuffledBallots[round-2]
		}

		Ks := make([]kyber.Point, 0, len(election.EncryptedBallots.Ballots))
		Cs := make([]kyber.Point, 0, len(election.EncryptedBallots.Ballots))

		for _, v := range encryptedBallots {
			ciphertext := new(electionTypes.Ciphertext)
			err = json.NewDecoder(bytes.NewBuffer(v)).Decode(ciphertext)
			if err != nil {
				return xerrors.Errorf("failed to unmarshal Ciphertext: %v", err)
			}

			K := suite.Point()
			err = K.UnmarshalBinary(ciphertext.K)
			if err != nil {
				return xerrors.Errorf("failed to unmarshal K: %v", err)
			}

			C := suite.Point()
			err = C.UnmarshalBinary(ciphertext.C)
			if err != nil {
				return xerrors.Errorf("failed to unmarshal C: %v", err)
			}

			Ks = append(Ks, K)
			Cs = append(Cs, C)
		}

		pubKey := suite.Point()
		err = pubKey.UnmarshalBinary(election.Pubkey)
		if err != nil {
			return xerrors.Errorf("couldn't unmarshal public key: %v", err)
		}

		rand := suite.RandomStream()
		Kbar, Cbar, prover := shuffleKyber.Shuffle(suite, nil, pubKey, Ks, Cs, rand)
		shuffleProof, err := proof.HashProve(suite, protocolName, prover)
		if err != nil {
			return xerrors.Errorf("Shuffle proof failed: %v", err.Error())
		}

		shuffledBallots := make([][]byte, 0, len(Kbar))

		for i := 0; i < len(Kbar); i++ {

			kMarshalled, err := Kbar[i].MarshalBinary()
			if err != nil {
				return xerrors.Errorf("failed to marshall kyber.Point: %v", err.Error())
			}

			cMarshalled, err := Cbar[i].MarshalBinary()
			if err != nil {
				return xerrors.Errorf("failed to marshall kyber.Point: %v", err.Error())
			}

			ciphertext := electionTypes.Ciphertext{K: kMarshalled, C: cMarshalled}
			js, err := json.Marshal(ciphertext)
			if err != nil {
				return xerrors.Errorf("failed to marshall Ciphertext: %v", err.Error())
			}

			shuffledBallots = append(shuffledBallots, js)

		}

		manager := getManager(h.signer, h.client)

		err = manager.Sync()
		if err != nil {
			return xerrors.Errorf("failed to sync manager: %v", err.Error())
		}

		shuffleBallotsTransaction := electionTypes.ShuffleBallotsTransaction{
			ElectionID:      startShuffleMessage.GetElectionId(),
			Round:           round,
			ShuffledBallots: shuffledBallots,
			Proof:           shuffleProof,
		}

		js, err := json.Marshal(shuffleBallotsTransaction)
		if err != nil {
			return xerrors.Errorf("failed to marshal ShuffleBallotsTransaction: %v", err.Error())
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

		tx, err := manager.Make(args...)
		if err != nil {
			return xerrors.Errorf("failed to make transaction: %v", err.Error())
		}

		dela.Logger.Info().Msg("TRANSACTION NONCE : " + strconv.Itoa(int(tx.GetNonce())))

		watchCtx, cancel := context.WithTimeout(context.Background(), shuffleTransactionTimeout)

		// err = h.p.Add(tx)

		events := h.service.Watch(watchCtx)

		err = h.p.Add(tx)
		if err != nil {
			cancel()
			return xerrors.Errorf("failed to add transaction to the pool: %v", err.Error())
		}

		notAccepted := false

	loopTxs:
		for event := range events {
			for _, res := range event.Transactions {
				fmt.Println("Tx", res.GetTransaction().GetID(), "expected", tx.GetID())
				if !bytes.Equal(res.GetTransaction().GetID(), tx.GetID()) {
					continue
				}

				dela.Logger.Info().
					Hex("id", tx.GetID()).
					Msg("transaction included in the block")

				accepted, msg := res.GetStatus()

				if !accepted {
					notAccepted = true
					dela.Logger.Info().Msg("Denied : " + msg)
					break loopTxs
				} else {
					dela.Logger.Info().Msg("ACCEPTED")

					if round == election.ShuffleThreshold {
						message := types.EndShuffle{}
						addrs := make([]mino.Address, 0, len(startShuffleMessage.GetAddresses())-1)
						for _, addr := range startShuffleMessage.GetAddresses() {
							if !(addr.Equal(h.me)) {
								addrs = append(addrs, addr)
							}
						}
						errs := out.Send(message, addrs...)
						err = <-errs
						if err != nil {
							cancel()
							return xerrors.Errorf("failed to send EndShuffle message: %v", err)
						}
						dela.Logger.Info().Msg("SENT END SHUFFLE MESSAGES")
					} else {
						dela.Logger.Info().Msg("WAITING FOR END SHUFFLE MESSAGE")
						addr, msg, err := in.Recv(context.Background())
						if err != nil {
							cancel()
							return xerrors.Errorf("got an error from '%s' while "+
								"receiving: %v", addr, err)
						}
						_, ok := msg.(types.EndShuffle)
						if !ok {
							cancel()
							return xerrors.Errorf("expected to receive an EndShuffle message, but "+
								"go the following: %T", msg)
						}
						dela.Logger.Info().Msg("RECEIVED END SHUFFLE MESSAGE")
					}

					if startShuffleMessage.GetAddresses()[0].Equal(h.me) {
						message := types.EndShuffle{}
						errs := out.Send(message, from)
						err = <-errs
						if err != nil {
							cancel()
							return xerrors.Errorf("failed to send EndShuffle message: %v", err)
						}
					}
					cancel()
					return nil
				}
			}
			if notAccepted {
				break
			}
		}
		cancel()
		dela.Logger.Info().Msg("NEXT ROUND")

	}
	dela.Logger.Info().Msg("Shuffle is done without your contribution")
	return nil
}

func (h *Handler) getElection(startShuffleMessage types.StartShuffle) (*electionTypes.Election, error) {
	electionIDBuff, err := hex.DecodeString(startShuffleMessage.GetElectionId())
	if err != nil {
		return nil, xerrors.Errorf("failed to decode election id: %v", err)
	}

	prf, err := h.service.GetProof(electionIDBuff)
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
