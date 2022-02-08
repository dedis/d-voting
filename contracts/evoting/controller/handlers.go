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
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/serde"
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

// OpenElection allows opening an election, which sets the public key based on
// the DKG actor.
// Body: hex-encoded electionID
func (h *votingProxy) OpenElection(w http.ResponseWriter, r *http.Request) {
	// hex-encoded string as byte array
	electionIDBuf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// hex-encoded string
	electionID := hex.EncodeToString(electionIDBuf)

	// sanity check that it is a hex-encoded string
	_, err = hex.DecodeString(electionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+electionID, http.StatusBadRequest)
		return
	}

	openElection := types.OpenElection{
		ElectionID: electionID,
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

	w.Header().Set("Content-Type", "application/json")
}

// CastVote is used to cast a vote in an election.
func (h *votingProxy) CastVote(w http.ResponseWriter, r *http.Request) {
	req := &types.CastVoteRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode CastVoteRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	if req.Token != token {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	fmt.Println("election metadata:", elecMD, req.ElectionID)

	if !elecMD.ElectionsIDs.Contains(req.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	fac := h.context.GetFactory(types.CiphervoteKey{})
	if fac == nil {
		http.Error(w, "empty ciphervote factory", http.StatusInternalServerError)
		return
	}

	msg, err := fac.Deserialize(h.context, req.Ballot)
	if err != nil {
		http.Error(w, "failed to deserialize ballot: "+err.Error(), http.StatusInternalServerError)
	}

	ciphervote, ok := msg.(types.Ciphervote)
	if !ok {
		http.Error(w, fmt.Sprintf("wrong type of ciphervote: %T", msg), http.StatusInternalServerError)
	}

	castVote := types.CastVote{
		ElectionID: req.ElectionID,
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

// ElectionIDs returns a list of all election IDs.
func (h *votingProxy) ElectionIDs(w http.ResponseWriter, r *http.Request) {
	req := &types.GetAllElectionsIDsRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode GetElectionInfoRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	response := types.GetAllElectionsIDsResponse{ElectionsIDs: elecMD.ElectionsIDs}

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
	req := &types.GetElectionInfoRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode GetElectionInfoRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	election, err := getElection(h.context, req.ElectionID, h.orderingSvc)
	if err != nil {
		http.Error(w, "failed to get election: "+err.Error(),
			http.StatusInternalServerError)
	}

	var pubkeyBuf []byte

	if election.Pubkey != nil {
		pubkeyBuf, err = election.Pubkey.MarshalBinary()
		if err != nil {
			http.Error(w, "failed to marshal pubkey: "+err.Error(),
				http.StatusInternalServerError)
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
	req := &types.GetAllElectionsInfoRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode GetAllElectionsInfoRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	allElectionsInfo := make([]types.GetElectionInfoResponse, len(elecMD.ElectionsIDs))

	fac := h.context.GetFactory(types.ElectionKey{})
	if fac == nil {
		http.Error(w, xerrors.New("election factory not found").Error(),
			http.StatusInternalServerError)
		return
	}

	for i, id := range elecMD.ElectionsIDs {
		election, err := getElection(h.context, id, h.orderingSvc)
		if err != nil {
			http.Error(w, "failed to get election: "+err.Error(),
				http.StatusInternalServerError)
		}

		pubkeyBuf, err := election.Pubkey.MarshalBinary()
		if err != nil {
			http.Error(w, "failed to marshal pubkey: "+err.Error(),
				http.StatusInternalServerError)
		}

		info := types.GetElectionInfoResponse{
			ElectionID:    string(election.ElectionID),
			Configuration: election.Configuration,
			Status:        uint16(election.Status),
			Pubkey:        hex.EncodeToString(pubkeyBuf),
			Result:        election.DecryptedBallots,
		}

		allElectionsInfo[i] = info
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
	req := &types.CloseElectionRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode CloseElectionRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(elecMD.ElectionsIDs, req.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	closeElection := types.CloseElection{
		ElectionID: req.ElectionID,
		UserID:     req.UserID,
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
	req := &types.ShuffleBallotsRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode ShuffleBallotsRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(elecMD.ElectionsIDs, req.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	election, err := getElection(h.context, req.ElectionID, h.orderingSvc)
	if err != nil {
		http.Error(w, "failed to get election: "+err.Error(),
			http.StatusInternalServerError)
	}

	if election.Status != types.Closed {
		http.Error(w, "The election must be closed !", http.StatusUnauthorized)
		return
	}

	if !(len(election.Suffragia.Ciphervotes) > 1) {
		http.Error(w, "only one vote has been casted !", http.StatusNotAcceptable)
		return
	}

	if election.AdminID != req.UserID {
		http.Error(w, "only the admin can shuffle the ballots !", http.StatusUnauthorized)
		return
	}

	electionIDBuff, err := hex.DecodeString(req.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	err = h.shuffleActor.Shuffle(electionIDBuff)
	if err != nil {
		http.Error(w, "failed to shuffle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := types.ShuffleBallotsResponse{
		Message: "shuffle started",
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
	req := &types.DecryptBallotsRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode DecryptBallotsRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(elecMD.ElectionsIDs, req.ElectionID) {
		http.Error(w, "The election does not exist", http.StatusNotFound)
		return
	}

	electionIDBuf, err := hex.DecodeString(req.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	election, err := getElection(h.context, req.ElectionID, h.orderingSvc)
	if err != nil {
		http.Error(w, "failed to get election: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	if election.Status != types.ShuffledBallots {
		http.Error(w, "the ballots must have been shuffled !", http.StatusUnauthorized)
		return
	}

	if election.AdminID != req.UserID {
		http.Error(w, "only the admin can decrypt the ballots!", http.StatusUnauthorized)
		return
	}

	if len(election.ShuffleInstances) == 0 {
		http.Error(w, "no shuffled instances", http.StatusInternalServerError)
		return
	}
	wrongBallots := 0

	actor, exists := h.dkg.GetActor(electionIDBuf)
	if !exists {
		http.Error(w, "failed to get actor:"+err.Error(), http.StatusInternalServerError)
		return
	}

	lastShuffled := election.ShuffleInstances[election.ShuffleThreshold-1].ShuffledBallots
	numVotes := len(lastShuffled)
	seqSize := len(lastShuffled[0])

	decryptedBallots := make([]types.Ballot, 0, numVotes)

	for j := 0; j < numVotes; j++ {

		// decryption of one ballot:
		marshalledBallot := strings.Builder{}

		for i := 0; i < seqSize; i++ {

			chunk, err := actor.Decrypt(lastShuffled[j][i].K, lastShuffled[j][i].C)
			if err != nil {
				http.Error(w, "failed to decrypt (K, C): "+err.Error(), http.StatusInternalServerError)
				return
			}
			marshalledBallot.Write(chunk)
		}

		var ballot types.Ballot
		err = ballot.Unmarshal(marshalledBallot.String(), election)
		if err != nil {
			// TODO do we ever send back through http if it's not an error ?
			wrongBallots++
		}

		decryptedBallots = append(decryptedBallots, ballot)
	}

	decryptBallots := types.DecryptBallots{
		ElectionID:       req.ElectionID,
		UserID:           req.UserID,
		DecryptedBallots: decryptedBallots,
	}

	data, err := decryptBallots.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal DecryptBallots: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdDecryptBallots, evoting.ElectionArg, data)
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
	req := &types.GetElectionResultRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode GetElectionResultRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(elecMD.ElectionsIDs, req.ElectionID) {
		http.Error(w, "The election does not exist", http.StatusNotFound)
		return
	}

	election, err := getElection(h.context, req.ElectionID, h.orderingSvc)
	if err != nil {
		http.Error(w, "failed to get election: "+err.Error(),
			http.StatusInternalServerError)
	}

	if election.Status != types.ResultAvailable {
		http.Error(w, "The result is not available.", http.StatusUnauthorized)
		return
	}

	response := types.GetElectionResultResponse{
		Result: election.DecryptedBallots,
	}

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
	req := new(types.CancelElectionRequest)

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "failed to decode CancelElectionRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	elecMD, err := h.getElectionsMetadata()
	if err != nil {
		http.Error(w, "failed to get election metadata", http.StatusNotFound)
		return
	}

	if !contains(elecMD.ElectionsIDs, req.ElectionID) {
		http.Error(w, "the election does not exist", http.StatusNotFound)
		return
	}

	cancelElection := types.CancelElection{
		ElectionID: req.ElectionID,
		UserID:     req.UserID,
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

// getElection gets the election from the snap. Returns the election ID NOT hex
// encoded.
func getElection(ctx serde.Context, electionIDHex string, srv ordering.Service) (types.Election, error) {
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

	fac := ctx.GetFactory(types.ElectionKey{})
	if fac == nil {
		return election, xerrors.New("election factory not found")
	}

	message, err := fac.Deserialize(ctx, electionBuff)
	if err != nil {
		return election, xerrors.Errorf("failed to deserialize Election: %v", err)
	}

	election, ok := message.(types.Election)
	if !ok {
		return election, xerrors.Errorf("wrong message type: %T", message)
	}

	return election, nil
}
