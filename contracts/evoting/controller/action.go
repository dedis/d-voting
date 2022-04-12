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

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	eproxy "github.com/dedis/d-voting/proxy"
	"github.com/dedis/d-voting/services/dkg"
	stypes "github.com/dedis/d-voting/services/dkg/pedersen/types"
	"github.com/dedis/d-voting/services/shuffle"
	"github.com/gorilla/mux"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli/node"
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
	"go.dedis.ch/dela/serde"
	sjson "go.dedis.ch/dela/serde/json"
	"golang.org/x/xerrors"
)

const contentType = "application/json"
const getElectionErr = "failed to get election: %v"

// getManager is the function called when we need a transaction manager. It
// allows us to use a different manager for the tests.
var getManager = func(signer crypto.Signer, s signed.Client) txn.Manager {
	return signed.NewManager(signer, s)
}

// RegisterAction is an action to register the HTTP handlers
//
// - implements node.ActionTemplate
type RegisterAction struct{}

// Execute implements node.ActionTemplate. It registers the handlers using the
// default proxy from the the injector.
func (a *RegisterAction) Execute(ctx node.Context) error {
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

	electionFac := types.NewElectionFactory(types.CiphervoteFactory{}, rosterFac)
	mngr := getManager(signer, client)

	ep := eproxy.NewProxy(ordering, mngr, p, sjson.NewContext(), electionFac)

	electionRouter := mux.NewRouter()

	electionRouter.HandleFunc("/evoting/elections", ep.NewElection).Methods("POST")
	electionRouter.HandleFunc("/evoting/elections", ep.Elections).Methods("GET")
	electionRouter.HandleFunc("/evoting/elections/{electionID}", ep.Election).Methods("GET")
	electionRouter.HandleFunc("/evoting/elections/{electionID}", ep.EditElection).Methods("PUT")
	electionRouter.HandleFunc("/evoting/elections/{electionID}/vote", ep.NewElectionVote).Methods("POST")

	electionRouter.NotFoundHandler = http.HandlerFunc(eproxy.NotFoundHandler)
	electionRouter.MethodNotAllowedHandler = http.HandlerFunc(eproxy.NotAllowedHandler)

	proxy.RegisterHandler("/evoting/elections", electionRouter.ServeHTTP)
	proxy.RegisterHandler("/evoting/elections/", electionRouter.ServeHTTP)

	dela.Logger.Info().Msg("d-voting proxy handlers registered")

	return nil
}

// getSigner creates a signer from a file.
func getSigner(filePath string) (crypto.Signer, error) {
	l := loader.NewFileLoader(filePath)

	signerData, err := l.Load()
	if err != nil {
		return nil, xerrors.Errorf("Failed to load signer from %q: %v", filePath, err)
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

	fmt.Println("Welcome in the scenario test")

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

	url := proxyAddr1 + "/evoting/elections"
	fmt.Fprintln(ctx.Out, "POST", url)

	resp, err := http.Post(url, contentType, bytes.NewBuffer(js))
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
		return xerrors.Errorf("failed to unmarshal create election response: %v - %s", err, body)
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
	fmt.Fprintf(ctx.Out, "Status of the election: "+strconv.Itoa(int(election.Status)))

	// ##################################### SETUP DKG #########################

	fmt.Fprintln(ctx.Out, "Init DKG")

	fmt.Fprintf(ctx.Out, "Node 1")

	err = initDKG(proxyAddr1, electionID)
	if err != nil {
		return xerrors.Errorf("failed to init dkg 1: %v", err)
	}

	fmt.Fprintf(ctx.Out, "Node 2")

	err = initDKG(proxyAddr2, electionID)
	if err != nil {
		return xerrors.Errorf("failed to init dkg 2: %v", err)
	}

	fmt.Fprintf(ctx.Out, "Node 3")

	err = initDKG(proxyAddr3, electionID)
	if err != nil {
		return xerrors.Errorf("failed to init dkg 3: %v", err)
	}

	fmt.Fprintf(ctx.Out, "Setup DKG on node 1")

	_, err = updateDKG(proxyAddr1, electionID, "setup")
	if err != nil {
		return xerrors.Errorf("failed to setup dkg on node 1: %v", err)
	}

	// ##################################### OPEN ELECTION #####################

	fmt.Fprintf(ctx.Out, "Open election")

	_, err = updateElection(proxyAddr1, electionID, "open")
	if err != nil {
		return xerrors.Errorf("failed to open election: %v", err)
	}

	// ##################################### GET ELECTION INFO #################

	fmt.Fprintln(ctx.Out, "Get election")

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	dela.Logger.Info().Msg("Title of the election : " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the election : " + string(election.ElectionID))
	dela.Logger.Info().Msg("Status of the election : " + strconv.Itoa(int(election.Status)))
	dela.Logger.Info().Msgf("Pubkey of the election : %x", election.Pubkey)

	// ############################# ATTEMPT TO CLOSE ELECTION #################

	fmt.Fprintln(ctx.Out, "Close election")

	status, err := updateElection(proxyAddr1, electionID, "close")
	if status != http.StatusInternalServerError {
		return xerrors.Errorf("unexpected error: %d: %v", status, err)
	}

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

	// Ballot 1
	ballot1, err := marshallBallot(b1, dkgActor, election.ChunksPerBallot())
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	castVoteRequest := types.CastVoteRequest{
		UserID: "user1",
		Ballot: ballot1,
	}

	fmt.Fprintln(ctx.Out, "cast first ballot")

	respBody, err := castVote(electionID, castVoteRequest, proxyAddr1)
	if err != nil {
		return xerrors.Errorf("failed to cast vote: %v", err)
	}

	dela.Logger.Info().Msg("Response body: " + respBody)

	// Ballot 2
	ballot2, err := marshallBallot(b2, dkgActor, election.ChunksPerBallot())
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot : %v", err)
	}

	castVoteRequest = types.CastVoteRequest{
		UserID: "user2",
		Ballot: ballot2,
	}

	fmt.Fprintln(ctx.Out, "cast second ballot")

	respBody, err = castVote(electionID, castVoteRequest, proxyAddr1)
	if err != nil {
		return xerrors.Errorf("failed to cast vote: %v", err)
	}

	dela.Logger.Info().Msg("Response body: " + respBody)

	// Ballot 3
	ballot3, err := marshallBallot(b3, dkgActor, election.ChunksPerBallot())
	if err != nil {
		return xerrors.Errorf("failed to marshall ballot: %v", err)
	}

	castVoteRequest = types.CastVoteRequest{
		UserID: "user3",
		Ballot: ballot3,
	}

	fmt.Fprintln(ctx.Out, "cast third ballot")

	respBody, err = castVote(electionID, castVoteRequest, proxyAddr1)
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

	_, err = updateElection(proxyAddr1, electionID, "close")
	if err != nil {
		return xerrors.Errorf("failed to close election: %v", err)
	}

	election, err = getElection(serdecontext, electionFac, electionID, service)
	if err != nil {
		return xerrors.Errorf(getElectionErr, err)
	}

	dela.Logger.Info().Msg("Title of the election: " + election.Configuration.MainTitle)
	dela.Logger.Info().Msg("Status of the election: " + strconv.Itoa(int(election.Status)))

	// ###################################### SHUFFLE BALLOTS ##################

	fmt.Fprintln(ctx.Out, "shuffle ballots")

	req, err := http.NewRequest(http.MethodPut, proxyAddr1+"/evoting/services/shuffle/"+electionID, nil)
	if err != nil {
		return xerrors.Errorf("failed to create shuffle request: %v", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return xerrors.Errorf("failed to execute the shuffle query: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected shuffle status: %s - %s", resp.Status, buf)
	}

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

	_, err = updateDKG(proxyAddr1, electionID, "computePubshares")
	if err != nil {
		return xerrors.Errorf("failed to compute pubshares: %v", err)
	}

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

	_, err = updateElection(proxyAddr1, electionID, "combineShares")
	if err != nil {
		return xerrors.Errorf("failed to combine shares: %v", err)
	}

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

	// ###################################### GET ALL ELECTION ##############

	resp, err = http.Get(proxyAddr1 + "/evoting/elections")
	if err != nil {
		return xerrors.Errorf("failed to get all elections")
	}

	var allElections types.GetAllElectionsInfoResponse

	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&allElections)
	if err != nil {
		return xerrors.Errorf("failed to decode getAllElections: %v", err)
	}

	dela.Logger.Info().Msgf("All elections: %v", allElections)

	if len(allElections.Elections) != 1 && allElections.Elections[0].ElectionID != electionID {
		return xerrors.Errorf("unexpected allElections: %v", allElections)
	}

	return nil
}

func encodeID(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
}

func marshallBallot(voteStr string, actor dkg.Actor, chunks int) (types.CiphervoteJSON, error) {

	var ballot = make(types.CiphervoteJSON, chunks)
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
			return types.CiphervoteJSON{}, xerrors.Errorf("failed to encrypt the plaintext: %v", err)
		}

		kbuff, err := K.MarshalBinary()
		if err != nil {
			return types.CiphervoteJSON{}, xerrors.Errorf("failed to marshal K: %v", err)
		}

		cbuff, err := C.MarshalBinary()
		if err != nil {
			return types.CiphervoteJSON{}, xerrors.Errorf("failed to marshal C: %v", err)
		}

		ballot[i] = types.EGPairJSON{
			K: kbuff,
			C: cbuff,
		}
	}

	return ballot, nil
}

// electionID is hex-encoded
func castVote(electionID string, castVoteRequest types.CastVoteRequest, proxyAddr string) (string, error) {
	js, err := json.Marshal(castVoteRequest)
	if err != nil {
		return "", xerrors.Errorf("failed to set marshall types.SimpleElection : %v", err)
	}

	resp, err := http.Post(proxyAddr+"/evoting/elections/"+electionID+"/vote", contentType, bytes.NewBuffer(js))
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

func updateElection(proxyAddr, electionIDHex, action string) (int, error) {
	msg := types.UpdateElectionRequest{
		Action: action,
	}

	buf, err := json.Marshal(&msg)
	if err != nil {
		return 0, xerrors.Errorf("failed to marshal update request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, proxyAddr+"/evoting/elections/"+electionIDHex, bytes.NewBuffer(buf))
	if err != nil {
		return 0, xerrors.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return resp.StatusCode, xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	return 0, nil
}

func initDKG(proxyAddr, electionIDHex string) error {
	resp, err := http.Post(proxyAddr+"/evoting/services/dkg/actors", contentType, bytes.NewBuffer([]byte(electionIDHex)))
	if err != nil {
		return xerrors.Errorf("failed to post request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	return nil
}

func updateDKG(proxyAddr, electionIDHex, action string) (int, error) {
	msg := stypes.UpdateDKG{
		Action: action,
	}

	buf, err := json.Marshal(&msg)
	if err != nil {
		return 0, xerrors.Errorf("failed to marshal update request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, proxyAddr+"/evoting/services/dkg/actors/"+electionIDHex, bytes.NewBuffer(buf))
	if err != nil {
		return 0, xerrors.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, xerrors.Errorf("failed to execute the query: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return resp.StatusCode, xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	return 0, nil
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
