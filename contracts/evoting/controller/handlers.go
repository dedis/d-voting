package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"encoding/base64"
	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	uuid "github.com/satori/go.uuid"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
	"go.dedis.ch/kyber/v3/sign/schnorr"
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
	decodedRequest := handleSignReq(w,r)
	req := &types.CreateElectionRequest{}

	err := json.Unmarshal(decodedRequest,req)
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

// OpenElection allows opening an election, which sets the public key based on
// the DKG actor.
// Body: hex-encoded electionID
func (h *votingProxy) OpenElection(w http.ResponseWriter, r *http.Request) {
	decodedRequest := handleSignReq(w,r)

	// hex-encoded string
	electionID := string(decodedRequest)

	// sanity check that it is a hex-encoded string
	_, err := hex.DecodeString(electionID)
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
	decodedRequest := handleSignReq(w,r)

	err := json.Unmarshal(decodedRequest,req)
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
	decodedRequest := handleSignReq(w,r)

	err := json.Unmarshal(decodedRequest,req)
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
	decodedRequest := handleSignReq(w,r)

	err := json.Unmarshal(decodedRequest,req)
	if err != nil {
		http.Error(w, "failed to decode GetElectionInfoRequest: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	election, err := getElection(h.context, h.electionFac, req.ElectionID, h.orderingSvc)
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
	decodedRequest := handleSignReq(w,r)
	req := &types.GetAllElectionsInfoRequest{}

	err := json.Unmarshal(decodedRequest,req)
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
	decodedRequest := handleSignReq(w,r)

	err := json.Unmarshal(decodedRequest,req)
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
	decodedRequest := handleSignReq(w,r)

	err := json.Unmarshal(decodedRequest,req)
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

	election, err := getElection(h.context, h.electionFac, req.ElectionID, h.orderingSvc)
	if err != nil {
		http.Error(w, xerrors.Errorf(getElectionErr, err).Error(),
			http.StatusInternalServerError)
	}

	if election.Status != types.Closed {
		http.Error(w, "The election must be closed !", http.StatusUnauthorized)
		return
	}

	if len(election.Suffragia.Ciphervotes) <= 1 {
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

// BeginDecryption starts the decryption process by gather the pubShares
func (h *votingProxy) BeginDecryption(w http.ResponseWriter, r *http.Request) {
	decodedRequest := handleSignReq(w,r)
	req := &types.BeginDecryptionRequest{}
	
	err :=json.Unmarshal(decodedRequest,req)
	if err != nil {
		http.Error(w, "failed to decode BeginDecryptionRequest: "+err.Error(),
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

	election, err := getElection(h.context, h.electionFac, req.ElectionID, h.orderingSvc)
	if err != nil {
		http.Error(w, xerrors.Errorf(getElectionErr, err).Error(),
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

	actor, exists := h.dkg.GetActor(electionIDBuf)
	if !exists {
		http.Error(w, "failed to get actor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = actor.ComputePubshares()
	if err != nil {
		http.Error(w, "failed to request the public shares: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	response := types.BeginDecryptionResponse{
		Message: "Decryption process started. Gathering public shares...",
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

}

// CombineShares decrypts the shuffled ballots in an election.
func (h *votingProxy) CombineShares(w http.ResponseWriter, r *http.Request) {
	req := &types.CombineSharesRequest{}
	decodedRequest := handleSignReq(w,r)

	err := json.Unmarshal(decodedRequest,req)
	if err != nil {
		http.Error(w, "failed to decode CombineSharesRequest: "+err.Error(),
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

	election, err := getElection(h.context, h.electionFac, req.ElectionID, h.orderingSvc)
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

	if election.AdminID != req.UserID {
		http.Error(w, "only the admin can decrypt the ballots!", http.StatusUnauthorized)
		return
	}

	decryptBallots := types.CombineShares{
		ElectionID: req.ElectionID,
		UserID:     req.UserID,
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

	response := types.CombineSharesResponse{}

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
	decodedRequest := handleSignReq(w,r)
	req := &types.GetElectionResultRequest{}

	err := json.Unmarshal(decodedRequest,req)
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

	election, err := getElection(h.context, h.electionFac, req.ElectionID, h.orderingSvc)
	if err != nil {
		http.Error(w, xerrors.Errorf(getElectionErr, err).Error(),
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
	decodedRequest := handleSignReq(w,r)
	req := new(types.CancelElectionRequest)

	err := json.Unmarshal(decodedRequest,req)
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


func checkSignature(buff string, signature []byte) error{
	hash256 := sha256.New()
	hash256.Write([]byte(buff))
	md := hash256.Sum(nil)
	err := schnorr.Verify(suite, kp.Public, md, signature)
	return err
}

func handleSignReq(w http.ResponseWriter, r *http.Request) []byte{
	signreq := &types.SignRequest{}

	err := json.NewDecoder(r.Body).Decode(signreq)
	if err != nil {
		http.Error(w, "failed to decode the request: "+err.Error(),
			http.StatusBadRequest)
		return []byte("")
	}

	buff := signreq.Payload
	signature := signreq.Signature

	err = checkSignature(buff, signature)
	if err != nil {
		http.Error(w, "wrong signature: "+err.Error(),
			http.StatusBadRequest)
		return []byte("")
	}

	// convert from base64url
	decodedRequest, err := base64.URLEncoding.DecodeString(buff)
	if err != nil {
		http.Error(w, "failed to from base64url"+err.Error(),
			http.StatusBadRequest)
		return []byte("")
	}
	return decodedRequest
}