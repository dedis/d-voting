package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.dedis.ch/kyber/v3"
	"io"
	"io/ioutil"
	"net/http"
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
	"go.dedis.ch/dela/mino/proxy"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

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
type registerAction struct{}

// Execute implements node.ActionTemplate. It registers the handlers using the
// default proxy from the the injector.
func (a *registerAction) Execute(ctx node.Context) error {
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

	var dkg dkg.DKG
	err = ctx.Injector.Resolve(&dkg)
	if err != nil {
		return xerrors.Errorf("failed to resolve dkg.DKG: %v", err)
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

	var proxy proxy.Proxy
	err = ctx.Injector.Resolve(&proxy)
	if err != nil {
		return xerrors.Errorf("failed to resolve proxy: %v", err)
	}

	registerVotingProxy(proxy, signer, client, dkg, shuffleActor,
		orderingSvc, p, m)

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
	proxyAddr := ctx.Flags.String("proxy-addr")

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

	// ###################################### CREATE SIMPLE ELECTION ######

	dela.Logger.Info().Msg("----------------------- CREATE SIMPLE ELECTION : ")

	// Define the configuration :
	configuration := types.Configuration{
		MainTitle: "electionTitle",
		Scaffold: []types.Subject{
			{
				ID:       "0xaaa",
				Title:    "subject1",
				Order:    nil,
				Subjects: nil,
				Selects: []types.Select{
					{
						ID:      "0xbbb",
						Title:   "Select your favorite snacks",
						MaxN:    3,
						MinN:    0,
						Choices: []string{"snickers", "mars", "vodka", "babibel"},
					},
				},
				Ranks: []types.Rank{},
				Texts: nil,
			},
			{
				ID:       "0xddd",
				Title:    "subject2",
				Order:    nil,
				Subjects: nil,
				Selects:  nil,
				Ranks:    nil,
				Texts: []types.Text{
					{
						ID:        "0xeee",
						Title:     "dissertation",
						MaxN:      1,
						MinN:      1,
						MaxLength: 3,
						Regex:     "",
						Choices:   []string{"write yes in your language"},
					},
				},
			},
		},
	}

	createSimpleElectionRequest := types.CreateElectionRequest{
		Configuration: configuration,
		AdminID:       "adminId",
	}

	js, err := json.Marshal(createSimpleElectionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	fmt.Println("create election js:", string(js))

	resp, err := http.Post(proxyAddr+createElectionEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	var electionResponse types.CreateElectionResponse

	err = json.Unmarshal(body, &electionResponse)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal create election response: %v", err)
	}

	electionID := electionResponse.ElectionID

	electionIDBuf, err := hex.DecodeString(electionID)
	if err != nil {
		return xerrors.Errorf("failed to decode electionID: %v", err)
	}

	proof, err := service.GetProof(electionIDBuf)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	dela.Logger.Info().Msg("Proof : " + string(proof.GetValue()))

	election := new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal SimpleElection : %v", err)
	}

	// sanity check, the electionID returned and the one stored in the election
	// type must be the same.
	if election.ElectionID != electionID {
		return xerrors.Errorf("electionID mismatch: %s != %s", election.ElectionID, electionID)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Admin Id of the election : " + election.AdminID)
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msgf("Max Ballot size : %d => %d chunks per ballot",
		election.BallotSize, election.ChunksPerBallot())

	// ##################################### SETUP DKG #########################

	resp, err = http.Post(proxyAddr+"/evoting/dkg", "application/json", bytes.NewBuffer([]byte(electionID)))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	pubkeyBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read body: %v", err)
	}

	pubKey := suite.Point()
	err = pubKey.UnmarshalBinary(pubkeyBuf)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal pubkey: %v", err)
	}

	fmt.Printf("Pubkey: %v\n", pubKey)

	// ##################################### OPEN ELECTION #####################

	resp, err = http.Post(proxyAddr+"/evoting/open", "application/json", bytes.NewBuffer([]byte(electionID)))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	// ##################################### GET ELECTION INFO #################

	dela.Logger.Info().Msg("----------------------- GET ELECTION INFO : ")

	getElectionInfoRequest := types.GetElectionInfoRequest{
		ElectionID: electionID,
		Token:      token,
	}

	js, err = json.Marshal(getElectionInfoRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr+getElectionInfoEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuf)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msgf("Pubkey of the election : %x", election.Pubkey)
	dela.Logger.Info().
		Hex("DKG public key", pubkeyBuf).
		Msg("DKG public key")

	// ##################################### GET ELECTION INFO #################

	dela.Logger.Info().Msg("----------------------- CLOSE ELECTION : ")

	closeElectionRequest := types.CloseElectionRequest{
		ElectionID: electionID,
		UserID:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(closeElectionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr+closeElectionEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	// Expecting an error since there must be at least two ballots before
	// closing
	if resp.StatusCode != http.StatusInternalServerError {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuf)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshall SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Admin Id of the election : " + election.AdminID)
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))

	// ##################################### CAST BALLOTS ######################

	dela.Logger.Info().Msg("----------------------- CAST BALLOTS : ")

	//Create the ballots :
	b1 := "select:0xbbb:0,0,1,0\n" +
		"text:0xeee:eWVz\n\n" //encoding of "yes"

	b2 := "select:0xbbb:1,1,0,0\n" +
		"text:0xeee:amE=\n\n" //encoding of "ja

	b3 := "select:0xbbb:0,0,0,1\n" +
		"text:0xeee:b3Vp\n\n" //encoding of "oui"

	ballot1, err := marshallBallot(strings.NewReader(b1), dkgActor, election.ChunksPerBallot())
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	ballot2, err := marshallBallot(strings.NewReader(b2), dkgActor, election.ChunksPerBallot())
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	ballot3, err := marshallBallot(strings.NewReader(b3), dkgActor, election.ChunksPerBallot())
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	castVoteRequest := types.CastVoteRequest{
		ElectionID: electionID,
		UserID:     "user1",
		Ballot:     ballot1,
		Token:      token,
	}

	js, err = json.Marshal(castVoteRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr+castVoteEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}
	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	castVoteRequest = types.CastVoteRequest{
		ElectionID: electionID,
		UserID:     "user2",
		Ballot:     ballot2,
		Token:      token,
	}

	js, err = json.Marshal(castVoteRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr+castVoteEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}
	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	castVoteRequest = types.CastVoteRequest{
		ElectionID: electionID,
		UserID:     "user3",
		Ballot:     ballot3,
		Token:      token,
	}

	js, err = json.Marshal(castVoteRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr+castVoteEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}
	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuf)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to set unmarshal SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Length encrypted ballots : " + strconv.Itoa(len(election.PublicBulletinBoard.Ballots)))
	dela.Logger.Info().Msgf("Ballot of user1 : %s", election.PublicBulletinBoard.Ballots[0])
	dela.Logger.Info().Msgf("Ballot of user2 : %s", election.PublicBulletinBoard.Ballots[1])
	dela.Logger.Info().Msgf("Ballot of user3 : %s", election.PublicBulletinBoard.Ballots[2])
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))

	// ##################################### CAST BALLOTS ######################

	// ###################################### CLOSE ELECTION ###################

	dela.Logger.Info().Msg("----------------------- CLOSE ELECTION : ")

	closeElectionRequest = types.CloseElectionRequest{
		ElectionID: electionID,
		UserID:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(closeElectionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr+closeElectionEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuf)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshall SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Admin Id of the election : " + election.AdminID)
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))

	// ##################################### CLOSE ELECTION ####################

	// ###################################### SHUFFLE BALLOTS ##################

	dela.Logger.Info().Msg("----------------------- SHUFFLE BALLOTS : ")

	shuffleBallotsRequest := types.ShuffleBallotsRequest{
		ElectionID: electionID,
		UserID:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(shuffleBallotsRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr+shuffleBallotsEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	// time.Sleep(20 * time.Second)

	proof, err = service.GetProof(electionIDBuf)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshall SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of shuffled ballots : " + strconv.Itoa(len(election.ShuffleInstances)))
	dela.Logger.Info().Msg("Number of encrypted ballots : " + strconv.Itoa(len(election.PublicBulletinBoard.Ballots)))

	// ###################################### SHUFFLE BALLOTS ##################

	// ###################################### DECRYPT BALLOTS ##################

	dela.Logger.Info().Msg("----------------------- DECRYPT BALLOTS : ")

	decryptBallotsRequest := types.DecryptBallotsRequest{
		ElectionID: electionID,
		UserID:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(decryptBallotsRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr+decryptBallotsEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuf)
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
	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))

	// ###################################### DECRYPT BALLOTS ##################

	// ###################################### GET ELECTION RESULT ##############

	dela.Logger.Info().Msg("----------------------- GET ELECTION RESULT : ")

	getElectionResultRequest := types.GetElectionResultRequest{
		ElectionID: electionID,
		Token:      token,
	}

	js, err = json.Marshal(getElectionResultRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr+getElectionResultEndpoint, "application/json", bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	proof, err = service.GetProof(electionIDBuf)
	if err != nil {
		return xerrors.Errorf("failed to read on the blockchain: %v", err)
	}

	election = new(types.Election)
	err = json.NewDecoder(bytes.NewBuffer(proof.GetValue())).Decode(election)
	if err != nil {
		return xerrors.Errorf("failed to unmarshall SimpleElection : %v", err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))

	if len(election.DecryptedBallots) != 3 {
		return xerrors.Errorf("unexpected number of decrypted ballot: %d != 3", len(election.DecryptedBallots))
	}

	//dela.Logger.Info().Msg(election.DecryptedBallots[0].Vote)
	//dela.Logger.Info().Msg(election.DecryptedBallots[1].Vote)
	//dela.Logger.Info().Msg(election.DecryptedBallots[2].Vote)

	// ###################################### GET ELECTION RESULT ##############

	return nil
}

func marshallBallot(vote io.Reader, actor dkg.Actor, chunks int) (types.EncryptedBallot, error) {

	var ballot = make([]types.Ciphertext, chunks)

	buf := make([]byte, 29)

	for i := 0; i < chunks; i++ {
		var K, C kyber.Point
		var err error

		n, err := vote.Read(buf)
		if err != nil {
			return nil, xerrors.Errorf("failed to read: %v", err)
		}

		K, C, _, err = actor.Encrypt(buf[:n])

		if err != nil {
			return types.EncryptedBallot{}, xerrors.Errorf("failed to encrypt the plaintext: %v", err)
		}

		var chunk types.Ciphertext

		err = chunk.FromPoints(K, C)
		if err != nil {
			return types.EncryptedBallot{}, err
		}

		ballot[i] = chunk
	}

	return ballot, nil
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
