package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/shuffle"
	uuid "github.com/satori/go.uuid"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/bls"
	"go.dedis.ch/dela/crypto/ed25519"
	"go.dedis.ch/dela/crypto/loader"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

const url = "http://localhost:"

const loginEndPoint = "/evoting/login"
const createElectionEndPoint = "/evoting/create"
const castVoteEndpoint = "/evoting/cast"
const getAllElectionsIdsEndpoint = "/evoting/allids"
const getElectionInfoEndpoint = "/evoting/info"
const getAllElectionsInfoEndpoint = "/evoting/all"
const closeElectionEndpoint = "/evoting/close"
const shuffleBallotsEndpoint = "/evoting/shuffle"
const decryptBallotsEndpoint = "/evoting/decrypt"
const getElectionResultEndpoint = "/evoting/result"
const cancelElectionEndpoint = "/evoting/cancel"

const token = "token"
const signerFilePath = "private.key"
const inclusionTimeout = 2 * time.Second

var suite = suites.MustFind("Ed25519")

// TODO : Merge evoting and DKG web server ?

// getManager is the function called when we need a transaction manager. It
// allows us to use a different manager for the tests.
var getManager = func(signer crypto.Signer, s signed.Client) txn.Manager {
	return signed.NewManager(signer, s)
}

// initHttpServer is an action to initialize the shuffle protocol
//
// - implements node.ActionTemplate
type initHttpServerAction struct {
	sync.Mutex
	client *Client
}

// Execute implements node.ActionTemplate. It implements the handling of
// endpoints and start the HTTP server
func (a *initHttpServerAction) Execute(ctx node.Context) error {
	portNumber := ctx.Flags.String("portNumber")

	signer, err := getSigner(signerFilePath)
	if err != nil {
		return xerrors.Errorf("failed to get the signer: %v", err)
	}

	var p pool.Pool
	err = ctx.Injector.Resolve(&p)
	if err != nil {
		return xerrors.Errorf("failed to resolve pool.Pool: %v", err)
	}

	var service ordering.Service
	err = ctx.Injector.Resolve(&service)
	if err != nil {
		return xerrors.Errorf("failed to resolve ordering.Service: %v", err)
	}

	var blocks *blockstore.InDisk
	err = ctx.Injector.Resolve(&blocks)
	if err != nil {
		return xerrors.Errorf("failed to resolve blockstore.InDisk: %v", err)
	}
	a.client.Blocks = blocks

	var dkgActor dkg.Actor
	err = ctx.Injector.Resolve(&dkgActor)
	if err != nil {
		return xerrors.Errorf("failed to resolve dkg.Actor: %v", err)
	}

	http.HandleFunc(loginEndPoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(loginEndPoint)

		userID := uuid.NewV4()
		userToken := token

		response := types.LoginResponse{
			UserID: userID.String(),
			Token:  userToken,
		}

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal LoginResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}

	})

	http.HandleFunc(createElectionEndPoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(createElectionEndPoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		createElectionRequest := new(types.CreateElectionRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(createElectionRequest)
		if err != nil {
			http.Error(w, "Failed to decode CreateElectionRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if createElectionRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		manager := getManager(signer, a.client)

		err = manager.Sync()
		if err != nil {
			http.Error(w, "Failed to sync manager: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// random id
		electionIDBuff := make([]byte, 32)
		_, err = rand.Read(electionIDBuff)
		if err != nil {
			http.Error(w, "Failed to generate random bytes: "+err.Error(), http.StatusInternalServerError)
			return
		}

		electionId := hex.EncodeToString(electionIDBuff)

		createElectionTransaction := types.CreateElectionTransaction{
			ElectionID:       electionId,
			Title:            createElectionRequest.Title,
			AdminId:          createElectionRequest.AdminId,
			ShuffleThreshold: createElectionRequest.ShuffleThreshold,
			Members:          createElectionRequest.Members,
			Format:           createElectionRequest.Format,
		}

		js, err := json.Marshal(createElectionTransaction)
		if err != nil {
			http.Error(w, "Failed to marshal CreateElectionTransaction: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tx, err := createTransaction(js, manager, evoting.CmdCreateElection, evoting.CreateElectionArg)
		if err != nil {
			http.Error(w, "Failed to create transaction: "+err.Error(), http.StatusBadRequest)
			return
		}

		watchCtx, cancel := context.WithTimeout(context.Background(), inclusionTimeout)
		defer cancel()

		events := service.Watch(watchCtx)

		err = p.Add(tx)
		if err != nil {
			http.Error(w, "Failed to add transaction to the pool: "+err.Error(), http.StatusInternalServerError)
			return
		}

		accepted, errorMessage := checkTransactionInclusion(events, tx)
		if !accepted {
			http.Error(w, "Transaction not accepted: "+errorMessage, http.StatusInternalServerError)
			return
		}

		response := types.CreateElectionResponse{
			ElectionID: electionId,
		}

		js, err = json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal CreateElectionResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc(getAllElectionsIdsEndpoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(getAllElectionsIdsEndpoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		getAllElectionsIdsRequest := new(types.GetAllElectionsIdsRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(getAllElectionsIdsRequest)
		if err != nil {
			http.Error(w, "Failed to decode GetElectionInfoRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if getAllElectionsIdsRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		electionsMetadata, err := getElectionsMetadata(service)
		if err != nil {
			http.Error(w, "Failed to get election metadata", http.StatusNotFound)
			return
		}

		response := types.GetAllElectionsIdsResponse{ElectionsIds: electionsMetadata.ElectionsIds}

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal GetAllElectionsIdsResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}

	})

	http.HandleFunc(getElectionInfoEndpoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(getElectionInfoEndpoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		getElectionInfoRequest := new(types.GetElectionInfoRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(getElectionInfoRequest)
		if err != nil {
			http.Error(w, "Failed to decode GetElectionInfoRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if getElectionInfoRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		electionsMetadata, err := getElectionsMetadata(service)
		if err != nil {
			http.Error(w, "Failed to get election metadata", http.StatusNotFound)
			return
		}

		if !contains(electionsMetadata.ElectionsIds, getElectionInfoRequest.ElectionID) {
			http.Error(w, "The election does not exist", http.StatusNotFound)
			return
		}

		electionIDBuff, err := hex.DecodeString(getElectionInfoRequest.ElectionID)
		if err != nil {
			http.Error(w, "Failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		proof, err := service.GetProof(electionIDBuff)
		if err != nil {
			http.Error(w, "Failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
			return
		}

		election := new(types.Election)
		err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
		if err != nil {
			http.Error(w, "Failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
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

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal GetElectionInfoResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}

	})

	http.HandleFunc(getAllElectionsInfoEndpoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(getAllElectionsInfoEndpoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		getAllElectionsInfoRequest := new(types.GetAllElectionsInfoRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(getAllElectionsInfoRequest)
		if err != nil {
			http.Error(w, "Failed to decode GetAllElectionsInfoRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if getAllElectionsInfoRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		electionsMetadata, err := getElectionsMetadata(service)
		if err != nil {
			http.Error(w, "Failed to get election metadata", http.StatusNotFound)
			return
		}

		allElectionsInfo := make([]types.GetElectionInfoResponse, 0, len(electionsMetadata.ElectionsIds))

		for _, id := range electionsMetadata.ElectionsIds {

			electionIDBuff, err := hex.DecodeString(id)
			if err != nil {
				http.Error(w, "Failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
				return
			}

			proof, err := service.GetProof(electionIDBuff)
			if err != nil {
				http.Error(w, "Failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
				return
			}

			election := new(types.Election)
			err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
			if err != nil {
				http.Error(w, "Failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
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

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal GetAllElectionsInfoResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}

	})

	http.HandleFunc(castVoteEndpoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(castVoteEndpoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		castVoteRequest := new(types.CastVoteRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(castVoteRequest)
		if err != nil {
			http.Error(w, "Failed to decode CastVoteRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if castVoteRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		electionsMetadata, err := getElectionsMetadata(service)
		if err != nil {
			http.Error(w, "Failed to get election metadata", http.StatusNotFound)
			return
		}

		if !contains(electionsMetadata.ElectionsIds, castVoteRequest.ElectionID) {
			http.Error(w, "The election does not exist", http.StatusNotFound)
			return
		}

		manager := getManager(signer, a.client)

		err = manager.Sync()
		if err != nil {
			http.Error(w, "Failed to sync manager: "+err.Error(), http.StatusInternalServerError)
			return
		}

		castVoteTransaction := types.CastVoteTransaction{
			ElectionID: castVoteRequest.ElectionID,
			UserId:     castVoteRequest.UserId,
			Ballot:     castVoteRequest.Ballot,
		}

		js, err := json.Marshal(castVoteTransaction)
		if err != nil {
			http.Error(w, "Failed to marshal CastVoteTransaction: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tx, err := createTransaction(js, manager, evoting.CmdCastVote, evoting.CastVoteArg)
		if err != nil {
			http.Error(w, "Failed to create transaction: "+err.Error(), http.StatusBadRequest)
			return
		}

		watchCtx, cancel := context.WithTimeout(context.Background(), inclusionTimeout)
		defer cancel()

		events := service.Watch(watchCtx)

		err = p.Add(tx)
		if err != nil {
			http.Error(w, "Failed to add transaction to the pool: "+err.Error(), http.StatusInternalServerError)
			return
		}

		accepted, errorMessage := checkTransactionInclusion(events, tx)
		if !accepted {
			http.Error(w, "Transaction not accepted: "+errorMessage, http.StatusInternalServerError)
			return
		}

		response := types.CastVoteResponse{}

		js, err = json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal CastVoteResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}

	})

	http.HandleFunc(closeElectionEndpoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(closeElectionEndpoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		closeElectionRequest := new(types.CloseElectionRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(closeElectionRequest)
		if err != nil {
			http.Error(w, "Failed to decode CloseElectionRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if closeElectionRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		electionsMetadata, err := getElectionsMetadata(service)
		if err != nil {
			http.Error(w, "Failed to get election metadata", http.StatusNotFound)
			return
		}

		if !contains(electionsMetadata.ElectionsIds, closeElectionRequest.ElectionID) {
			http.Error(w, "The election does not exist", http.StatusNotFound)
			return
		}

		manager := getManager(signer, a.client)

		err = manager.Sync()
		if err != nil {
			http.Error(w, "Failed to sync manager: "+err.Error(), http.StatusInternalServerError)
			return
		}

		closeElectionTransaction := types.CloseElectionTransaction{
			ElectionID: closeElectionRequest.ElectionID,
			UserId:     closeElectionRequest.UserId,
		}

		js, err := json.Marshal(closeElectionTransaction)
		if err != nil {
			http.Error(w, "Failed to marshal CloseElectionTransaction: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tx, err := createTransaction(js, manager, evoting.CmdCloseElection, evoting.CloseElectionArg)
		if err != nil {
			http.Error(w, "Failed to create transaction: "+err.Error(), http.StatusBadRequest)
			return
		}

		watchCtx, cancel := context.WithTimeout(context.Background(), inclusionTimeout)
		defer cancel()

		events := service.Watch(watchCtx)

		err = p.Add(tx)
		if err != nil {
			http.Error(w, "Failed to add transaction to the pool: "+err.Error(), http.StatusInternalServerError)
			return
		}

		accepted, errorMessage := checkTransactionInclusion(events, tx)
		if !accepted {
			http.Error(w, "Transaction not accepted: "+errorMessage, http.StatusInternalServerError)
			return
		}

		response := types.CloseElectionResponse{}

		js, err = json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal CloseElectionResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc(shuffleBallotsEndpoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(shuffleBallotsEndpoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		shuffleBallotsRequest := new(types.ShuffleBallotsRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(shuffleBallotsRequest)
		if err != nil {
			http.Error(w, "Failed to decode ShuffleBallotsRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if shuffleBallotsRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		electionsMetadata, err := getElectionsMetadata(service)
		if err != nil {
			http.Error(w, "Failed to get election metadata", http.StatusNotFound)
			return
		}

		if !contains(electionsMetadata.ElectionsIds, shuffleBallotsRequest.ElectionID) {
			http.Error(w, "The election does not exist", http.StatusNotFound)
			return
		}

		electionIDBuff, err := hex.DecodeString(shuffleBallotsRequest.ElectionID)
		if err != nil {
			http.Error(w, "Failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		proof, err := service.GetProof(electionIDBuff)
		if err != nil {
			http.Error(w, "Failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
			return
		}

		election := new(types.Election)
		err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
		if err != nil {
			http.Error(w, "Failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if election.Status != types.Closed {
			http.Error(w, "The election must be closed !", http.StatusUnauthorized)
			return
		}

		if !(len(election.EncryptedBallots) > 1) {
			http.Error(w, "Only one vote has been casted !", http.StatusNotAcceptable)
			return
		}

		if election.AdminId != shuffleBallotsRequest.UserId {
			http.Error(w, "Only the admin can shuffle the ballots !", http.StatusUnauthorized)
			return
		}

		addrs := make([]mino.Address, len(election.Members))
		pubkeys := make([]crypto.PublicKey, len(election.Members))

		var m mino.Mino
		err = ctx.Injector.Resolve(&m)
		if err != nil {
			http.Error(w, "Failed to resolve mino.Mino: "+err.Error(), http.StatusInternalServerError)
			return
		}

		for i, member := range election.Members {
			addr, pubkey, err := decodeMember(member.Address, member.PublicKey, m)
			if err != nil {
				http.Error(w, "Failed to decode CollectiveAuthorityMember: "+err.Error(), http.StatusInternalServerError)
				return
			}

			addrs[i] = addr
			pubkeys[i] = pubkey
		}

		collectiveAuthority := authority.New(addrs, pubkeys)

		var shuffleActor shuffle.Actor
		err = ctx.Injector.Resolve(&shuffleActor)
		if err != nil {
			http.Error(w, "Failed to resolve shuffle.Actor: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = shuffleActor.Shuffle(collectiveAuthority, string(election.ElectionID))

		if err != nil {
			http.Error(w, "Failed to shuffle: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := types.ShuffleBallotsResponse{
			Message: fmt.Sprintf("shuffle started for nodes %v", addrs),
		}

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal ShuffleBallotsResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}

	})

	http.HandleFunc(decryptBallotsEndpoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(decryptBallotsEndpoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		decryptBallotsRequest := new(types.DecryptBallotsRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(decryptBallotsRequest)
		if err != nil {
			http.Error(w, "Failed to decode DecryptBallotsRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if decryptBallotsRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		electionsMetadata, err := getElectionsMetadata(service)
		if err != nil {
			http.Error(w, "Failed to get election metadata", http.StatusNotFound)
			return
		}

		if !contains(electionsMetadata.ElectionsIds, decryptBallotsRequest.ElectionID) {
			http.Error(w, "The election does not exist", http.StatusNotFound)
			return
		}

		electionIDBuff, err := hex.DecodeString(decryptBallotsRequest.ElectionID)
		if err != nil {
			http.Error(w, "Failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		proof, err := service.GetProof(electionIDBuff)
		if err != nil {
			http.Error(w, "Failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
			return
		}

		election := new(types.Election)
		err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
		if err != nil {
			http.Error(w, "Failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if election.Status != types.ShuffledBallots {
			http.Error(w, "The ballots must have been shuffled !", http.StatusUnauthorized)
			return
		}

		if election.AdminId != decryptBallotsRequest.UserId {
			http.Error(w, "Only the admin can decrypt the ballots !", http.StatusUnauthorized)
			return
		}

		Ks := make([]kyber.Point, 0, len(election.ShuffledBallots))
		Cs := make([]kyber.Point, 0, len(election.ShuffledBallots))

		for _, v := range election.ShuffledBallots[election.ShuffleThreshold] {
			ciphertext := new(types.Ciphertext)
			err = json.NewDecoder(bytes.NewBuffer(v)).Decode(ciphertext)
			if err != nil {
				http.Error(w, "Failed to unmarshal Ciphertext: "+err.Error(), http.StatusInternalServerError)
				return
			}

			K := suite.Point()
			err = K.UnmarshalBinary(ciphertext.K)
			if err != nil {
				http.Error(w, "Failed to unmarshal Kyber.Point: "+err.Error(), http.StatusInternalServerError)
				return
			}

			C := suite.Point()
			err = C.UnmarshalBinary(ciphertext.C)
			if err != nil {
				http.Error(w, "Failed to unmarshal Kyber.Point: "+err.Error(), http.StatusInternalServerError)
				return
			}

			Ks = append(Ks, K)
			Cs = append(Cs, C)
		}

		decryptedBallots := make([]types.Ballot, 0, len(election.ShuffledBallots))

		for i := 0; i < len(Ks); i++ {
			message, err := dkgActor.Decrypt(Ks[i], Cs[i], string(election.ElectionID))
			if err != nil {
				http.Error(w, "Failed to decrypt (K,C): "+err.Error(), http.StatusInternalServerError)
				return
			}

			decryptedBallots = append(decryptedBallots, types.Ballot{Vote: string(message)})
		}

		manager := getManager(signer, a.client)

		err = manager.Sync()
		if err != nil {
			http.Error(w, "Failed to sync manager: "+err.Error(), http.StatusInternalServerError)
			return
		}

		decryptBallotsTransaction := types.DecryptBallotsTransaction{
			ElectionID:       decryptBallotsRequest.ElectionID,
			UserId:           decryptBallotsRequest.UserId,
			DecryptedBallots: decryptedBallots,
		}

		js, err := json.Marshal(decryptBallotsTransaction)
		if err != nil {
			http.Error(w, "Failed to marshal DecryptBallotsTransaction: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tx, err := createTransaction(js, manager, evoting.CmdDecryptBallots, evoting.DecryptBallotsArg)
		if err != nil {
			http.Error(w, "Failed to create transaction: "+err.Error(), http.StatusBadRequest)
			return
		}

		watchCtx, cancel := context.WithTimeout(context.Background(), inclusionTimeout)
		defer cancel()

		events := service.Watch(watchCtx)

		err = p.Add(tx)
		if err != nil {
			http.Error(w, "Failed to add transaction to the pool: "+err.Error(), http.StatusInternalServerError)
			return
		}

		accepted, errorMessage := checkTransactionInclusion(events, tx)
		if !accepted {
			http.Error(w, "Transaction not accepted: "+errorMessage, http.StatusInternalServerError)
			return
		}

		response := types.DecryptBallotsResponse{}

		js, err = json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal DecryptBallotsResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc(getElectionResultEndpoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(getElectionResultEndpoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		getElectionResultRequest := new(types.GetElectionResultRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(getElectionResultRequest)
		if err != nil {
			http.Error(w, "Failed to decode GetElectionResultRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if getElectionResultRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		electionsMetadata, err := getElectionsMetadata(service)
		if err != nil {
			http.Error(w, "Failed to get election metadata", http.StatusNotFound)
			return
		}

		if !contains(electionsMetadata.ElectionsIds, getElectionResultRequest.ElectionID) {
			http.Error(w, "The election does not exist", http.StatusNotFound)
			return
		}

		electionIDBuff, err := hex.DecodeString(getElectionResultRequest.ElectionID)
		if err != nil {
			http.Error(w, "Failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		proof, err := service.GetProof(electionIDBuff)
		if err != nil {
			http.Error(w, "Failed to read on the blockchain: "+err.Error(), http.StatusInternalServerError)
			return
		}

		election := new(types.Election)
		err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
		if err != nil {
			http.Error(w, "Failed to unmarshal Election: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if election.Status != types.ResultAvailable {
			http.Error(w, "The result is not available.", http.StatusUnauthorized)
			return
		}

		response := types.GetElectionResultResponse{Result: election.DecryptedBallots}

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal GetElectionResultResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc(cancelElectionEndpoint, func(w http.ResponseWriter, r *http.Request) {

		a.Lock()
		defer a.Unlock()

		dela.Logger.Info().Msg(cancelElectionEndpoint)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read Body: "+err.Error(), http.StatusBadRequest)
			return
		}

		cancelElectionRequest := new(types.CancelElectionRequest)
		err = json.NewDecoder(bytes.NewBuffer(body)).Decode(cancelElectionRequest)
		if err != nil {
			http.Error(w, "Failed to decode CancelElectionRequest: "+err.Error(), http.StatusBadRequest)
			return
		}

		if cancelElectionRequest.Token != token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		electionsMetadata, err := getElectionsMetadata(service)
		if err != nil {
			http.Error(w, "Failed to get election metadata", http.StatusNotFound)
			return
		}

		if !contains(electionsMetadata.ElectionsIds, cancelElectionRequest.ElectionID) {
			http.Error(w, "The election does not exist", http.StatusNotFound)
			return
		}

		manager := getManager(signer, a.client)

		err = manager.Sync()
		if err != nil {
			http.Error(w, "Failed to sync manager: "+err.Error(), http.StatusInternalServerError)
			return
		}

		cancelElectionTransaction := types.CancelElectionTransaction{
			ElectionID: cancelElectionRequest.ElectionID,
			UserId:     cancelElectionRequest.UserId,
		}

		js, err := json.Marshal(cancelElectionTransaction)
		if err != nil {
			http.Error(w, "Failed to marshal CancelElectionTransaction: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tx, err := createTransaction(js, manager, evoting.CmdCancelElection, evoting.CancelElectionArg)
		if err != nil {
			http.Error(w, "Failed to create transaction: "+err.Error(), http.StatusBadRequest)
			return
		}

		watchCtx, cancel := context.WithTimeout(context.Background(), inclusionTimeout)
		defer cancel()

		events := service.Watch(watchCtx)

		err = p.Add(tx)
		if err != nil {
			http.Error(w, "Failed to add transaction to the pool: "+err.Error(), http.StatusInternalServerError)
			return
		}

		accepted, errorMessage := checkTransactionInclusion(events, tx)
		if !accepted {
			http.Error(w, "Transaction not accepted: "+errorMessage, http.StatusInternalServerError)
			return
		}

		response := types.CancelElectionResponse{}

		js, err = json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal CreateElectionResponse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(js)
		if err != nil {
			http.Error(w, "Failed to write in ResponseWriter: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Fatal(http.ListenAndServe(":"+portNumber, nil))

	return nil
}

func getElectionsMetadata(service ordering.Service) (*types.ElectionsMetadata, error) {
	electionsMetadata := new(types.ElectionsMetadata)

	electionMetadataProof, err := service.GetProof([]byte(evoting.ElectionsMetadataKey))
	if err != nil {
		return nil, xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	err = json.NewDecoder(bytes.NewBuffer(electionMetadataProof.GetValue())).Decode(electionsMetadata)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal ElectionMetadata: %v", err)
	}

	return electionsMetadata, nil
}

func createTransaction(js []byte, manager txn.Manager, commandType evoting.Command, commandArg string) (txn.Transaction, error) {
	args := make([]txn.Arg, 3)
	args[0] = txn.Arg{
		Key:   native.ContractArg,
		Value: []byte(evoting.ContractName),
	}
	args[1] = txn.Arg{
		Key:   evoting.CmdArg,
		Value: []byte(commandType),
	}
	args[2] = txn.Arg{
		Key:   commandArg,
		Value: js,
	}

	tx, err := manager.Make(args...)
	if err != nil {
		return nil, xerrors.Errorf("failed to create transaction from manager: %v", err)
	}
	return tx, nil
}

func checkTransactionInclusion(events <-chan ordering.Event, transaction txn.Transaction) (bool, string) {
	for event := range events {
		for _, res := range event.Transactions {
			if !bytes.Equal(res.GetTransaction().GetID(), transaction.GetID()) {
				continue
			}

			dela.Logger.Debug().
				Hex("id", transaction.GetID()).
				Msg("transaction included in the block")

			accepted, msg := res.GetStatus()
			if !accepted {
				dela.Logger.Info().Msg("transaction denied : " + msg)
			}

			return accepted, msg
		}
	}
	return false, "transaction not found"
}

func decodeMember(address string, publicKey string, m mino.Mino) (mino.Address, crypto.PublicKey, error) {

	// 1. Deserialize the address.
	addrBuf, err := base64.StdEncoding.DecodeString(address)
	if err != nil {
		return nil, nil, xerrors.Errorf("base64 address: %v", err)
	}

	addr := m.GetAddressFactory().FromText(addrBuf)

	// 2. Deserialize the public key.
	publicKeyFactory := ed25519.NewPublicKeyFactory()

	pubkeyBuf, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, nil, xerrors.Errorf("base64 public key: %v", err)
	}

	pubkey, err := publicKeyFactory.FromBytes(pubkeyBuf)
	if err != nil {
		return nil, nil, xerrors.Errorf("Failed to decode public key: %v", err)
	}

	return addr, pubkey, nil
}

// TODO : the user has to create the file in advance, maybe we should create
//  it here ?
// getSigner creates a signer from a file.
func getSigner(filePath string) (crypto.Signer, error) {
	l := loader.NewFileLoader(filePath)

	signerData, err := l.Load()
	if err != nil {
		return nil, xerrors.Errorf("Failed to load signer: %v", err)
	}

	signer, err := bls.NewSignerFromBytes(signerData)
	if err != nil {
		return nil, xerrors.Errorf("Failed to unmarshal signer: %v", err)
	}

	return signer, nil
}

// scenarioTestAction is an action to
//
// - implements node.ActionTemplate
type scenarioTestAction struct {
}

// Execute implements node.ActionTemplate. It creates
func (a *scenarioTestAction) Execute(ctx node.Context) error {

	var service ordering.Service
	err := ctx.Injector.Resolve(&service)
	if err != nil {
		return xerrors.Errorf("failed to resolve service: %v", err)
	}

	var dkgActor dkg.Actor
	err = ctx.Injector.Resolve(&dkgActor)
	if err != nil {
		return xerrors.Errorf("failed to resolve actor: %v", err)
	}

	pubkey, err := dkgActor.GetPublicKey()
	if err != nil {
		return xerrors.Errorf("failed to retrieve the public key: %v", err)
	}

	pubkeyBuf, err := pubkey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to encode pubkey: %v", err)
	}

	// ###################################### CREATE SIMPLE ELECTION ######

	dela.Logger.Info().Msg("----------------------- CREATE SIMPLE ELECTION : ")

	roster, err := a.readMembers(ctx)
	if err != nil {
		return xerrors.Errorf("failed to read roster: %v", err)
	}

	createSimpleElectionRequest := types.CreateElectionRequest{
		Title:            "TitleTest",
		AdminId:          "adminId",
		Token:            "token",
		Members:          roster,
		ShuffleThreshold: 2,
	}

	js, err := json.Marshal(createSimpleElectionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err := http.Post(url+strconv.Itoa(1000)+createElectionEndPoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	createSimpleElectionResponse := new(types.CreateElectionResponse)

	err = json.NewDecoder(bytes.NewBuffer(body)).Decode(createSimpleElectionResponse)
	if err != nil {
		return xerrors.Errorf("failed to set unmarshal CastVoteTransaction : %v", err)
	}

	electionId := createSimpleElectionResponse.ElectionID

	electionIDBuff, err := hex.DecodeString(createSimpleElectionResponse.ElectionID)
	if err != nil {
		return xerrors.Errorf("failed to decode electionID: %v", err)
	}

	proof, err := service.GetProof(electionIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	dela.Logger.Info().Msg("Proof : " + string(proof.GetValue()))
	election := new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Title)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Admin Id of the election : " + election.AdminId)
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))

	// ##################################### CREATE SIMPLE ELECTION ############

	// ##################################### GET ELECTION INFO #################

	dela.Logger.Info().Msg("----------------------- GET ELECTION INFO : ")

	getElectionInfoRequest := types.GetElectionInfoRequest{
		ElectionID: electionId,
		Token:      token,
	}

	js, err = json.Marshal(getElectionInfoRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(url+strconv.Itoa(1000)+getElectionInfoEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Title)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Pubkey of the election : " + string(election.Pubkey))
	dela.Logger.Info().
		Hex("DKG public key", pubkeyBuf).
		Msg("DKG public key")

	// ##################################### GET ELECTION INFO #################

	dela.Logger.Info().Msg("----------------------- CLOSE ELECTION : ")

	closeElectionRequest := types.CloseElectionRequest{
		ElectionID: electionId,
		UserId:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(closeElectionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(url+strconv.Itoa(1000)+closeElectionEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshall SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Title)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Admin Id of the election : " + election.AdminId)
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))

	// ##################################### CAST BALLOTS ######################

	dela.Logger.Info().Msg("----------------------- CAST BALLOTS : ")

	ballot1, err := marshallBallot("ballot1", dkgActor)
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	ballot2, err := marshallBallot("ballot2", dkgActor)
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	ballot3, err := marshallBallot("ballot3", dkgActor)
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	castVoteRequest := types.CastVoteRequest{
		ElectionID: electionId,
		UserId:     "user1",
		Ballot:     ballot1,
		Token:      token,
	}

	js, err = json.Marshal(castVoteRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(url+strconv.Itoa(1000)+castVoteEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}
	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	castVoteRequest = types.CastVoteRequest{
		ElectionID: electionId,
		UserId:     "user2",
		Ballot:     ballot2,
		Token:      token,
	}

	js, err = json.Marshal(castVoteRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(url+strconv.Itoa(1000)+castVoteEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}
	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	castVoteRequest = types.CastVoteRequest{
		ElectionID: electionId,
		UserId:     "user3",
		Ballot:     ballot3,
		Token:      token,
	}

	js, err = json.Marshal(castVoteRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(url+strconv.Itoa(1000)+castVoteEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}
	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to set unmarshal SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Length encrypted ballots : " + strconv.Itoa(len(election.EncryptedBallots)))
	dela.Logger.Info().Msg("Ballot of user1 : " + string(election.EncryptedBallots["user1"]))
	dela.Logger.Info().Msg("Ballot of user2 : " + string(election.EncryptedBallots["user2"]))
	dela.Logger.Info().Msg("Ballot of user3 : " + string(election.EncryptedBallots["user3"]))
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))

	// ##################################### CAST BALLOTS ######################

	// ###################################### CLOSE ELECTION ###################

	dela.Logger.Info().Msg("----------------------- CLOSE ELECTION : ")

	closeElectionRequest = types.CloseElectionRequest{
		ElectionID: electionId,
		UserId:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(closeElectionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(url+strconv.Itoa(1000)+closeElectionEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshall SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Title)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Admin Id of the election : " + election.AdminId)
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))

	// ##################################### CLOSE ELECTION ####################

	// ###################################### SHUFFLE BALLOTS ##################

	dela.Logger.Info().Msg("----------------------- SHUFFLE BALLOTS : ")

	shuffleBallotsRequest := types.ShuffleBallotsRequest{
		ElectionID: electionId,
		UserId:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(shuffleBallotsRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(url+strconv.Itoa(1000)+shuffleBallotsEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	// time.Sleep(20 * time.Second)

	proof, err = service.GetProof(electionIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshall SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Title)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of shuffled ballots : " + strconv.Itoa(len(election.ShuffledBallots)))
	dela.Logger.Info().Msg("Number of encrypted ballots : " + strconv.Itoa(len(election.EncryptedBallots)))

	// ###################################### SHUFFLE BALLOTS ##################

	// ###################################### DECRYPT BALLOTS ##################

	dela.Logger.Info().Msg("----------------------- DECRYPT BALLOTS : ")

	decryptBallotsRequest := types.DecryptBallotsRequest{
		ElectionID: electionId,
		UserId:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(decryptBallotsRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(url+strconv.Itoa(1000)+decryptBallotsEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshall SimpleElection : %v", err)
	}

	// dela.Logger.Info().Msg("----------------------- Election : " +
	// string(proof.GetValue()))
	dela.Logger.Info().Msg("Title of the election : " + election.Title)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))

	// ###################################### DECRYPT BALLOTS ##################

	// ###################################### GET ELECTION RESULT ##############

	dela.Logger.Info().Msg("----------------------- GET ELECTION RESULT : ")

	getElectionResultRequest := types.GetElectionResultRequest{
		ElectionID: electionId,
		Token:      token,
	}

	js, err = json.Marshal(getElectionResultRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(url+strconv.Itoa(1000)+getElectionResultEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuff)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshall SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Title)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))
	dela.Logger.Info().Msg(election.DecryptedBallots[0].Vote)
	dela.Logger.Info().Msg(election.DecryptedBallots[1].Vote)
	dela.Logger.Info().Msg(election.DecryptedBallots[2].Vote)

	// ###################################### GET ELECTION RESULT ##############

	return nil
}

func marshallBallot(vote string, actor dkg.Actor) ([]byte, error) {

	K, C, _, err := actor.Encrypt([]byte(vote))
	if err != nil {
		return nil, xerrors.Errorf("failed to encrypt the plaintext: %v", err)
	}

	Kmarshalled, err := K.MarshalBinary()
	if err != nil {
		return nil, xerrors.Errorf("failed to marshall the K element of the ciphertext pair: %v", err)
	}

	Cmarshalled, err := C.MarshalBinary()
	if err != nil {
		return nil, xerrors.Errorf("failed to marshall the C element of the ciphertext pair: %v", err)
	}

	ballot := types.Ciphertext{K: Kmarshalled, C: Cmarshalled}
	js, err := json.Marshal(ballot)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshall Ciphertext: %v", err)
	}

	return js, nil

}

func (a scenarioTestAction) readMembers(ctx node.Context) ([]types.CollectiveAuthorityMember, error) {
	members := ctx.Flags.StringSlice("member")

	roster := make([]types.CollectiveAuthorityMember, len(members))

	for i, member := range members {
		addr, pubkey, err := decodeMemberFromContext(member)
		if err != nil {
			return nil, xerrors.Errorf("failed to decode: %v", err)
		}

		roster[i] = types.CollectiveAuthorityMember{
			Address:   addr,
			PublicKey: pubkey,
		}
	}

	return roster, nil
}

func decodeMemberFromContext(str string) (string, string, error) {
	parts := strings.Split(str, ":")
	if len(parts) != 2 {
		return "", "", xerrors.New("invalid member base64 string")
	}

	return parts[0], parts[1], nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
