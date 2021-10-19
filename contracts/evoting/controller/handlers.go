package controller

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	uuid "github.com/satori/go.uuid"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
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
		http.Error(w, "failed to decode CreateElectionRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	electionID, err := types.RandomID()
	if err != nil {
		http.Error(w, "failed to create id: "+err.Error(), http.StatusInternalServerError)
	}

	createElectionTransaction := types.CreateElectionTransaction{
		ElectionID:       electionID,
		Title:            createElectionRequest.Title,
		AdminId:          createElectionRequest.AdminId,
		ShuffleThreshold: createElectionRequest.ShuffleThreshold,
		Members:          createElectionRequest.Members,
		Format:           createElectionRequest.Format,
	}

	payload, err := json.Marshal(createElectionTransaction)
	if err != nil {
		http.Error(w, "failed to marshal CreateElectionTransaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCreateElection, evoting.CreateElectionArg, payload)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.CreateElectionResponse{
		ElectionID: electionID,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
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
		http.Error(w, "fnvalid token", http.StatusUnauthorized)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIds, castVoteRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	castVoteTransaction := types.CastVoteTransaction{
		ElectionID: castVoteRequest.ElectionID,
		UserId:     castVoteRequest.UserId,
		Ballot:     castVoteRequest.Ballot,
	}

	payload, err := json.Marshal(castVoteTransaction)
	if err != nil {
		http.Error(w, "failed to marshal CastVoteTransaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCastVote, evoting.CastVoteArg, payload)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.CastVoteResponse{}
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ElectionIDs returns a list of all election IDs.
func (h *votingProxy) ElectionIDs(w http.ResponseWriter, r *http.Request) {
	getAllElectionsIdsRequest := &types.GetAllElectionsIdsRequest{}
	err := json.NewDecoder(r.Body).Decode(getAllElectionsIdsRequest)
	if err != nil {
		http.Error(w, "failed to decode GetElectionInfoRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	response := types.GetAllElectionsIdsResponse{ElectionsIds: electionsMetadata.ElectionsIds}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ElectionInfo returns the information for a gien election.
func (h *votingProxy) ElectionInfo(w http.ResponseWriter, r *http.Request) {
	getElectionInfoRequest := &types.GetElectionInfoRequest{}
	err := json.NewDecoder(r.Body).Decode(getElectionInfoRequest)
	if err != nil {
		http.Error(w, "failed to decode GetElectionInfoRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIds, getElectionInfoRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	electionIDBuff, err := hex.DecodeString(getElectionInfoRequest.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	proof, err := h.orderingSvc.GetProof(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
		return
	}

	election := &types.Election{}
	err = json.Unmarshal(proof.GetValue(), election)
	if err != nil {
		http.Error(w, "failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.GetElectionInfoResponse{
		ElectionID: string(election.ElectionID),
		Title:      election.Title,
		Status:     uint16(election.Status),
		Pubkey:     hex.EncodeToString(election.Pubkey),
		Result:     election.DecryptedBallots,
		Format:     election.Format,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// AllElectionInfo returns the information for all elections.
func (h *votingProxy) AllElectionInfo(w http.ResponseWriter, r *http.Request) {
	getAllElectionsInfoRequest := &types.GetAllElectionsInfoRequest{}
	err := json.NewDecoder(r.Body).Decode(getAllElectionsInfoRequest)
	if err != nil {
		http.Error(w, "failed to decode GetAllElectionsInfoRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	allElectionsInfo := make([]types.GetElectionInfoResponse, 0, len(electionsMetadata.ElectionsIds))

	for _, id := range electionsMetadata.ElectionsIds {
		electionIDBuff, err := hex.DecodeString(id)
		if err != nil {
			http.Error(w, "failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		proof, err := h.orderingSvc.GetProof(electionIDBuff)
		if err != nil {
			http.Error(w, "failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
			return
		}

		election := &types.Election{}
		err = json.Unmarshal(proof.GetValue(), election)
		if err != nil {
			http.Error(w, "failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
			return
		}

		info := types.GetElectionInfoResponse{
			ElectionID: string(election.ElectionID),
			Title:      election.Title,
			Status:     uint16(election.Status),
			Pubkey:     hex.EncodeToString(election.Pubkey),
			Result:     election.DecryptedBallots,
			Format:     election.Format,
		}

		allElectionsInfo = append(allElectionsInfo, info)
	}

	response := types.GetAllElectionsInfoResponse{AllElectionsInfo: allElectionsInfo}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// CloseElection closes an election.
func (h *votingProxy) CloseElection(w http.ResponseWriter, r *http.Request) {
	closeElectionRequest := &types.CloseElectionRequest{}
	err := json.NewDecoder(r.Body).Decode(closeElectionRequest)
	if err != nil {
		http.Error(w, "failed to decode CloseElectionRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIds, closeElectionRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	closeElectionTransaction := types.CloseElectionTransaction{
		ElectionID: closeElectionRequest.ElectionID,
		UserId:     closeElectionRequest.UserId,
	}

	payload, err := json.Marshal(closeElectionTransaction)
	if err != nil {
		http.Error(w, "failed to marshal CloseElectionTransaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCloseElection, evoting.CloseElectionArg, payload)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.CloseElectionResponse{}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ShuffleBallots shuffles the ballots in an election.
func (h *votingProxy) ShuffleBallots(w http.ResponseWriter, r *http.Request) {
	shuffleBallotsRequest := &types.ShuffleBallotsRequest{}
	err := json.NewDecoder(r.Body).Decode(shuffleBallotsRequest)
	if err != nil {
		http.Error(w, "failed to decode ShuffleBallotsRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIds, shuffleBallotsRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	electionIDBuff, err := hex.DecodeString(shuffleBallotsRequest.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	proof, err := h.orderingSvc.GetProof(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
		return
	}

	election := &types.Election{}
	err = json.Unmarshal(proof.GetValue(), election)
	if err != nil {
		http.Error(w, "failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if election.Status != types.Closed {
		http.Error(w, "The election must be closed !", http.StatusUnauthorized)
		return
	}

	if !(len(election.EncryptedBallots.Ballots) > 1) {
		http.Error(w, "only one vote has been casted !", http.StatusNotAcceptable)
		return
	}

	if election.AdminId != shuffleBallotsRequest.UserId {
		http.Error(w, "only the admin can shuffle the ballots !", http.StatusUnauthorized)
		return
	}

	addrs := make([]mino.Address, len(election.Members))
	pubkeys := make([]crypto.PublicKey, len(election.Members))

	for i, member := range election.Members {
		addr, pubkey, err := decodeMember(member.Address, member.PublicKey, h.mino)
		if err != nil {
			http.Error(w, "failed to decode CollectiveAuthorityMember: "+err.Error(), http.StatusInternalServerError)
			return
		}

		addrs[i] = addr
		pubkeys[i] = pubkey
	}

	collectiveAuthority := authority.New(addrs, pubkeys)

	err = h.shuffleActor.Shuffle(collectiveAuthority, string(election.ElectionID))

	if err != nil {
		http.Error(w, "failed to shuffle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.ShuffleBallotsResponse{
		Message: fmt.Sprintf("shuffle started for nodes %v", addrs),
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// DecryptBallots decrypts the shuffled ballots in an election.
func (h *votingProxy) DecryptBallots(w http.ResponseWriter, r *http.Request) {
	decryptBallotsRequest := &types.DecryptBallotsRequest{}
	err := json.NewDecoder(r.Body).Decode(decryptBallotsRequest)
	if err != nil {
		http.Error(w, "failed to decode DecryptBallotsRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIds, decryptBallotsRequest.ElectionID) {
		http.Error(w, "The election does not exist", http.StatusNotFound)
		return
	}

	electionIDBuff, err := hex.DecodeString(decryptBallotsRequest.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	proof, err := h.orderingSvc.GetProof(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
		return
	}

	election := &types.Election{}
	err = json.Unmarshal(proof.GetValue(), election)
	if err != nil {
		http.Error(w, "failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if election.Status != types.ShuffledBallots {
		http.Error(w, "the ballots must have been shuffled !", http.StatusUnauthorized)
		return
	}

	if election.AdminId != decryptBallotsRequest.UserId {
		http.Error(w, "only the admin can decrypt the ballots!", http.StatusUnauthorized)
		return
	}

	ks, cs, err := election.ShuffledBallots[election.ShuffleThreshold-1].GetKsCs()
	if err != nil {
		http.Error(w, "failed to get ks, cs:"+err.Error(), http.StatusInternalServerError)
		return
	}

	decryptedBallots := make([]types.Ballot, 0, len(election.ShuffledBallots))

	for i := 0; i < len(ks); i++ {
		message, err := h.dkgActor.Decrypt(ks[i], cs[i], string(election.ElectionID))
		if err != nil {
			http.Error(w, "failed to decrypt (K,C): "+err.Error(), http.StatusInternalServerError)
			return
		}

		decryptedBallots = append(decryptedBallots, types.Ballot{Vote: string(message)})
	}

	decryptBallotsTransaction := types.DecryptBallotsTransaction{
		ElectionID:       decryptBallotsRequest.ElectionID,
		UserId:           decryptBallotsRequest.UserId,
		DecryptedBallots: decryptedBallots,
	}

	payload, err := json.Marshal(decryptBallotsTransaction)
	if err != nil {
		http.Error(w, "failed to marshal DecryptBallotsTransaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.submitAndWaitForTxn(r.Context(), evoting.CmdDecryptBallots, evoting.DecryptBallotsArg, payload)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.DecryptBallotsResponse{}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ElectionResult calculates and returns the results of the election.
func (h *votingProxy) ElectionResult(w http.ResponseWriter, r *http.Request) {
	getElectionResultRequest := &types.GetElectionResultRequest{}
	err := json.NewDecoder(r.Body).Decode(getElectionResultRequest)
	if err != nil {
		http.Error(w, "failed to decode GetElectionResultRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIds, getElectionResultRequest.ElectionID) {
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
		http.Error(w, "failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// CancelElection cancels an election.
func (h *votingProxy) CancelElection(w http.ResponseWriter, r *http.Request) {
	cancelElectionRequest := new(types.CancelElectionRequest)
	err := json.NewDecoder(r.Body).Decode(cancelElectionRequest)
	if err != nil {
		http.Error(w, "failed to decode CancelElectionRequest: "+err.Error(), http.StatusBadRequest)
		return
	}

	electionsMetadata, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(electionsMetadata.ElectionsIds, cancelElectionRequest.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	cancelElectionTransaction := types.CancelElectionTransaction{
		ElectionID: cancelElectionRequest.ElectionID,
		UserId:     cancelElectionRequest.UserId,
	}

	payload, err := json.Marshal(cancelElectionTransaction)
	if err != nil {
		http.Error(w, "failed to marshal CancelElectionTransaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCancelElection, evoting.CancelElectionArg, payload)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.CancelElectionResponse{}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *votingProxy) submitAndWaitForTxn(ctx context.Context, cmd evoting.Command,
	cmdArg string, payload []byte) error {
	h.Lock()
	defer h.Unlock()

	manager := getManager(h.signer, h.client)

	err := manager.Sync()
	if err != nil {
		return xerrors.Errorf("failed to sync manager: %v", err)
	}

	tx, err := createTransaction(manager, cmd, cmdArg, payload)
	if err != nil {
		return xerrors.Errorf("failed to create transaction: %v", err)
	}

	watchCtx, cancel := context.WithTimeout(ctx, inclusionTimeout)
	defer cancel()

	events := h.orderingSvc.Watch(watchCtx)

	err = h.pool.Add(tx)
	if err != nil {
		return xerrors.Errorf("failed to add transaction to the pool: %v", err)
	}

	ok := h.waitForTxnID(events, tx.GetID())
	if !ok {
		return xerrors.Errorf("transaction not processed within timeout")
	}

	return nil
}
