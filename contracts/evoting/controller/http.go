package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/shuffle"
	"github.com/rs/zerolog"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"golang.org/x/xerrors"
)

const (
	loginEndpoint               = "/evoting/login"
	createElectionEndpoint      = "/evoting/create"
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

const srvShutdownTimeout = 10 * time.Second

// HTTP exposes an http proxy for all evoting contract commands.
type HTTP struct {
	sync.Mutex

	mux        *http.ServeMux
	server     *http.Server
	ln         net.Listener
	listenAddr string

	quit chan struct{}

	signer      crypto.Signer
	orderingSvc ordering.Service
	mino        mino.Mino

	shuffleActor shuffle.Actor
	dkgActor     dkg.Actor

	pool   pool.Pool
	client *Client

	logger zerolog.Logger
}

// logging is a utility function that logs the http server events
func logging(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				logger.Info().
					Str("method", r.Method).
					Str("url", r.URL.Path).
					Str("remoteAddr", r.RemoteAddr).
					Str("agent", r.UserAgent()).Msg("")
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// auth is a middleware which verifies the token in the request body for all
// endpoints not in allow list
func auth(allowlist ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !contains(allowlist, r.URL.Path) {
				type tokenReq struct {
					Token string
				}

				req := &tokenReq{}
				err := json.NewDecoder(r.Body).Decode(&req)
				if err != nil {
					http.Error(w, "failed to parse token: "+err.Error(), http.StatusInternalServerError)
					return
				}

				if req.Token != token {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func NewHTTP(listenAddr string, signer crypto.Signer, client *Client,
	dkgActor dkg.Actor, shuffleActor shuffle.Actor, oSvc ordering.Service,
	p pool.Pool, m mino.Mino) *HTTP {

	mux := http.NewServeMux()
	logger := dela.Logger.With().Timestamp().Str("role", "evoting-http").Logger()

	h := &HTTP{
		mux: mux,
		server: &http.Server{
			Addr:    listenAddr,
			Handler: logging(logger)(mux),
		},
		logger:       logger,
		listenAddr:   listenAddr,
		quit:         make(chan struct{}),
		signer:       signer,
		client:       client,
		dkgActor:     dkgActor,
		shuffleActor: shuffleActor,
		orderingSvc:  oSvc,
		pool:         p,
		mino:         m,
	}

	mux.HandleFunc(loginEndpoint, h.Login)
	mux.HandleFunc(createElectionEndpoint, h.CreateElection)
	mux.HandleFunc(castVoteEndpoint, h.CastVote)
	mux.HandleFunc(getAllElectionsIdsEndpoint, h.ElectionIDs)
	mux.HandleFunc(getElectionInfoEndpoint, h.ElectionInfo)
	mux.HandleFunc(getAllElectionsInfoEndpoint, h.AllElectionInfo)
	mux.HandleFunc(closeElectionEndpoint, h.CloseElection)
	mux.HandleFunc(shuffleBallotsEndpoint, h.ShuffleBallots)
	mux.HandleFunc(decryptBallotsEndpoint, h.DecryptBallots)
	mux.HandleFunc(getElectionResultEndpoint, h.ElectionResult)
	mux.HandleFunc(cancelElectionEndpoint, h.CancelElection)

	return h
}

func (h *HTTP) Listen() {
	h.logger.Info().Msg("Client server is starting...")

	done := make(chan struct{})

	go func() {
		<-h.quit
		h.logger.Info().Msg("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), srvShutdownTimeout)
		defer cancel()

		h.server.SetKeepAlivesEnabled(false)
		err := h.server.Shutdown(ctx)
		if err != nil {
			h.logger.Fatal().Msgf("Could not gracefully shutdown the server: %v", err)
		}
		close(done)
	}()

	addr := h.listenAddr
	if addr == "" {
		addr = ":0"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		h.logger.Panic().Msgf("failed to create conn '%s': %v", addr, err)
		return
	}

	h.ln = ln
	h.logger.Info().Msgf("Server is ready to handle requests at %s", ln.Addr())

	err = h.server.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		h.logger.Fatal().Msgf("Could not listen on %s: %v", h.listenAddr, err)
	}

	<-done
	h.logger.Info().Msg("Server stopped")
}

func (h *HTTP) Stop() {
	h.quit <- struct{}{}
}

// waitForTxnID blocks until `ID` is included or `events` is closed.
func (h *HTTP) waitForTxnID(events <-chan ordering.Event, ID []byte) bool {
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

func (h *HTTP) getElectionsMetadata() (*types.ElectionsMetadata, error) {
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
