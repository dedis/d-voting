package controller

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	eproxy "github.com/dedis/d-voting/proxy"
	ptypes "github.com/dedis/d-voting/proxy/types"
	"github.com/dedis/d-voting/services/dkg"
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

const (
	contentType            = "application/json"
	formPath               = "/evoting/forms"
	formPathSlash          = formPath + "/"
	formIDPath             = formPathSlash + "{formID}"
	unexpectedStatus       = "unexpected status: %s, body: %s"
	failRetrieveDecryption = "failed to retrieve decryption key: %v"
	selectString           = "select:"
	getFormErr             = "failed to get form: %v"
	castFailed             = "failed to cast vote: %v"
	responseBody           = "response body: "
)

var suite = suites.MustFind("ed25519")

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

	formFac := types.NewFormFactory(types.CiphervoteFactory{}, rosterFac)
	mngr := getManager(signer, client)

	proxykeyHex := ctx.Flags.String("proxykey")

	proxykeyBuf, err := hex.DecodeString(proxykeyHex)
	if err != nil {
		return xerrors.Errorf("failed to decode proxykeyHex: %v", err)
	}

	proxykey := suite.Point()

	err = proxykey.UnmarshalBinary(proxykeyBuf)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal proxy key: %v", err)
	}

	ep := eproxy.NewForm(ordering, mngr, p, sjson.NewContext(), formFac, proxykey)

	router := mux.NewRouter()

	router.HandleFunc(formPath, ep.NewForm).Methods("POST")
	router.HandleFunc(formPath, ep.Forms).Methods("GET")
	router.HandleFunc(formPath, eproxy.AllowCORS).Methods("OPTIONS")
	router.HandleFunc(formIDPath, ep.Form).Methods("GET")
	router.HandleFunc(formIDPath, ep.EditForm).Methods("PUT")
	router.HandleFunc(formIDPath, eproxy.AllowCORS).Methods("OPTIONS")
	router.HandleFunc(formIDPath, ep.DeleteForm).Methods("DELETE")
	router.HandleFunc(formIDPath+"/vote", ep.NewFormVote).Methods("POST")

	router.NotFoundHandler = http.HandlerFunc(eproxy.NotFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(eproxy.NotAllowedHandler)

	proxy.RegisterHandler(formPath, router.ServeHTTP)
	proxy.RegisterHandler(formPathSlash, router.ServeHTTP)

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

// Execute implements node.ActionTemplate. It creates a form and
// simulates the full voting process
func (a *scenarioTestAction) Execute(ctx node.Context) error {
	secretkeyHex := ctx.Flags.String("secretkey")

	secretkeyBuf, err := hex.DecodeString(secretkeyHex)
	if err != nil {
		return xerrors.Errorf("failed to decode secretkeyHex: %v", err)
	}

	secret := suite.Scalar()

	err = secret.UnmarshalBinary(secretkeyBuf)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal secret key: %v", err)
	}

	proxyAddr1 := ctx.Flags.String("proxy-addr1")
	proxyAddr2 := ctx.Flags.String("proxy-addr2")
	proxyAddr3 := ctx.Flags.String("proxy-addr3")

	fmt.Println("Welcome in the scenario test")

	var rosterFac authority.Factory
	err = ctx.Injector.Resolve(&rosterFac)
	if err != nil {
		return xerrors.Errorf("failed to resolve authority factory: %v", err)
	}

	serdecontext := sjson.NewContext()
	formFac := types.NewFormFactory(types.CiphervoteFactory{}, rosterFac)

	var service ordering.Service
	err = ctx.Injector.Resolve(&service)
	if err != nil {
		return xerrors.Errorf("failed to resolve service: %v", err)
	}

	// ###################################### CREATE SIMPLE FORM ######

	formID, form, _, err := setupSimpleForm(ctx, secret,
		proxyAddr1, serdecontext, formFac, service)

	if err != nil {
		return xerrors.Errorf("failed to simple form: %v", err)
	}

	// ##################################### SETUP DKG #########################

	fmt.Fprintln(ctx.Out, "Init DKG")

	proxys := []string{proxyAddr1, proxyAddr2, proxyAddr3}

	// Initializing the DKG for the nodes.
	for i := 1; i <= 3; i++ {
		fmt.Fprintf(ctx.Out, "Node %d ", i)
		err = initDKG(secret, proxys[i], formID)
		if err != nil {
			return xerrors.Errorf("failed to init dkg %d: %v", i, err)
		}

	}

	fmt.Fprintf(ctx.Out, "Setup DKG on node 1")

	_, err = updateDKG(secret, proxyAddr1, formID, "setup")
	if err != nil {
		return xerrors.Errorf("failed to setup dkg on node 1: %v", err)
	}

	// ##################################### OPEN FORM #####################

	fmt.Fprintf(ctx.Out, "Open form")

	_, err = updateForm(secret, proxyAddr1, formID, "open")
	if err != nil {
		return xerrors.Errorf("failed to open form: %v", err)
	}

	// ##################################### GET FORM INFO #################

	fmt.Fprintln(ctx.Out, "Get form")

	form, err = getForm(serdecontext, formFac, formID, service)
	if err != nil {
		return xerrors.Errorf(getFormErr, err)
	}

	logFormStatus(form)
	dela.Logger.Info().Msgf("Pubkey of the form : %x", form.Pubkey)

	// ############################# ATTEMPT TO CLOSE FORM #################

	fmt.Fprintln(ctx.Out, "Close form")

	status, err := updateForm(secret, proxyAddr1, formID, "close")
	if status != http.StatusInternalServerError {
		return xerrors.Errorf("unexpected error: %d: %v", status, err)
	}

	// ##################################### CAST BALLOTS ######################

	fmt.Fprintln(ctx.Out, "cast ballots")

	// Create the ballots
	//b1 := string(selectString + encodeID("bb") + ":0,0,1,0\n" +
		//"text:" + encodeID("ee") + ":eWVz\n\n") //encoding of "yes"

	//b2 := string(selectString + encodeID("bb") + ":1,1,0,0\n" +
		//"text:" + encodeID("ee") + ":amE=\n\n") //encoding of "ja

	//todo b3 := string(selectString + encodeID("bb") + ":0,0,0,1\n" +
	//	"text:" + encodeID("ee") + ":b3Vp\n\n") //encoding of "oui"

	var dkg dkg.DKG
	err = ctx.Injector.Resolve(&dkg)
	if err != nil {
		return xerrors.Errorf("failed to resolve DKG: %v", err)
	}

	//dkgActor, exists := dkg.GetActor(formIDBuf)
	//if !exists {
	//	return xerrors.Errorf("failed to get actor: %v", err)
	//}

	// Ballot 1
//todo	ballot1, err := marshallBallot(b1, dkgActor, form.ChunksPerBallot())
	//if err != nil {
	//	return xerrors.Errorf("failed to marshall ballot : %v", err)
	//}

	castVoteRequest := ptypes.CastVoteRequest{
		UserID: "user1",
//		Ballot: ballot1,
	}

	signed, err := createSignedRequest(secret, castVoteRequest)
	if err != nil {
		return createSignedErr(err)
	}

	fmt.Fprintln(ctx.Out, "cast first ballot")

	respBody, err := castVote(formID, signed, proxyAddr1)
	if err != nil {
		return xerrors.Errorf(castFailed, err)
	}

	dela.Logger.Info().Msg(responseBody + respBody)

	// Ballot 2
	//ballot2, err := marshallBallot(b2, dkgActor, form.ChunksPerBallot())
	//if err != nil {
	//	return xerrors.Errorf("failed to marshall ballot : %v", err)
	//}

	castVoteRequest = ptypes.CastVoteRequest{
		UserID: "user2",
	//	Ballot: ballot2,
	}

	signed, err = createSignedRequest(secret, castVoteRequest)
	if err != nil {
		return createSignedErr(err)
	}

	fmt.Fprintln(ctx.Out, "cast second ballot")

	respBody, err = castVote(formID, signed, proxyAddr1)
	if err != nil {
		return xerrors.Errorf(castFailed, err)
	}

	dela.Logger.Info().Msg(responseBody + respBody)

	// Ballot 3
	//ballot3, err := marshallBallot(b3, dkgActor, form.ChunksPerBallot())
	//if err != nil {
	//	return xerrors.Errorf("failed to marshall ballot: %v", err)
	//}

	castVoteRequest = ptypes.CastVoteRequest{
		UserID: "user3",
	//	Ballot: ballot3,
	}

	signed, err = createSignedRequest(secret, castVoteRequest)
	if err != nil {
		return createSignedErr(err)
	}

	fmt.Fprintln(ctx.Out, "cast third ballot")

	respBody, err = castVote(formID, signed, proxyAddr1)
	if err != nil {
		return xerrors.Errorf(castFailed, err)
	}

	dela.Logger.Info().Msg(responseBody + respBody)

	form, err = getForm(serdecontext, formFac, formID, service)
	if err != nil {
		return xerrors.Errorf(getFormErr, err)
	}

	encryptedBallots := form.Suffragia.Ciphervotes
	dela.Logger.Info().Msg("Length encrypted ballots: " + strconv.Itoa(len(encryptedBallots)))
	dela.Logger.Info().Msgf("Ballot of user1: %s", encryptedBallots[0])
	dela.Logger.Info().Msgf("Ballot of user2: %s", encryptedBallots[1])
	dela.Logger.Info().Msgf("Ballot of user3: %s", encryptedBallots[2])

	// ############################# CLOSE FORM FOR REAL ###################

	fmt.Fprintln(ctx.Out, "Close form (for real)")

	_, err = updateForm(secret, proxyAddr1, formID, "close")
	if err != nil {
		return xerrors.Errorf("failed to close form: %v", err)
	}

	form, err = getForm(serdecontext, formFac, formID, service)
	if err != nil {
		return xerrors.Errorf(getFormErr, err)
	}

	dela.Logger.Info().Msg("Title of the form: " + form.Configuration.MainTitle)
	dela.Logger.Info().Msg("Status of the form: " + strconv.Itoa(int(form.Status)))

	// ###################################### SHUFFLE BALLOTS ##################

	fmt.Fprintln(ctx.Out, "shuffle ballots")

	shuffleRequest := ptypes.UpdateShuffle{
		Action: "shuffle",
	}

	signed, err = createSignedRequest(secret, shuffleRequest)
	if err != nil {
		return createSignedErr(err)
	}

	req, err := http.NewRequest(http.MethodPut, proxyAddr1+"/evoting/services/shuffle/"+formID, bytes.NewBuffer(signed))
	if err != nil {
		return xerrors.Errorf("failed to create shuffle request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return xerrors.Errorf("failed to execute the shuffle query: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected shuffle status: %s - %s", resp.Status, buf)
	}

	// time.Sleep(20 * time.Second)

	form, err = getForm(serdecontext, formFac, formID, service)
	if err != nil {
		return xerrors.Errorf(getFormErr, err)
	}

	logFormStatus(form)
	dela.Logger.Info().Msg("Number of shuffled ballots : " + strconv.Itoa(len(form.ShuffleInstances)))
	dela.Logger.Info().Msg("Number of encrypted ballots : " + strconv.Itoa(len(form.Suffragia.Ciphervotes)))

	// ###################################### REQUEST PUBLIC SHARES ############

	fmt.Fprintln(ctx.Out, "request public shares")

	_, err = updateDKG(secret, proxyAddr1, formID, "computePubshares")
	if err != nil {
		return xerrors.Errorf("failed to compute pubshares: %v", err)
	}

	time.Sleep(10 * time.Second)

	form, err = getForm(serdecontext, formFac, formID, service)
	if err != nil {
		return xerrors.Errorf(getFormErr, err)
	}

	validSubmissions := len(form.PubsharesUnits.Pubshares)

	logFormStatus(form)
	dela.Logger.Info().Msg("Number of Pubshare units submitted: " + strconv.Itoa(validSubmissions))

	// ###################################### DECRYPT BALLOTS ##################

	fmt.Fprintln(ctx.Out, "decrypt ballots")

	_, err = updateForm(secret, proxyAddr1, formID, "combineShares")
	if err != nil {
		return xerrors.Errorf("failed to combine shares: %v", err)
	}

	form, err = getForm(serdecontext, formFac, formID, service)
	if err != nil {
		return xerrors.Errorf(getFormErr, err)
	}

	// dela.Logger.Info().Msg("----------------------- Form : " +
	// string(proof.GetValue()))

	logFormStatus(form)
	dela.Logger.Info().Msg("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))

	// ###################################### GET FORM RESULT ##############

	fmt.Fprintln(ctx.Out, "Get form result")

	form, err = getForm(serdecontext, formFac, formID, service)
	if err != nil {
		return xerrors.Errorf(getFormErr, err)
	}

	logFormStatus(form)
	dela.Logger.Info().Msg("Number of decrypted ballots : " + strconv.Itoa(len(form.DecryptedBallots)))

	if len(form.DecryptedBallots) != 3 {
		return xerrors.Errorf("unexpected number of decrypted ballot: %d != 3", len(form.DecryptedBallots))
	}

	// dela.Logger.Info().Msg(form.DecryptedBallots[0].Vote)
	// dela.Logger.Info().Msg(form.DecryptedBallots[1].Vote)
	// dela.Logger.Info().Msg(form.DecryptedBallots[2].Vote)

	// ###################################### GET ALL FORM ##############

	resp, err = http.Get(proxyAddr1 + formPath)

	if err != nil {
		return xerrors.Errorf("failed to get all forms")
	}

	var allForms ptypes.GetFormsResponse

	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&allForms)
	if err != nil {
		return xerrors.Errorf("failed to decode getAllForms: %v", err)
	}

	dela.Logger.Info().Msgf("All forms: %v", allForms)

	if len(allForms.Forms) != 1 && allForms.Forms[0].FormID != formID {
		return xerrors.Errorf("unexpected allForms: %v", allForms)
	}

	return nil
}

func setupSimpleForm(ctx node.Context, secret kyber.Scalar, proxyAddr1 string,
	serdecontext serde.Context, formFac types.FormFactory,
	service ordering.Service) (string, types.Form, []byte, error) {

	fmt.Fprintln(ctx.Out, "Create form")

	// Define the configuration
	configuration := fake.BasicConfiguration

	createSimpleFormRequest := ptypes.CreateFormRequest{
		Configuration: configuration,
		AdminID:       "adminId",
	}

	signed, err := createSignedRequest(secret, createSimpleFormRequest)
	if err != nil {
		return "", types.Form{}, nil, createSignedErr(err)
	}

	fmt.Fprintln(ctx.Out, "create form js:", signed)

	url := proxyAddr1 + formPath
	fmt.Fprintln(ctx.Out, "POST", url)

	resp, err := http.Post(url, contentType, bytes.NewBuffer(signed))
	if err != nil {
		return "", types.Form{}, nil, xerrors.Errorf(failRetrieveDecryption, err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		return "", types.Form{}, nil, xerrors.Errorf(unexpectedStatus, resp.Status, buf)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", types.Form{}, nil, xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	fmt.Fprintln(ctx.Out, "response body:", string(body))

	resp.Body.Close()

	var formResponse ptypes.CreateFormResponse

	err = json.Unmarshal(body, &formResponse)
	if err != nil {
		return "", types.Form{}, nil, xerrors.Errorf("failed to unmarshal create form response: %v - %s", err, body)
	}

	formID := formResponse.FormID

	formIDBuf, err := hex.DecodeString(formID)
	if err != nil {
		return "", types.Form{}, nil, xerrors.Errorf("failed to decode formID '%s': %v", formID, err)
	}

	form, err := getForm(serdecontext, formFac, formID, service)
	if err != nil {
		return "", types.Form{}, nil, xerrors.Errorf(getFormErr, err)
	}

	// sanity check, the formID returned and the one stored in the form
	// type must be the same.
	if form.FormID != formID {
		return "", types.Form{}, nil, xerrors.Errorf("formID mismatch: %s != %s", form.FormID, formID)
	}

	fmt.Fprintf(ctx.Out, "Title of the form: "+form.Configuration.MainTitle)
	fmt.Fprintf(ctx.Out, "ID of the form: "+form.FormID)
	fmt.Fprintf(ctx.Out, "Status of the form: "+strconv.Itoa(int(form.Status)))

	return formID, form, formIDBuf, nil
}

func logFormStatus(form types.Form) {
	dela.Logger.Info().Msg("Title of the form : " + form.Configuration.MainTitle)
	dela.Logger.Info().Msg("ID of the form : " + form.FormID)
	dela.Logger.Info().Msg("Status of the form : " + strconv.Itoa(int(form.Status)))
}

func encodeID(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
}

/*
func marshallBallot(voteStr string, actor dkg.Actor, chunks int) (ptypes.CiphervoteJSON, error) {

	var ballot = make(ptypes.CiphervoteJSON, chunks)
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
			return ptypes.CiphervoteJSON{}, xerrors.Errorf("failed to encrypt the plaintext: %v", err)
		}

		kbuff, err := K.MarshalBinary()
		if err != nil {
			return ptypes.CiphervoteJSON{}, xerrors.Errorf("failed to marshal K: %v", err)
		}

		cbuff, err := C.MarshalBinary()
		if err != nil {
			return ptypes.CiphervoteJSON{}, xerrors.Errorf("failed to marshal C: %v", err)
		}

		ballot[i] = ptypes.EGPairJSON{
			K: kbuff,
			C: cbuff,
		}
	}

	return ballot, nil
}
*/

// formID is hex-encoded
func castVote(formID string, signed []byte, proxyAddr string) (string, error) {
	resp, err := http.Post(proxyAddr+formPathSlash+formID+"/vote", contentType, bytes.NewBuffer(signed))

	if err != nil {
		return "", xerrors.Errorf(failRetrieveDecryption, err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		return "", xerrors.Errorf(unexpectedStatus, resp.Status, buf)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", xerrors.Errorf("failed to read the body of the response: %v", err)
	}

	resp.Body.Close()

	return string(body), nil
}

func updateForm(secret kyber.Scalar, proxyAddr, formIDHex, action string) (int, error) {
	msg := ptypes.UpdateFormRequest{
		Action: action,
	}

	signed, err := createSignedRequest(secret, msg)
	if err != nil {
		return 0, createSignedErr(err)
	}

	req, err := http.NewRequest(http.MethodPut, proxyAddr+formPathSlash+formIDHex, bytes.NewBuffer(signed))

	if err != nil {
		return 0, xerrors.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, xerrors.Errorf(failRetrieveDecryption, err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, xerrors.Errorf(unexpectedStatus, resp.Status, buf)
	}

	return 0, nil
}

func initDKG(secret kyber.Scalar, proxyAddr, formIDHex string) error {
	setupDKG := ptypes.NewDKGRequest{
		FormID: formIDHex,
	}

	signed, err := createSignedRequest(secret, setupDKG)
	if err != nil {
		return createSignedErr(err)
	}

	resp, err := http.Post(proxyAddr+"/evoting/services/dkg/actors", contentType, bytes.NewBuffer(signed))
	if err != nil {
		return xerrors.Errorf("failed to post request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		return xerrors.Errorf(unexpectedStatus, resp.Status, buf)
	}

	return nil
}

func updateDKG(secret kyber.Scalar, proxyAddr, formIDHex, action string) (int, error) {
	msg := ptypes.UpdateDKG{
		Action: action,
	}

	signed, err := createSignedRequest(secret, msg)
	if err != nil {
		return 0, createSignedErr(err)
	}

	req, err := http.NewRequest(http.MethodPut, proxyAddr+"/evoting/services/dkg/actors/"+formIDHex, bytes.NewBuffer(signed))
	if err != nil {
		return 0, xerrors.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, xerrors.Errorf("failed to execute the query: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, xerrors.Errorf(unexpectedStatus, resp.Status, buf)
	}

	return 0, nil
}

// getForm gets the form from the snap. Returns the form ID NOT hex
// encoded.
func getForm(ctx serde.Context, formFac serde.Factory, formIDHex string,
	srv ordering.Service) (types.Form, error) {

	var form types.Form

	formID, err := hex.DecodeString(formIDHex)
	if err != nil {
		return form, xerrors.Errorf("failed to decode formIDHex: %v", err)
	}

	proof, err := srv.GetProof(formID)
	if err != nil {
		return form, xerrors.Errorf("failed to get proof: %v", err)
	}

	formBuff := proof.GetValue()
	if len(formBuff) == 0 {
		return form, xerrors.Errorf("form does not exist")
	}

	message, err := formFac.Deserialize(ctx, formBuff)
	if err != nil {
		return form, xerrors.Errorf("failed to deserialize Form: %v", err)
	}

	form, ok := message.(types.Form)
	if !ok {
		return form, xerrors.Errorf("wrong message type: %T", message)
	}

	return form, nil
}

func createSignedErr(err error) error {
	return xerrors.Errorf("failed to create signed request: %v", err)
}

func createSignedRequest(secret kyber.Scalar, msg interface{}) ([]byte, error) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal json: %v", err)
	}

	payload := base64.URLEncoding.EncodeToString(jsonMsg)

	hash := sha256.New()

	hash.Write([]byte(payload))
	md := hash.Sum(nil)

	signature, err := schnorr.Sign(suite, secret, md)
	if err != nil {
		return nil, xerrors.Errorf("failed to sign: %v", err)
	}

	signed := ptypes.SignedRequest{
		Payload:   payload,
		Signature: hex.EncodeToString(signature),
	}

	signedJSON, err := json.Marshal(signed)
	if err != nil {
		return nil, xerrors.Errorf("failed to create json signed: %v", err)
	}

	return signedJSON, nil
}
