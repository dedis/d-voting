package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

// NewElection implements proxy.Proxy
func (h *proxy) NewElection(w http.ResponseWriter, r *http.Request) {
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

// NewElectionVote implements proxy.Proxy
func (h *proxy) NewElectionVote(w http.ResponseWriter, r *http.Request) {
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

	if !elecMD.ElectionsIDs.Contains(electionID) {
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

	response := types.CastVoteResponse{}
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// EditElection implements proxy.Proxy
func (h *proxy) EditElection(w http.ResponseWriter, r *http.Request) {
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
		h.openElection(electionID, w, r)
	case "close":
		h.closeElection(electionID, w, r)
	case "combineShares":
		h.combineShares(electionID, w, r)
	case "cancel":
		h.cancelElection(electionID, w, r)
	}
}

// openElection allows opening an election, which sets the public key based on
// the DKG actor.
func (h *proxy) openElection(elecID string, w http.ResponseWriter, r *http.Request) {
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
func (h *proxy) closeElection(electionIDHex string, w http.ResponseWriter, r *http.Request) {

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
func (h *proxy) combineShares(electionIDHex string, w http.ResponseWriter, r *http.Request) {

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
func (h *proxy) cancelElection(electionIDHex string, w http.ResponseWriter, r *http.Request) {

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

// Election implements proxy.Proxy
func (h *proxy) Election(w http.ResponseWriter, r *http.Request) {
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

// Elections implements proxy.Proxy
func (h *proxy) Elections(w http.ResponseWriter, r *http.Request) {

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	allElectionsInfo := make([]types.LightElection, len(elecMD.ElectionsIDs))

	for i, id := range elecMD.ElectionsIDs {
		election, err := getElection(h.context, h.electionFac, id, h.orderingSvc)
		if err != nil {
			http.Error(w, xerrors.Errorf("failed to get election: %v", err).Error(),
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
