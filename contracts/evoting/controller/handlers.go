package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/gorilla/mux"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

// CreateElection allows creating an election.
func (h *votingProxy) CreateElection(w http.ResponseWriter, r *http.Request) {
	req := &types.CreateElectionRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode CreateElectionRequest: "+err.Error(),
			http.StatusBadRequest)
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

	response := types.CreateElectionResponse{
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

// CastVote is used to cast a vote in an election.
func (h *votingProxy) CastVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars == nil || vars["electionID"] == "" {
		http.Error(w, fmt.Sprintf("electionID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	electionID := vars["electionID"]

	req := &types.CastVoteRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode CastVoteRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	fmt.Println("election metadata:", elecMD, electionID)

	if !elecMD.ElectionsIDs.Contains(electionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	msg, err := h.ciphervoteFac.Deserialize(h.context, req.Ballot)
	if err != nil {
		http.Error(w, "failed to deserialize ballot: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ciphervote, ok := msg.(types.Ciphervote)
	if !ok {
		http.Error(w, fmt.Sprintf("wrong type of ciphervote: %T", msg), http.StatusInternalServerError)
		return
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

	response := types.CastVoteResponse{}
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// UpdateElection defines the handler on the PUT elections/{electionID}
func (h *votingProxy) UpdateElection(w http.ResponseWriter, r *http.Request) {
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

	if !elecMD.ElectionsIDs.Contains(electionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	req := &types.UpdateElectionRequest{}

	err = json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode UpdateElectionRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	switch req.Action {
	case "open":
		h.OpenElection(electionID, w, r)
	case "close":
		h.CloseElection(electionID, w, r)
	case "combineShares":
		h.CombineShares(electionID, w, r)
	case "cancel":
		h.CancelElection(electionID, w, r)
	}
}

// OpenElection allows opening an election, which sets the public key based on
// the DKG actor.
// Body: hex-encoded electionID
func (h *votingProxy) OpenElection(elecID string, w http.ResponseWriter, r *http.Request) {
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

// CloseElection closes an election.
func (h *votingProxy) CloseElection(electionIDHex string, w http.ResponseWriter, r *http.Request) {

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

// CombineShares decrypts the shuffled ballots in an election.
func (h *votingProxy) CombineShares(electionIDHex string, w http.ResponseWriter, r *http.Request) {

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

// CancelElection cancels an election.
func (h *votingProxy) CancelElection(electionIDHex string, w http.ResponseWriter, r *http.Request) {

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

// submitAndWaitForTxn submits a transaction and waits for it to be included.
// Returns the transaction ID.
func (h *votingProxy) submitAndWaitForTxn(ctx context.Context, cmd evoting.Command,
	cmdArg string, payload []byte) ([]byte, error) {
	h.Lock()
	defer h.Unlock()

	manager := getManager(h.signer, h.client)

	err := manager.Sync()
	if err != nil {
		return nil, xerrors.Errorf("failed to sync manager: %v", err)
	}

	tx, err := createTransaction(manager, cmd, cmdArg, payload)
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

// ElectionInfo returns the information for a given election.
func (h *votingProxy) ElectionInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars == nil || vars["electionID"] == "" {
		http.Error(w, fmt.Sprintf("electionID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	electionID := vars["electionID"]

	election, err := getElection(h.context, h.electionFac, electionID, h.orderingSvc)
	if err != nil {
		http.Error(w, xerrors.Errorf(getElectionErr, err).Error(),
			http.StatusInternalServerError)
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

	response := types.GetElectionInfoResponse{
		ElectionID:    string(election.ElectionID),
		Configuration: election.Configuration,
		Status:        uint16(election.Status),
		Pubkey:        hex.EncodeToString(pubkeyBuf),
		Result:        election.DecryptedBallots,
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// AllElectionInfo returns the information for all elections.
func (h *votingProxy) AllElectionInfo(w http.ResponseWriter, r *http.Request) {

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	allElectionsInfo := make([]types.LightElection, len(elecMD.ElectionsIDs))

	for i, id := range elecMD.ElectionsIDs {
		election, err := getElection(h.context, h.electionFac, id, h.orderingSvc)
		if err != nil {
			http.Error(w, xerrors.Errorf(getElectionErr, err).Error(),
				http.StatusInternalServerError)
		}

		pubkeyBuf, err := election.Pubkey.MarshalBinary()
		if err != nil {
			http.Error(w, "failed to marshal pubkey: "+err.Error(),
				http.StatusInternalServerError)
		}

		info := types.LightElection{
			ElectionID: string(election.ElectionID),
			Status:     uint16(election.Status),
			Pubkey:     hex.EncodeToString(pubkeyBuf),
		}

		allElectionsInfo[i] = info
	}

	response := types.GetAllElectionsInfoResponse{Elections: allElectionsInfo}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
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

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
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

func notAllowedHandler(w http.ResponseWriter, r *http.Request) {
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
