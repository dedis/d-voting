package proxy

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/rs/zerolog"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

const inclusionTimeout = 2 * time.Second

var suite = suites.MustFind("ed25519")

// Proxy defines the public proxy handlers. Function names follow the convention
// used in URL helpers on rails:
// https://guides.rubyonrails.org/routing.html#path-and-url-helpers
//
// For the API specification look at /docs/api.md.
type Proxy interface {
	// POST /elections
	NewElection(http.ResponseWriter, *http.Request)
	// POST /elections/{electionID}/vote
	NewElectionVote(http.ResponseWriter, *http.Request)
	// PUT /elections/{electionID}
	EditElection(http.ResponseWriter, *http.Request)
	// GET /elections
	Elections(http.ResponseWriter, *http.Request)
	// GET /elections/{electionID}
	Election(http.ResponseWriter, *http.Request)
}

// NewProxy returns a new initialized proxy
func NewProxy(srv ordering.Service, mngr txn.Manager, p pool.Pool,
	ctx serde.Context, fac serde.Factory) Proxy {

	logger := dela.Logger.With().Timestamp().Str("role", "evoting-proxy").Logger()

	return &proxy{
		logger:      logger,
		orderingSvc: srv,
		context:     ctx,
		electionFac: fac,
		mngr:        mngr,
		pool:        p,
	}
}

// proxy defines HTTP handlers to manipulate the evoting smart contract
type proxy struct {
	sync.Mutex

	orderingSvc ordering.Service
	logger      zerolog.Logger
	context     serde.Context
	electionFac serde.Factory
	mngr        txn.Manager
	pool        pool.Pool
}

// NotFoundHandler defines a generic handler for 404
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	err := types.HTTPError{
		Title:   "Not found",
		Code:    http.StatusNotFound,
		Message: "The requested endpoint was not found",
		Args: map[string]interface{}{
			"url":    r.URL.String(),
			"method": r.Method,
		},
	}

	buf, _ := json.MarshalIndent(&err, "", "  ")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, string(buf))
}

// NotAllowedHandler degines a generic handler for 405
func NotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	err := types.HTTPError{
		Title:   "Not allowed",
		Code:    http.StatusMethodNotAllowed,
		Message: "The requested endpoint was not allowed",
		Args: map[string]interface{}{
			"url":    r.URL.String(),
			"method": r.Method,
		},
	}

	buf, _ := json.MarshalIndent(&err, "", "  ")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(w, string(buf))
}

// waitForTxnID blocks until `ID` is included or `events` is closed.
func (h *proxy) waitForTxnID(events <-chan ordering.Event, ID []byte) bool {
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

func (h *proxy) getElectionsMetadata() (*types.ElectionsMetadata, error) {
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

// getElection gets the election from the snap. Returns the election ID NOT hex
// encoded.
func getElection(ctx serde.Context, electionFac serde.Factory, electionIDHex string,
	srv ordering.Service) (types.Election, error) {

	var election types.Election

	electionID, err := hex.DecodeString(electionIDHex)
	if err != nil {
		return election, xerrors.Errorf("failed to decode electionIDHex: %v", err)
	}

	proof, err := srv.GetProof(electionID)
	if err != nil {
		return election, xerrors.Errorf("failed to get proof: %v", err)
	}

	electionBuff := proof.GetValue()
	if len(electionBuff) == 0 {
		return election, xerrors.Errorf("election does not exist")
	}

	message, err := electionFac.Deserialize(ctx, electionBuff)
	if err != nil {
		return election, xerrors.Errorf("failed to deserialize Election: %v", err)
	}

	election, ok := message.(types.Election)
	if !ok {
		return election, xerrors.Errorf("wrong message type: %T", message)
	}

	return election, nil
}

// submitAndWaitForTxn submits a transaction and waits for it to be included.
// Returns the transaction ID.
func (h *proxy) submitAndWaitForTxn(ctx context.Context, cmd evoting.Command,
	cmdArg string, payload []byte) ([]byte, error) {
	h.Lock()
	defer h.Unlock()

	err := h.mngr.Sync()
	if err != nil {
		return nil, xerrors.Errorf("failed to sync manager: %v", err)
	}

	tx, err := createTransaction(h.mngr, cmd, cmdArg, payload)
	if err != nil {
		return nil, xerrors.Errorf("failed to create transaction: %v", err)
	}

	watchCtx, cancel := context.WithTimeout(ctx, inclusionTimeout)
	defer cancel()

	events := h.orderingSvc.Watch(watchCtx)

	err = h.pool.Add(tx)
	if err != nil {
		return nil, xerrors.Errorf("failed to add transaction to the pool: %v", err)
	}

	ok := h.waitForTxnID(events, tx.GetID())
	if !ok {
		return nil, xerrors.Errorf("transaction not processed within timeout")
	}

	return tx.GetID(), nil
}

func createTransaction(manager txn.Manager, commandType evoting.Command,
	commandArg string, buf []byte) (txn.Transaction, error) {

	args := []txn.Arg{
		{
			Key:   native.ContractArg,
			Value: []byte(evoting.ContractName),
		},
		{
			Key:   evoting.CmdArg,
			Value: []byte(commandType),
		},
		{
			Key:   commandArg,
			Value: buf,
		},
	}

	tx, err := manager.Make(args...)
	if err != nil {
		return nil, xerrors.Errorf("failed to create transaction from manager: %v", err)
	}
	return tx, nil
}
