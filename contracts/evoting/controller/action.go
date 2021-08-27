package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/shuffle"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/bls"
	"go.dedis.ch/dela/crypto/ed25519"
	"go.dedis.ch/dela/crypto/loader"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

const url = "http://localhost:"

const token = "token"
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
type serveAction struct{}

// Execute implements node.ActionTemplate. It implements the handling of
// endpoints and start the HTTP server
func (a *serveAction) Execute(ctx node.Context) error {
	listenAddr := ctx.Flags.String("listenAddr")
	signerFilePath := ctx.Flags.String("signer")

	signer, err := getSigner(signerFilePath)
	if err != nil {
		return xerrors.Errorf("failed to get the signer: %v", err)
	}

	var p pool.Pool
	err = ctx.Injector.Resolve(&p)
	if err != nil {
		return xerrors.Errorf("failed to resolve pool.Pool: %v", err)
	}

	var orderingSvc ordering.Service
	err = ctx.Injector.Resolve(&orderingSvc)
	if err != nil {
		return xerrors.Errorf("failed to resolve ordering.Service: %v", err)
	}

	var blocks *blockstore.InDisk
	err = ctx.Injector.Resolve(&blocks)
	if err != nil {
		return xerrors.Errorf("failed to resolve blockstore.InDisk: %v", err)
	}

	client := &Client{Blocks: blocks}

	var dkgActor dkg.Actor
	err = ctx.Injector.Resolve(&dkgActor)
	if err != nil {
		return xerrors.Errorf("failed to resolve dkg.Actor: %v", err)
	}

	var m mino.Mino
	err = ctx.Injector.Resolve(&m)
	if err != nil {
		return xerrors.Errorf("failed to resolve mino: %v", err)
	}

	var shuffleActor shuffle.Actor
	err = ctx.Injector.Resolve(&shuffleActor)
	if err != nil {
		return xerrors.Errorf("failed to resolve shuffle actor: %v", err)
	}

	srv := NewHTTP(listenAddr, signer, client, dkgActor, shuffleActor,
		orderingSvc, p, m)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)

		<-ch
		srv.Stop()
	}()

	srv.Listen()
	return nil
}

func createTransaction(manager txn.Manager, commandType evoting.Command, commandArg string, buf []byte) (txn.Transaction, error) {
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

func decodeMember(address string, publicKey string, m mino.Mino) (mino.Address, crypto.PublicKey, error) {

	// 1. Deserialize the address.
	addrBuf, err := base64.StdEncoding.DecodeString(address)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to b64 decode address: %v", err)
	}

	addr := m.GetAddressFactory().FromText(addrBuf)

	// 2. Deserialize the public key.
	publicKeyFactory := ed25519.NewPublicKeyFactory()

	pubkeyBuf, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to b64 decode public key: %v", err)
	}

	pubkey, err := publicKeyFactory.FromBytes(pubkeyBuf)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to decode public key: %v", err)
	}

	return addr, pubkey, nil
}

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

	resp, err := http.Post(url+strconv.Itoa(1000)+createElectionEndpoint, "application/json", bytes.NewBuffer(js))
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
