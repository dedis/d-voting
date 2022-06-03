package proxy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	ptypes "github.com/dedis/d-voting/proxy/types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"golang.org/x/xerrors"
)

func newSignedErr(err error) error {
	return xerrors.Errorf("failed to created signed request: %v", err)
}

func getSignedErr(err error) error {
	return xerrors.Errorf("failed to get and verify signed request: %v", err)
}

// NewElection returns a new initialized election proxy
func NewElection(srv ordering.Service, mngr txn.Manager, p pool.Pool,
	ctx serde.Context, fac serde.Factory, pk kyber.Point) Election {

	logger := dela.Logger.With().Timestamp().Str("role", "evoting-proxy").Logger()

	return &election{
		logger:      logger,
		orderingSvc: srv,
		context:     ctx,
		electionFac: fac,
		mngr:        mngr,
		pool:        p,
		pk:          pk,
	}
}

// election defines HTTP handlers to manipulate the evoting smart contract
//
// - implements proxy.Election
type election struct {
	sync.Mutex

	orderingSvc ordering.Service
	logger      zerolog.Logger
	context     serde.Context
	electionFac serde.Factory
	mngr        txn.Manager
	pool        pool.Pool
	pk          kyber.Point
}

// NewElection implements proxy.Proxy
func (h *election) NewElection(w http.ResponseWriter, r *http.Request) {
	var req ptypes.CreateElectionRequest

	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.GetAndVerify(h.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	createElection := types.CreateElection{
		Configuration: req.Configuration,
		AdminID:       req.AdminID,
	}

	data, err := createElection.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal CreateElectionTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	txID, err := h.submitAndWaitForTxn(r.Context(), evoting.CmdCreateElection, evoting.ElectionArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	hash := sha256.New()
	hash.Write(txID)
	electionID := hash.Sum(nil)

	response := ptypes.CreateElectionResponse{
		ElectionID: hex.EncodeToString(electionID),
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// NewElectionVote implements proxy.Proxy
func (h *election) NewElectionVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars == nil || vars["electionID"] == "" {
		http.Error(w, fmt.Sprintf("electionID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	electionID := vars["electionID"]

	var req ptypes.CastVoteRequest

	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.GetAndVerify(h.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if elecMD.ElectionsIDs.Contains(electionID) < 0 {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	ciphervote := make(types.Ciphervote, len(req.Ballot))

	for i, egpair := range req.Ballot {
		k := suite.Point()

		err = k.UnmarshalBinary(egpair.K)
		if err != nil {
			http.Error(w, "failed to unmarshal K: "+err.Error(), http.StatusInternalServerError)
			return
		}

		c := suite.Point()

		err = c.UnmarshalBinary(egpair.C)
		if err != nil {
			http.Error(w, "failed to unmarshal C: "+err.Error(), http.StatusInternalServerError)
			return
		}

		ciphervote[i] = types.EGPair{
			K: k,
			C: c,
		}
	}

	castVote := types.CastVote{
		ElectionID: electionID,
		UserID:     req.UserID,
		Ballot:     ciphervote,
	}

	data, err := castVote.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal CastVoteTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCastVote, evoting.ElectionArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// EditElection implements proxy.Proxy
func (h *election) EditElection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars == nil || vars["electionID"] == "" {
		http.Error(w, fmt.Sprintf("electionID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	electionID := vars["electionID"]

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if elecMD.ElectionsIDs.Contains(electionID) < 0 {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	var req ptypes.UpdateElectionRequest

	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.GetAndVerify(h.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	switch req.Action {
	case "open":
		h.openElection(electionID, w, r)
	case "close":
		h.closeElection(electionID, w, r)
	case "combineShares":
		h.combineShares(electionID, w, r)
	case "cancel":
		h.cancelElection(electionID, w, r)
	default:
		BadRequestError(w, r, xerrors.Errorf("invalid action: %s", req.Action), nil)
		return
	}
}

// openElection allows opening an election, which sets the public key based on
// the DKG actor.
func (h *election) openElection(elecID string, w http.ResponseWriter, r *http.Request) {
	openElection := types.OpenElection{
		ElectionID: elecID,
	}

	data, err := openElection.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal OpenElectionTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdOpenElection, evoting.ElectionArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// closeElection closes an election.
func (h *election) closeElection(electionIDHex string, w http.ResponseWriter, r *http.Request) {

	closeElection := types.CloseElection{
		ElectionID: electionIDHex,
	}

	data, err := closeElection.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal CloseElectionTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCloseElection, evoting.ElectionArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// combineShares decrypts the shuffled ballots in an election.
func (h *election) combineShares(electionIDHex string, w http.ResponseWriter, r *http.Request) {

	election, err := getElection(h.context, h.electionFac, electionIDHex, h.orderingSvc)
	if err != nil {
		http.Error(w, "failed to get election: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	if election.Status != types.PubSharesSubmitted {
		http.Error(w, "the submission of public shares must be over!",
			http.StatusUnauthorized)
		return
	}

	decryptBallots := types.CombineShares{
		ElectionID: electionIDHex,
	}

	data, err := decryptBallots.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal decryptBallots: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCombineShares, evoting.ElectionArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// cancelElection cancels an election.
func (h *election) cancelElection(electionIDHex string, w http.ResponseWriter, r *http.Request) {

	cancelElection := types.CancelElection{
		ElectionID: electionIDHex,
	}

	data, err := cancelElection.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal CancelElection: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCancelElection, evoting.ElectionArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Election implements proxy.Proxy. The request should not be signed because it
// is fetching public data.
func (h *election) Election(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	vars := mux.Vars(r)

	if vars == nil || vars["electionID"] == "" {
		http.Error(w, fmt.Sprintf("electionID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	electionID := vars["electionID"]

	election, err := getElection(h.context, h.electionFac, electionID, h.orderingSvc)
	if err != nil {
		http.Error(w, xerrors.Errorf("failed to get election: %v", err).Error(), http.StatusInternalServerError)
		return
	}

	var pubkeyBuf []byte

	if election.Pubkey != nil {
		pubkeyBuf, err = election.Pubkey.MarshalBinary()
		if err != nil {
			http.Error(w, "failed to marshal pubkey: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
	}

	roster := make([]string, 0, election.Roster.Len())

	iter := election.Roster.AddressIterator()
	for iter.HasNext() {
		roster = append(roster, iter.GetNext().String())
	}

	response := ptypes.GetElectionResponse{
		ElectionID:      string(election.ElectionID),
		Configuration:   election.Configuration,
		Status:          uint16(election.Status),
		Pubkey:          hex.EncodeToString(pubkeyBuf),
		Result:          election.DecryptedBallots,
		Roster:          roster,
		ChunksPerBallot: election.ChunksPerBallot(),
		BallotSize:      election.BallotSize,
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// Elections implements proxy.Proxy. The request should not be signed because it
// is fecthing public data.
func (h *election) Elections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to get election metadata: %v", err), nil)
		return
	}

	allElectionsInfo := make([]ptypes.LightElection, len(elecMD.ElectionsIDs))

	for i, id := range elecMD.ElectionsIDs {
		election, err := getElection(h.context, h.electionFac, id, h.orderingSvc)
		if err != nil {
			InternalError(w, r, xerrors.Errorf("failed to get election: %v", err), nil)
			return
		}

		var pubkeyBuf []byte

		if election.Pubkey != nil {
			pubkeyBuf, err = election.Pubkey.MarshalBinary()
			if err != nil {
				InternalError(w, r, xerrors.Errorf("failed to marshal pubkey: %v", err), nil)
				return
			}
		}

		info := ptypes.LightElection{
			ElectionID: string(election.ElectionID),
			Title:      election.Configuration.MainTitle,
			Status:     uint16(election.Status),
			Pubkey:     hex.EncodeToString(pubkeyBuf),
		}

		allElectionsInfo[i] = info
	}

	response := ptypes.GetElectionsResponse{Elections: allElectionsInfo}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to write response: %v", err), nil)
		return
	}
}

// DeleteElection implements proxy.Proxy
func (h *election) DeleteElection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars == nil || vars["electionID"] == "" {
		http.Error(w, fmt.Sprintf("electionID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	electionID := vars["electionID"]

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if elecMD.ElectionsIDs.Contains(electionID) < 0 {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	// auth should contain the hex-encoded signature on the hex-encoded election
	// ID
	auth := r.Header.Get("Authorization")

	sig, err := hex.DecodeString(auth)
	if err != nil {
		BadRequestError(w, r, xerrors.Errorf("failed to decode auth: %v", err), nil)
		return
	}

	err = schnorr.Verify(suite, h.pk, []byte(electionID), sig)
	if err != nil {
		ForbiddenError(w, r, xerrors.Errorf("signature verification failed: %v", err), nil)
		return
	}

	deleteElection := types.DeleteElection{
		ElectionID: electionID,
	}

	data, err := deleteElection.Serialize(h.context)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to marshal DeleteElection: %v", err), nil)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdDeleteElection, evoting.ElectionArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// waitForTxnID blocks until `ID` is included or `events` is closed.
func (h *election) waitForTxnID(events <-chan ordering.Event, ID []byte) error {
	for event := range events {
		for _, res := range event.Transactions {
			if !bytes.Equal(res.GetTransaction().GetID(), ID) {
				continue
			}

			ok, msg := res.GetStatus()
			if !ok {
				return xerrors.Errorf("transaction %x denied : %s", ID, msg)
			}

			return nil
		}
	}

	return xerrors.New("transaction not found")
}

func (h *election) getElectionsMetadata() (types.ElectionsMetadata, error) {
	var md types.ElectionsMetadata

	proof, err := h.orderingSvc.GetProof([]byte(evoting.ElectionsMetadataKey))
	if err != nil {
		// if the proof doesn't exist we assume there is no metadata, thus no
		// elections has been created so far.
		return md, nil
	}

	// if there is not election created yet the metadata will be empty
	if len(proof.GetValue()) == 0 {
		return types.ElectionsMetadata{}, nil
	}

	err = json.Unmarshal(proof.GetValue(), &md)
	if err != nil {
		return md, xerrors.Errorf("failed to unmarshal ElectionMetadata: %v", err)
	}

	return md, nil
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
func (h *election) submitAndWaitForTxn(ctx context.Context, cmd evoting.Command,
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

	err = h.waitForTxnID(events, tx.GetID())
	if err != nil {
		return nil, xerrors.Errorf("failed to wait for transaction: %v", err)
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
