package proxy

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dedis/d-voting/proxy/types"
	dkgSrv "github.com/dedis/d-voting/services/dkg"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"golang.org/x/xerrors"
)

// test that NewDkg is working properly
func TestNewDkg(t *testing.T) {
	ctx := node.Context{
		Injector: node.NewInjector(),
	}

	//create the txn.Manager
	var mngr txn.Manager
	ctx.Injector.Inject(&mngr)
	var d dkgSrv.DKG
	ctx.Injector.Inject(&d)
	var pk kyber.Point
	ctx.Injector.Inject(&pk)

	dkgInterface := NewDKG(mngr, d, pk)
	//check that the dkg is not nil
	require.NotNil(t, dkgInterface)
	//the txn.Manager of the dkg should be the same as the one we injected$
	require.Equal(t, mngr, dkgInterface.(dkg).manager)
	//the dkg of the dkg should be the same as the one we injected
	require.Equal(t, d, dkgInterface.(dkg).dkgService)
	//the pk of the dkg should be the same as the one we injected
	require.Equal(t, pk, dkgInterface.(dkg).pk)
}

// test that NewDKGActor is working properly
func TestNewDKGActor(t *testing.T) {

	var w http.ResponseWriter = httptest.NewRecorder()

	//print(r.Body)

	ctx := node.Context{
		Injector: node.NewInjector(),
	}

	//create the txn.Manager
	var mngr txn.Manager
	ctx.Injector.Inject(&mngr)

	pk, err := hex.DecodeString("adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3")
	require.NoError(t, err)

	public := suite.Point()
	err = public.UnmarshalBinary(pk)
	require.NoError(t, err)

	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err)

	request := types.NewDKGRequest{
		FormID: "abcd",
	}

	dkgInterface := NewDKG(mngr, mockDKGService{}, public)

	requestt, e := createSignedRequest(secret, request)
	require.NoError(t, e)

	r, e := http.NewRequest("POST", "/dkg", strings.NewReader(string(requestt)))
	require.NoError(t, e)

	dkgInterface.NewDKGActor(w, r)

	require.Equal(t, 200, w.(*httptest.ResponseRecorder).Result().StatusCode)

	//formIDBuffer, err := hex.DecodeString("abcd")
	//require.NoError(t, err)

	//recorder:=httptest.NewRecorder()

	//dkgInterface.Actor(recorder, httptest.NewRequest( "GET", "/services/dkg/"+request.FormID, nil))

	//require.Equal(t, 200, recorder.Result().StatusCode)

}

// test that NewDKGActor is setting the right status code when the request is not valid
func TestNewDKGActorInvalidRequest(t *testing.T) {
	var w http.ResponseWriter = httptest.NewRecorder()

	ctx := node.Context{
		Injector: node.NewInjector(),
	}

	//create the txn.Manager
	var mngr txn.Manager
	ctx.Injector.Inject(&mngr)

	pk, err := hex.DecodeString("adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3")
	require.NoError(t, err)

	public := suite.Point()
	err = public.UnmarshalBinary(pk)
	require.NoError(t, err)

	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err)

	dkgInterface := NewDKG(mngr, mockDKGService{}, public)

	r, e := http.NewRequest("POST", "/dkg", strings.NewReader("abcd"))
	if e != nil {
		t.Fatal(e)
	}

	dkgInterface.NewDKGActor(w, r)

	require.Equal(t, 500, w.(*httptest.ResponseRecorder).Result().StatusCode)

}

// test that NewDKGActor is setting the right status code when the request cannot be verified
func TestNewDKGActorInvalidSignature(t *testing.T) {
	var w http.ResponseWriter = httptest.NewRecorder()

	//print(r.Body)

	ctx := node.Context{
		Injector: node.NewInjector(),
	}

	//create the txn.Manager
	var mngr txn.Manager
	ctx.Injector.Inject(&mngr)

	pk, err := hex.DecodeString("badbad10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3")
	require.NoError(t, err)

	public := suite.Point()
	err = public.UnmarshalBinary(pk)
	require.NoError(t, err)

	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err)

	request := types.NewDKGRequest{
		FormID: "abcd",
	}

	dkgInterface := NewDKG(mngr, mockDKGService{}, public)

	requestt, err := createSignedRequest(secret, request)
	require.NoError(t, err)

	r, err := http.NewRequest("POST", "/dkg", strings.NewReader(string(requestt)))
	require.NoError(t, err)

	dkgInterface.NewDKGActor(w, r)

	require.Equal(t, 500, w.(*httptest.ResponseRecorder).Result().StatusCode)

}

// test that NewDKGActor is setting the right status code when the formID cannot be decoded
func TestNewDKGActorInvalidFormID(t *testing.T) {
	var w http.ResponseWriter = httptest.NewRecorder()

	//print(r.Body)

	ctx := node.Context{
		Injector: node.NewInjector(),
	}

	//create the txn.Manager
	var mngr txn.Manager
	ctx.Injector.Inject(&mngr)

	pk, err := hex.DecodeString("adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3")
	require.NoError(t, err)

	public := suite.Point()
	err = public.UnmarshalBinary(pk)
	require.NoError(t, err)

	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err)

	request := types.NewDKGRequest{
		FormID: "abcdefg",
	}

	dkgInterface := NewDKG(mngr, mockDKGService{}, public)

	requestt, err := createSignedRequest(secret, request)

	require.NoError(t, err)

	r, err := http.NewRequest("POST", "/dkg", strings.NewReader(string(requestt)))

	require.NoError(t, err)

	dkgInterface.NewDKGActor(w, r)

	require.Equal(t, 400, w.(*httptest.ResponseRecorder).Result().StatusCode)

}

// test that NewDKGActor is setting the right status code when the form does not exist
func TestNewDKGActorFormDoesNotExist(t *testing.T) {
	var w http.ResponseWriter = httptest.NewRecorder()

	//print(r.Body)

	ctx := node.Context{
		Injector: node.NewInjector(),
	}

	//create the txn.Manager
	var mngr txn.Manager
	ctx.Injector.Inject(&mngr)

	pk, err := hex.DecodeString("adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3")
	require.NoError(t, err)

	public := suite.Point()
	err = public.UnmarshalBinary(pk)
	require.NoError(t, err)

	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err)

	request := types.NewDKGRequest{
		FormID: "abcd",
	}

	dkgInterface := NewDKG(mngr, mockDKGServiceError{}, public)

	requestt, err := createSignedRequest(secret, request)
	
	require.NoError(t, err)


	r, err := http.NewRequest("POST", "/dkg", strings.NewReader(string(requestt)))

	require.NoError(t, err)

	dkgInterface.NewDKGActor(w, r)

	require.Equal(t, 500, w.(*httptest.ResponseRecorder).Result().StatusCode)
}

// test that Actor is working correctly
func TestActor(t *testing.T) {
	var w http.ResponseWriter = httptest.NewRecorder()

	//print(r.Body)

	ctx := node.Context{
		Injector: node.NewInjector(),
	}

	//create the txn.Manager
	var mngr txn.Manager
	ctx.Injector.Inject(&mngr)

	pk, err := hex.DecodeString("adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3")
	require.NoError(t, err)

	public := suite.Point()
	err = public.UnmarshalBinary(pk)
	require.NoError(t, err)

	secretkeyBuf, err := hex.DecodeString("28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409")
	require.NoError(t, err)

	secret := suite.Scalar()
	err = secret.UnmarshalBinary(secretkeyBuf)
	require.NoError(t, err)

	request := types.NewDKGRequest{
		FormID: "abcd",
	}

	dkgInterface := NewDKG(mngr, mockDKGService{}, public)

	requestt, err:= createSignedRequest(secret, request)
	require.NoError(t, err)

	r, err := http.NewRequest("GET", "/services/dkg/actors/1234", strings.NewReader(string(requestt)))
	require.NoError(t, err)

	dkgInterface.Actor(w, r)

	//require.Equal(t, 200, w.(*httptest.ResponseRecorder).Result().Status)

}

// ---------------------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------
// Auxiliar functions
//

// mock dkgService
type mockDKGService struct {
	dkgSrv.DKG
}

func (m mockDKGService) Listen(_ []byte, _ txn.Manager) (dkgSrv.Actor, error) {
	return nil, nil
}
func (m mockDKGService) GetActor(formID []byte) (dkgSrv.Actor, bool) {
	return nil, false
}

// mock dkgService that returns an error
type mockDKGServiceError struct {
	dkgSrv.DKG
}

func (m mockDKGServiceError) Listen(_ []byte, _ txn.Manager) (dkgSrv.Actor, error) {
	return nil, errors.New("error")
}

func (m mockDKGServiceError) GetActor(formID []byte) (dkgSrv.Actor, bool) {
	return nil, false
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

	signed := types.SignedRequest{
		Payload:   payload,
		Signature: hex.EncodeToString(signature),
	}

	signedJSON, err := json.Marshal(signed)
	if err != nil {
		return nil, xerrors.Errorf("failed to create json signed: %v", err)
	}

	return signedJSON, nil
}
