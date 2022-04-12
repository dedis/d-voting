package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/mino/proxy"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

// HTTP exposes an http proxy for all evoting contract commands.
type votingProxy struct {
	sync.Mutex

	orderingSvc ordering.Service
	logger      zerolog.Logger
	context     serde.Context
	electionFac serde.Factory
	mngr        txn.Manager
	pool        pool.Pool
}

func registerVotingProxy(proxy proxy.Proxy, mngr txn.Manager, oSvc ordering.Service,
	p pool.Pool, ctx serde.Context, electionFac serde.Factory) {

	logger := dela.Logger.With().Timestamp().Str("role", "evoting-proxy").Logger()

	h := &votingProxy{
		logger:      logger,
		orderingSvc: oSvc,
		context:     ctx,
		electionFac: electionFac,
		mngr:        mngr,
		pool:        p,
	}

	electionRouter := mux.NewRouter()

	electionRouter.HandleFunc("/evoting/elections", h.CreateElection).Methods("POST")
	electionRouter.HandleFunc("/evoting/elections", h.GetElections).Methods("GET")
	electionRouter.HandleFunc("/evoting/elections/{electionID}", h.GetElection).Methods("GET")
	electionRouter.HandleFunc("/evoting/elections/{electionID}", h.UpdateElection).Methods("PUT")
	electionRouter.HandleFunc("/evoting/elections/{electionID}/vote", h.CastVote).Methods("POST")

	electionRouter.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	electionRouter.MethodNotAllowedHandler = http.HandlerFunc(notAllowedHandler)

	proxy.RegisterHandler("/evoting/elections", electionRouter.ServeHTTP)
	proxy.RegisterHandler("/evoting/elections/", electionRouter.ServeHTTP)
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
