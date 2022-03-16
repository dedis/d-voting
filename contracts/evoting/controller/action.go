package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.dedis.ch/kyber/v3"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/shuffle"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/core/validation"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/bls"
	"go.dedis.ch/dela/crypto/loader"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/mino/proxy"
	sjson "go.dedis.ch/dela/serde/json"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

const token = "token"
const inclusionTimeout = 2 * time.Second
const contentType = "application/json"
const getElectionErr = "failed to get election: %v"

var suite = suites.MustFind("Ed25519")

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

	var ordering ordering.Service

	err = ctx.Injector.Resolve(&ordering)
	if err != nil {
		return xerrors.Errorf("failed to resolve ordering: %v", err)
	}

	var validation validation.Service

	err = ctx.Injector.Resolve(&validation)
	if err != nil {
		return xerrors.Errorf("failed to resolve validation: %v", err)
	}

	client := client{
		srvc: ordering,
		mgr:  validation,
	}

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

	var rosterFac authority.Factory
	err = ctx.Injector.Resolve(&rosterFac)
	if err != nil {
		return xerrors.Errorf("failed to resolve authority factory: %v", err)
	}

	serdecontext := sjson.NewContext()
	electionFac := types.NewElectionFactory(types.CiphervoteFactory{}, rosterFac)
	ciphervoteFac := types.CiphervoteFactory{}

	registerVotingProxy(proxy, signer, client, dkg, shuffleActor,
		orderingSvc, p, m, serdecontext, electionFac, ciphervoteFac)

	return nil
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

// scenarioTestAction is an action to run a test scenario
//
// - implements node.ActionTemplate
type scenarioTestAction struct {
}

// Execute implements node.ActionTemplate. It creates an election and
// simulates the full election process
func (a *scenarioTestAction) Execute(ctx node.Context) error {
	proxyAddr1 := ctx.Flags.String("proxy-addr1")
	proxyAddr2 := ctx.Flags.String("proxy-addr2")
	proxyAddr3 := ctx.Flags.String("proxy-addr3")

	var rosterFac authority.Factory
	err := ctx.Injector.Resolve(&rosterFac)
	if err != nil {
		return xerrors.Errorf("failed to resolve authority factory: %v", err)
	}

	serdecontext := sjson.NewContext()
	electionFac := types.NewElectionFactory(types.CiphervoteFactory{}, rosterFac)

	var service ordering.Service
	err = ctx.Injector.Resolve(&service)
	if err != nil {
		return xerrors.Errorf("failed to resolve service: %v", err)
	}

	// ###################################### CREATE SIMPLE ELECTION ######

	fmt.Fprintln(ctx.Out, "Create election")

	// Define the configuration
	configuration := fake.BasicConfiguration

	createSimpleElectionRequest := types.CreateElectionRequest{
		Configuration: configuration,
		AdminID:       "adminId",
	}

	js, err := json.Marshal(createSimpleElectionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	fmt.Fprintln(ctx.Out, "create election js:", string(js))

	resp, err := http.Post(proxyAddr1+createElectionEndpoint, contentType, bytes.NewBuffer(js))
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

	fmt.Fprintln(ctx.Out, "response body:", string(body))

	resp.Body.Close()

	var electionResponse types.CreateElectionResponse

	err = json.Unmarshal(body, &electionResponse)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal create election response: %v", err)
	}

	electionID := electionResponse.ElectionID

	electionIDBuf, err := hex.DecodeString(electionID)
	if err != nil {
		return xerrors.Errorf("failed to decode electionID '%s': %v", electionID, err)
	}

	election, err := getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	// sanity check, the electionID returned and the one stored in the election
	// type must be the same.
	if election.ElectionID != electionID {
		return xerrors.Errorf("electionID mismatch: %s != %s", election.ElectionID, electionID)
	}

	fmt.Fprintf(ctx.Out, "Title of the election: "+election.Configuration.MainTitle)
	fmt.Fprintf(ctx.Out, "ID of the election: "+election.ElectionID)
	fmt.Fprintf(ctx.Out, "Admin Id of the election: "+election.AdminID)
	fmt.Fprintf(ctx.Out, "Status of the election: "+strconv.Itoa(int(election.Status)))

	// ##################################### SETUP DKG #########################

	fmt.Fprintln(ctx.Out, "Init DKG")

	const initEndpoint = "/evoting/dkg/init"

	fmt.Fprintf(ctx.Out, "Node 1")

	resp, err = http.Post(proxyAddr1+initEndpoint, contentType, bytes.NewBuffer([]byte(electionID)))
	if err != nil {
		return xerrors.Errorf("failed to retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	fmt.Fprintf(ctx.Out, "Node 2")

	resp, err = http.Post(proxyAddr2+initEndpoint, contentType, bytes.NewBuffer([]byte(electionID)))
	if err != nil {
		return xerrors.Errorf("failed to retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	fmt.Fprintf(ctx.Out, "Node 3")

	resp, err = http.Post(proxyAddr3+initEndpoint, contentType, bytes.NewBuffer([]byte(electionID)))
	if err != nil {
		return xerrors.Errorf("failed to retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	resp, err = http.Post(proxyAddr1+"/evoting/dkg/setup", contentType, bytes.NewBuffer([]byte(electionID)))
	if err != nil {
		return xerrors.Errorf("failed to retrieve the decryption from the server: %v", err)
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

	fmt.Fprintf(ctx.Out, "Pubkey: %v\n", pubKey)

	// ##################################### OPEN ELECTION #####################

	fmt.Fprintf(ctx.Out, "Open election")

	resp, err = http.Post(proxyAddr1+"/evoting/open", contentType, bytes.NewBuffer([]byte(electionID)))
	if err != nil {
		return xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	// ##################################### GET ELECTION INFO #################

	fmt.Fprintln(ctx.Out, "Get election info")

	getElectionInfoRequest := types.GetElectionInfoRequest{
		ElectionID: electionID,
		Token:      token,
	}

	js, err = json.Marshal(getElectionInfoRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr1+getElectionInfoEndpoint, contentType, bytes.NewBuffer(js))
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

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msgf("Pubkey of the election : %x", election.Pubkey)
	dela.Logger.Info().
		Hex("DKG public key", pubkeyBuf).
		Msg("DKG public key")

	// ############################# ATTEMPT TO CLOSE ELECTION #################

	fmt.Fprintln(ctx.Out, "Close election")

	closeElectionRequest := types.CloseElectionRequest{
		ElectionID: electionID,
		UserID:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(closeElectionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr1+closeElectionEndpoint, contentType, bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed to retrieve the decryption from the server: %v", err)
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

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Admin Id of the election : " + election.AdminID)
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))

	// ##################################### CAST BALLOTS ######################

	fmt.Fprintln(ctx.Out, "cast ballots")

	// Create the ballots
	b1 := string("select:" + encodeID("bb") + ":0,0,1,0\n" +
		"text:" + encodeID("ee") + ":eWVz\n\n") //encoding of "yes"

	b2 := string("select:" + encodeID("bb") + ":1,1,0,0\n" +
		"text:" + encodeID("ee") + ":amE=\n\n") //encoding of "ja

	b3 := string("select:" + encodeID("bb") + ":0,0,0,1\n" +
		"text:" + encodeID("ee") + ":b3Vp\n\n") //encoding of "oui"

	var dkg dkg.DKG
	err = ctx.Injector.Resolve(&dkg)
	if err != nil {
		return xerrors.Errorf("failed to resolve DKG: %v", err)
	}

	dkgActor, exists := dkg.GetActor(electionIDBuf)
	if !exists {
		return xerrors.Errorf("failed to get actor: %v", err)
	}

	const ballotSerializeErr = "failed to serialize ballot: %v"

	// Ballot 1
	ballot1, err := marshallBallot(b1, dkgActor, election.ChunksPerBallot())
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	data1, err := ballot1.Serialize(serdecontext)
	if err != nil {
		return xerrors.Errorf(ballotSerializeErr, err)
	}

	castVoteRequest := types.CastVoteRequest{
		ElectionID: electionID,
		UserID:     "user1",
		Ballot:     data1,
		Token:      token,
	}

	fmt.Fprintln(ctx.Out, "cast first ballot")

	respBody, err := castVote(castVoteRequest, proxyAddr1)
	if err != nil {
		return xerrors.Errorf("failed to cast vote: %v", err)
	}

	dela.Logger.Info().Msg("Response body: " + respBody)

	// Ballot 2
	ballot2, err := marshallBallot(b2, dkgActor, election.ChunksPerBallot())
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	data2, err := ballot2.Serialize(serdecontext)
	if err != nil {
		return xerrors.Errorf(ballotSerializeErr, err)
	}

	castVoteRequest = types.CastVoteRequest{
		ElectionID: electionID,
		UserID:     "user2",
		Ballot:     data2,
		Token:      token,
	}

	fmt.Fprintln(ctx.Out, "cast second ballot")

	respBody, err = castVote(castVoteRequest, proxyAddr1)
	if err != nil {
		return xerrors.Errorf("failed to cast vote: %v", err)
	}

	dela.Logger.Info().Msg("Response body: " + respBody)

	// Ballot 3
	ballot3, err := marshallBallot(b3, dkgActor, election.ChunksPerBallot())
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot: %v", err)
	}

	data3, err := ballot3.Serialize(serdecontext)
	if err != nil {
		return xerrors.Errorf(ballotSerializeErr, err)
	}

	castVoteRequest = types.CastVoteRequest{
		ElectionID: electionID,
		UserID:     "user3",
		Ballot:     data3,
		Token:      token,
	}

	fmt.Fprintln(ctx.Out, "cast third ballot")

	respBody, err = castVote(castVoteRequest, proxyAddr1)
	if err != nil {
		return xerrors.Errorf("failed to cast vote: %v", err)
	}

	dela.Logger.Info().Msg("Response body: " + respBody)

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	encryptedBallots := election.Suffragia.Ciphervotes
	dela.Logger.Info().Msg("Length encrypted ballots: " + strconv.Itoa(len(encryptedBallots)))
	dela.Logger.Info().Msgf("Ballot of user1: %s", encryptedBallots[0])
	dela.Logger.Info().Msgf("Ballot of user2: %s", encryptedBallots[1])
	dela.Logger.Info().Msgf("Ballot of user3: %s", encryptedBallots[2])

	// ############################# CLOSE ELECTION FOR REAL ###################

	fmt.Fprintln(ctx.Out, "Close election (for real)")

	closeElectionRequest = types.CloseElectionRequest{
		ElectionID: electionID,
		UserID:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(closeElectionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr1+closeElectionEndpoint, contentType, bytes.NewBuffer(js))
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

	dela.Logger.Info().Msg("Response body: " + string(body))
	resp.Body.Close()

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	dela.Logger.Info().Msg("Title of the election: " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("Admin Id of the election: " + election.AdminID)
	dela.Logger.Info().Msg("Status of the election: " + strconv.Itoa(int(election.Status)))

	// ###################################### SHUFFLE BALLOTS ##################

	fmt.Fprintln(ctx.Out, "shuffle ballots")

	shuffleBallotsRequest := types.ShuffleBallotsRequest{
		ElectionID: electionID,
		UserID:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(shuffleBallotsRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr1+shuffleBallotsEndpoint, contentType, bytes.NewBuffer(js))
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

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of shuffled ballots : " + strconv.Itoa(len(election.ShuffleInstances)))
	dela.Logger.Info().Msg("Number of encrypted ballots : " + strconv.Itoa(len(election.Suffragia.Ciphervotes)))

	// ###################################### REQUEST PUBLIC SHARES ############

	fmt.Fprintln(ctx.Out, "request public shares")

	beginDecryptionRequest := types.BeginDecryptionRequest{
		ElectionID: electionID,
		UserID:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(beginDecryptionRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection: %v", err)
	}

	resp, err = http.Post(proxyAddr1+beginDecryptionEndpoint, contentType, bytes.NewBuffer(js))
	if err != nil {
		return xerrors.Errorf("failed to request beginning of decryption on the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read body of the response: %v", err)
	}

	dela.Logger.Info().Msg("Response body : " + string(body))
	resp.Body.Close()

	time.Sleep(10 * time.Second)

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	validSubmissions := len(election.PubsharesUnits.Pubshares)

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of Pubshare units submitted: " + strconv.Itoa(validSubmissions))

	// ###################################### DECRYPT BALLOTS ##################

	fmt.Fprintln(ctx.Out, "decrypt ballots")

	decryptBallotsRequest := types.CombineSharesRequest{
		ElectionID: electionID,
		UserID:     "adminId",
		Token:      token,
	}

	js, err = json.Marshal(decryptBallotsRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr1+combineSharesEndpoint, contentType, bytes.NewBuffer(js))
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

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	// dela.Logger.Info().Msg("----------------------- Election : " +
	// string(proof.GetValue()))
	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))

	// ###################################### GET ELECTION RESULT ##############

	fmt.Fprintln(ctx.Out, "Get election result")

	getElectionResultRequest := types.GetElectionResultRequest{
		ElectionID: electionID,
		Token:      token,
	}

	js, err = json.Marshal(getElectionResultRequest)
	if err != nil {
		return xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err = http.Post(proxyAddr1+getElectionResultEndpoint, contentType, bytes.NewBuffer(js))
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

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msg("Number of decrypted ballots : " + strconv.Itoa(len(election.DecryptedBallots)))

	if len(election.DecryptedBallots) != 3 {
		return xerrors.Errorf("unexpected number of decrypted ballot: %d != 3", len(election.DecryptedBallots))
	}

	// dela.Logger.Info().Msg(election.DecryptedBallots[0].Vote)
	// dela.Logger.Info().Msg(election.DecryptedBallots[1].Vote)
	// dela.Logger.Info().Msg(election.DecryptedBallots[2].Vote)

	return nil
}

func encodeID(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
}

func marshallBallot(voteStr string, actor dkg.Actor, chunks int) (types.Ciphervote, error) {

	var ballot = make(types.Ciphervote, chunks)
	vote := strings.NewReader(voteStr)

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
			return types.Ciphervote{}, xerrors.Errorf("failed to encrypt the plaintext: %v", err)
		}

		ballot[i] = types.EGPair{
			K: K,
			C: C,
		}
	}

	return ballot, nil
}

func castVote(castVoteRequest types.CastVoteRequest, proxyAddr string) (string, error) {
	js, err := json.Marshal(castVoteRequest)
	if err != nil {
		return "", xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err := http.Post(proxyAddr+castVoteEndpoint, contentType, bytes.NewBuffer(js))
	if err != nil {
		return "", xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return "", xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	resp.Body.Close()

	return string(body), nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
