package controller

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/shuffle"
	"github.com/rs/zerolog"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/mino/proxy"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

const (
	loginEndpoint               = "/evoting/login"
	createElectionEndpoint      = "/evoting/create"
	openElectionEndpoint        = "/evoting/open"
	castVoteEndpoint            = "/evoting/cast"
	getAllElectionsIdsEndpoint  = "/evoting/allids"
	getElectionInfoEndpoint     = "/evoting/info"
	getAllElectionsInfoEndpoint = "/evoting/all"
	closeElectionEndpoint       = "/evoting/close"
	shuffleBallotsEndpoint      = "/evoting/shuffle"
	decryptBallotsEndpoint      = "/evoting/decrypt"
	getElectionResultEndpoint   = "/evoting/result"
	cancelElectionEndpoint      = "/evoting/cancel"
)

// HTTP exposes an http proxy for all evoting contract commands.
type votingProxy struct {
	sync.Mutex

	signer      crypto.Signer
	orderingSvc ordering.Service
	mino        mino.Mino

	shuffleActor shuffle.Actor
	dkg          dkg.DKG

	pool   pool.Pool
	client signed.Client

	logger zerolog.Logger

	context serde.Context
}

func registerVotingProxy(proxy proxy.Proxy, signer crypto.Signer,
	client signed.Client, dkg dkg.DKG, shuffleActor shuffle.Actor,
	oSvc ordering.Service, p pool.Pool, m mino.Mino, ctx serde.Context) {

	logger := dela.Logger.With().Timestamp().Str("role", "evoting-proxy").Logger()

	h := &votingProxy{
		logger:       logger,
		signer:       signer,
		client:       client,
		dkg:          dkg,
		shuffleActor: shuffleActor,
		orderingSvc:  oSvc,
		pool:         p,
		mino:         m,
		context:      ctx,
	}

	proxy.RegisterHandler(loginEndpoint, h.Login)
	proxy.RegisterHandler(createElectionEndpoint, h.CreateElection)
	proxy.RegisterHandler(openElectionEndpoint, h.OpenElection)
	proxy.RegisterHandler(castVoteEndpoint, h.CastVote)
	proxy.RegisterHandler(getAllElectionsIdsEndpoint, h.ElectionIDs)
	proxy.RegisterHandler(getElectionInfoEndpoint, h.ElectionInfo)
	proxy.RegisterHandler(getAllElectionsInfoEndpoint, h.AllElectionInfo)
	proxy.RegisterHandler(closeElectionEndpoint, h.CloseElection)
	proxy.RegisterHandler(shuffleBallotsEndpoint, h.ShuffleBallots)
	proxy.RegisterHandler(decryptBallotsEndpoint, h.DecryptBallots)
	proxy.RegisterHandler(getElectionResultEndpoint, h.ElectionResult)
	proxy.RegisterHandler(cancelElectionEndpoint, h.CancelElection)
}

// waitForTxnID blocks until `ID` is included or `events` is closed.
func (h *votingProxy) waitForTxnID(events <-chan ordering.Event, ID []byte) bool {
	for event := range events {
		for _, res := range event.Transactions {
			if !bytes.Equal(res.GetTransaction().GetID(), ID) {
				continue
			}

			ok, msg := res.GetStatus()
			if !ok {
				h.logger.Info().Msgf("transaction %x denied : %s", ID, msg)
			}
			return ok
		}
	}
	return false
}

func (h *votingProxy) getElectionsMetadata() (*types.ElectionsMetadata, error) {
	electionsMetadata := &types.ElectionsMetadata{}

	electionMetadataProof, err := h.orderingSvc.GetProof([]byte(evoting.ElectionsMetadataKey))
	if err != nil {
		return nil, xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	err = json.Unmarshal(electionMetadataProof.GetValue(), electionsMetadata)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal ElectionMetadata: %v", err)
	}

	return electionsMetadata, nil
}
