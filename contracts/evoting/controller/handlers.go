package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/xerrors"
)

// Login responds with the user token.
func (h *votingProxy) Login(w http.ResponseWriter, r *http.Request) {
	userID := uuid.NewV4()
	userToken := token

	response := types.LoginResponse{
		UserID: userID.String(),
		Token:  userToken,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateElection allows creating an election.
func (h *votingProxy) CreateElection(w http.ResponseWriter, r *http.Request) {
	createElectionRequest := &types.CreateElectionRequest{}
	err := json.NewDecoder(r.Body).Decode(createElectionRequest)
	if err != nil {
		http.Error(w, "failed to decode CreateElectionRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	createElectionTransaction := types.CreateElectionTransaction{
		Configuration: createElectionRequest.Configuration,
		AdminID:       createElectionRequest.AdminID,
	}

	payload, err := json.Marshal(createElectionTransaction)
	if err != nil {
		http.Error(w, "failed to marshal CreateElectionTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	txID, err := h.submitAndWaitForTxn(r.Context(), evoting.CmdCreateElection, evoting.CreateElectionArg, payload)
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

// OpenElection allows opening an election, which sets the public key based on
// the DKG actor.
// Body: hex-encoded electionID
func (h *votingProxy) OpenElection(w http.ResponseWriter, r *http.Request) {
	electionIDHex, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// sanity check that this is a well hex-encoded string
	_, err = hex.DecodeString(string(electionIDHex))
	if err != nil {
		http.Error(w, "failed to decode electionID: "+string(electionIDHex),
			http.StatusBadRequest)
		return
	}

	openElecTransaction := types.OpenElectionTransaction{
		ElectionID: string(electionIDHex),
	}

	payload, err := json.Marshal(openElecTransaction)
	if err != nil {
		http.Error(w, "failed to marshal OpenElectionTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdOpenElection, evoting.OpenElectionArg, payload)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

// CastVote is used to cast a vote in an election.
func (h *votingProxy) CastVote(w http.ResponseWriter, r *http.Request) {
	castVoteRequest := &types.CastVoteRequest{}
	err := json.NewDecoder(r.Body).Decode(castVoteRequest)
	if err != nil {
		http.Error(w, "failed to decode CastVoteRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	if castVoteRequest.Token != token {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	fmt.Println("election metadata:", electionsMetadata, castVoteRequest.ElectionID)

	if !electionsMetadata.ElectionsIDs.Contains(castVoteRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	castVoteTransaction := types.CastVoteTransaction{
		ElectionID: castVoteRequest.ElectionID,
		UserID:     castVoteRequest.UserID,
		Ballot:     castVoteRequest.Ballot,
	}

	payload, err := json.Marshal(castVoteTransaction)
	if err != nil {
		http.Error(w, "failed to marshal CastVoteTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCastVote, evoting.CastVoteArg, payload)
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

// ElectionIDs returns a list of all election IDs.
func (h *votingProxy) ElectionIDs(w http.ResponseWriter, r *http.Request) {
	getAllElectionsIDsRequest := &types.GetAllElectionsIDsRequest{}
	err := json.NewDecoder(r.Body).Decode(getAllElectionsIDsRequest)
	if err != nil {
		http.Error(w, "failed to decode GetElectionInfoRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	response := types.GetAllElectionsIDsResponse{ElectionsIDs: electionsMetadata.ElectionsIDs}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// ElectionInfo returns the information for a given election.
func (h *votingProxy) ElectionInfo(w http.ResponseWriter, r *http.Request) {
	getElectionInfoRequest := &types.GetElectionInfoRequest{}
	err := json.NewDecoder(r.Body).Decode(getElectionInfoRequest)
	if err != nil {
		http.Error(w, "failed to decode GetElectionInfoRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIDs, getElectionInfoRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	electionIDBuff, err := hex.DecodeString(getElectionInfoRequest.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	proof, err := h.orderingSvc.GetProof(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to read on the blockchain: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	election := &types.Election{}
	err = json.Unmarshal(proof.GetValue(), election)
	if err != nil {
		http.Error(w, "failed to unmarshal Election: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	response := types.GetElectionInfoResponse{
		ElectionID:    string(election.ElectionID),
		Configuration: election.Configuration,
		Status:        uint16(election.Status),
		Pubkey:        hex.EncodeToString(election.Pubkey),
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
	getAllElectionsInfoRequest := &types.GetAllElectionsInfoRequest{}
	err := json.NewDecoder(r.Body).Decode(getAllElectionsInfoRequest)
	if err != nil {
		http.Error(w, "failed to decode GetAllElectionsInfoRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	allElectionsInfo := make([]types.GetElectionInfoResponse, 0, len(electionsMetadata.ElectionsIDs))

	for _, id := range electionsMetadata.ElectionsIDs {
		electionIDBuff, err := hex.DecodeString(id)
		if err != nil {
			http.Error(w, "failed to decode electionID: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		proof, err := h.orderingSvc.GetProof(electionIDBuff)
		if err != nil {
			http.Error(w, "failed to read on the blockchain: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		election := &types.Election{}
		err = json.Unmarshal(proof.GetValue(), election)
		if err != nil {
			http.Error(w, "failed to unmarshal Election: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		info := types.GetElectionInfoResponse{
			ElectionID:    string(election.ElectionID),
			Configuration: election.Configuration,
			Status:        uint16(election.Status),
			Pubkey:        hex.EncodeToString(election.Pubkey),
			Result:        election.DecryptedBallots,
		}

		allElectionsInfo = append(allElectionsInfo, info)
	}

	response := types.GetAllElectionsInfoResponse{AllElectionsInfo: allElectionsInfo}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// CloseElection closes an election.
func (h *votingProxy) CloseElection(w http.ResponseWriter, r *http.Request) {
	closeElectionRequest := &types.CloseElectionRequest{}
	err := json.NewDecoder(r.Body).Decode(closeElectionRequest)
	if err != nil {
		http.Error(w, "failed to decode CloseElectionRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIDs, closeElectionRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	// retrieve election to find length of random vector :
	electionIDBuff, err := hex.DecodeString(closeElectionRequest.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	proof, err := h.orderingSvc.GetProof(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to read on the blockchain: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	election := &types.Election{}
	err = json.Unmarshal(proof.GetValue(), election)
	if err != nil {
		http.Error(w, "failed to unmarshal Election: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	closeElectionTransaction := types.CloseElectionTransaction{
		ElectionID: closeElectionRequest.ElectionID,
		UserID:     closeElectionRequest.UserID,
	}

	payload, err := json.Marshal(closeElectionTransaction)
	if err != nil {
		http.Error(w, "failed to marshal CloseElectionTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCloseElection, evoting.CloseElectionArg, payload)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.CloseElectionResponse{}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// ShuffleBallots shuffles the ballots in an election.
func (h *votingProxy) ShuffleBallots(w http.ResponseWriter, r *http.Request) {
	shuffleBallotsRequest := &types.ShuffleBallotsRequest{}
	err := json.NewDecoder(r.Body).Decode(shuffleBallotsRequest)
	if err != nil {
		http.Error(w, "failed to decode ShuffleBallotsRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIDs, shuffleBallotsRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	electionIDBuff, err := hex.DecodeString(shuffleBallotsRequest.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	proof, err := h.orderingSvc.GetProof(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to read on the blockchain: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	election := &types.Election{}
	err = json.Unmarshal(proof.GetValue(), election)
	if err != nil {
		http.Error(w, "failed to unmarshal Election: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	if election.Status != types.Closed {
		http.Error(w, "The election must be closed !", http.StatusUnauthorized)
		return
	}

	if !(len(election.PublicBulletinBoard.Ballots) > 1) {
		http.Error(w, "only one vote has been casted !", http.StatusNotAcceptable)
		return
	}

	if election.AdminID != shuffleBallotsRequest.UserID {
		http.Error(w, "only the admin can shuffle the ballots !", http.StatusUnauthorized)
		return
	}

	err = h.shuffleActor.Shuffle(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to shuffle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.ShuffleBallotsResponse{
		Message: fmt.Sprintf("shuffle started"),
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// DecryptBallots decrypts the shuffled ballots in an election.
func (h *votingProxy) DecryptBallots(w http.ResponseWriter, r *http.Request) {
	decryptBallotsRequest := &types.DecryptBallotsRequest{}
	err := json.NewDecoder(r.Body).Decode(decryptBallotsRequest)
	if err != nil {
		http.Error(w, "failed to decode DecryptBallotsRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIDs, decryptBallotsRequest.ElectionID) {
		http.Error(w, "The election does not exist", http.StatusNotFound)
		return
	}

	electionIDBuff, err := hex.DecodeString(decryptBallotsRequest.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	proof, err := h.orderingSvc.GetProof(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to read on the blockchain: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	election := &types.Election{}
	err = json.Unmarshal(proof.GetValue(), election)
	if err != nil {
		http.Error(w, "failed to unmarshal Election: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	if election.Status != types.ShuffledBallots {
		http.Error(w, "the ballots must have been shuffled !", http.StatusUnauthorized)
		return
	}

	if election.AdminID != decryptBallotsRequest.UserID {
		http.Error(w, "only the admin can decrypt the ballots!", http.StatusUnauthorized)
		return
	}

	X, Y, err := election.ShuffleInstances[election.ShuffleThreshold-1].ShuffledBallots.GetElGPairs()
	if err != nil {
		http.Error(w, "failed to get X, Y:"+err.Error(), http.StatusInternalServerError)
		return
	}

	decryptedBallots := make([]types.Ballot, 0, len(election.ShuffleInstances))
	wrongBallots := 0

	for i := 0; i < len(X); i++ {
		// decryption of one ballot:
		marshalledBallot := strings.Builder{}
		for j := 0; j < len(X[i]); j++ {
			chunk, err := h.dkgActor.Decrypt(X[j][i], Y[j][i], electionIDBuff)
			if err != nil {
				http.Error(w, "failed to decrypt (K,C): "+err.Error(), http.StatusInternalServerError)
				return
			}
			marshalledBallot.Write(chunk)
		}

		var ballot types.Ballot
		err = ballot.Unmarshal(marshalledBallot.String(), *election)
		if err != nil {
			wrongBallots += 1 // TODO do we ever send back through http if it's not an error?
			// ==> Should communicate this number there
		}

		decryptedBallots = append(decryptedBallots, ballot)
	}

	decryptBallotsTransaction := types.DecryptBallotsTransaction{
		ElectionID:       decryptBallotsRequest.ElectionID,
		UserID:           decryptBallotsRequest.UserID,
		DecryptedBallots: decryptedBallots,
	}

	payload, err := json.Marshal(decryptBallotsTransaction)
	if err != nil {
		http.Error(w, "failed to marshal DecryptBallotsTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdDecryptBallots, evoting.DecryptBallotsArg, payload)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.DecryptBallotsResponse{}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// ElectionResult calculates and returns the results of the election.
func (h *votingProxy) ElectionResult(w http.ResponseWriter, r *http.Request) {
	getElectionResultRequest := &types.GetElectionResultRequest{}
	err := json.NewDecoder(r.Body).Decode(getElectionResultRequest)
	if err != nil {
		http.Error(w, "failed to decode GetElectionResultRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIDs, getElectionResultRequest.ElectionID) {
		http.Error(w, "The election does not exist", http.StatusNotFound)
		return
	}

	electionIDBuff, err := hex.DecodeString(getElectionResultRequest.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	proof, err := h.orderingSvc.GetProof(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to read on the blockchain: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	election := &types.Election{}
	err = json.Unmarshal(proof.GetValue(), election)
	if err != nil {
		http.Error(w, "failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if election.Status != types.ResultAvailable {
		http.Error(w, "The result is not available.", http.StatusUnauthorized)
		return
	}

	response := types.GetElectionResultResponse{Result: election.DecryptedBallots}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// CancelElection cancels an election.
func (h *votingProxy) CancelElection(w http.ResponseWriter, r *http.Request) {
	cancelElectionRequest := new(types.CancelElectionRequest)
	err := json.NewDecoder(r.Body).Decode(cancelElectionRequest)
	if err != nil {
		http.Error(w, "failed to decode CancelElectionRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIDs, cancelElectionRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	cancelElectionTransaction := types.CancelElectionTransaction{
		ElectionID: cancelElectionRequest.ElectionID,
		UserID:     cancelElectionRequest.UserID,
	}

	payload, err := json.Marshal(cancelElectionTransaction)
	if err != nil {
		http.Error(w, "failed to marshal CancelElectionTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCancelElection, evoting.CancelElectionArg, payload)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.CancelElectionResponse{}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
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
