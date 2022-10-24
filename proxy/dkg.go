package proxy

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dedis/d-voting/proxy/types"
	dkgSrv "github.com/dedis/d-voting/services/dkg"
	"github.com/gorilla/mux"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

// NewDKG returns a new initialized DKG proxy
func NewDKG(mngr txn.Manager, d dkgSrv.DKG, pk kyber.Point) DKG {
	return dkg{
		manager:    mngr,
		dkgService: d,
		pk:         pk,
	}
}

// dkg defines the DKG handlers
//
// - implements proxy.DKG
type dkg struct {
	// manager is the transaction manager
	manager txn.Manager
	// dkgService is the DKG service
	dkgService dkgSrv.DKG
	// pk is the public key of the proxy
	pk kyber.Point
}

// NewDKGActor implements proxy.DKG
// Create a new DKG actor for the given electionID
func (d dkg) NewDKGActor(w http.ResponseWriter, r *http.Request) {
	var req types.NewDKGRequest

	// Read the request
	signed, err := types.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	// Verify the request
	err = signed.GetAndVerify(d.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	formIDBuf, err := hex.DecodeString(req.FormID)
	if err != nil {
		http.Error(w, "failed to decode formID: "+req.FormID,
			http.StatusBadRequest)
		return
	}

	// subscribe to the DKG service
	_, err = d.dkgService.Listen(formIDBuf, d.manager)
	if err != nil {
		http.Error(w, "failed to start actor: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// Actor implements proxy.DKG
// Send the actor status
func (d dkg) Actor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	vars := mux.Vars(r)

	// check if the formID is present
	if vars == nil || vars["formID"] == "" {
		http.Error(w, fmt.Sprintf("formID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	formID := vars["formID"]

	formIDBuf, err := hex.DecodeString(formID)
	if err != nil {
		BadRequestError(w, r, xerrors.Errorf("failed to decode formID: %v", err), nil)
		return
	}

	actor, found := d.dkgService.GetActor(formIDBuf)
	if !found {
		NotFoundErr(w, r, xerrors.New("actor not found"), nil)
		return
	}

	// get the status
	status := actor.Status()
	var httpErr types.HTTPError

	// if the status has an error, return it
	if status.Err != nil {
		httpErr = types.HTTPError{
			Title:   "Setup failed",
			Code:    0,
			Message: status.Err.Error(),
			Args:    status.Args,
		}
	}

	// return the status
	response := types.GetActorInfo{
		Status: int(status.Status),
		Error:  httpErr,
	}

	w.Header().Set("Content-Type", "application/json")

	// encode the response
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to write response: %v", err), nil)
		return
	}
}

// EditDKGActor implements proxy.DKG
// Setups the DKG actor or begins decryption
func (d dkg) EditDKGActor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars == nil || vars["formID"] == "" {
		http.Error(w, fmt.Sprintf("formID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	formID := vars["formID"]

	formIDBuf, err := hex.DecodeString(formID)
	if err != nil {
		http.Error(w, "failed to decode formID: "+formID, http.StatusBadRequest)
		return
	}

	// get the actor
	a, exists := d.dkgService.GetActor(formIDBuf)
	if !exists {
		http.Error(w, "actor does not exist", http.StatusInternalServerError)
		return
	}

	var req types.UpdateDKG

	// Read the request
	signed, err := types.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	// Verify the signature
	err = signed.GetAndVerify(d.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	switch req.Action {
	// setup the DKG
	case "setup":
		// As the setup can be long, we run it asynchronously. One can fetch the
		// status of the actor to know when the setup is over.
		go func() {
			_, err := a.Setup()
			if err != nil {
				dela.Logger.Err(err).Msg("failed to setup")
			}
		}()
		// begin the decryption
	case "computePubshares":
		err = a.ComputePubshares()
		if err != nil {
			http.Error(w, "failed to compute pubshares: "+err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		BadRequestError(w, r, xerrors.Errorf("invalid action: %s", req.Action), nil)
		return
	}
}
