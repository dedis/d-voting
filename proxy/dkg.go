package proxy

import (
	"encoding/hex"
	"fmt"
	"net/http"

	ptypes "github.com/dedis/d-voting/proxy/types"
	dkgSrv "github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"github.com/gorilla/mux"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/kyber/v3"
)

// NewDKG returns a new initialized DKG proxy
func NewDKG(mngr txn.Manager, d dkgSrv.DKG, pk kyber.Point) DKG {
	return dkg{
		mngr: mngr,
		d:    d,
		pk:   pk,
	}
}

// dkg defines the DKG handlers
//
// - implements proxy.DKG
type dkg struct {
	mngr txn.Manager
	d    dkgSrv.DKG
	pk   kyber.Point
}

// NewDKGActor implements proxy.DKG
func (d dkg) NewDKGActor(w http.ResponseWriter, r *http.Request) {
	var req ptypes.NewDKGRequest

	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.GetAndVerify(d.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	electionIDBuf, err := hex.DecodeString(req.ElectionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+req.ElectionID,
			http.StatusBadRequest)
		return
	}

	_, err = d.d.Listen(electionIDBuf, d.mngr)
	if err != nil {
		http.Error(w, "failed to start actor: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// EditDKGActor implements proxy.DKG
func (d dkg) EditDKGActor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars == nil || vars["electionID"] == "" {
		http.Error(w, fmt.Sprintf("electionID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	electionID := vars["electionID"]

	electionIDBuf, err := hex.DecodeString(electionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+electionID, http.StatusBadRequest)
		return
	}

	a, exists := d.d.GetActor(electionIDBuf)
	if !exists {
		http.Error(w, "actor does not exist", http.StatusInternalServerError)
		return
	}

	var req types.UpdateDKG

	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.GetAndVerify(d.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	switch req.Action {
	case "setup":
		_, err := a.Setup()
		if err != nil {
			http.Error(w, "failed to setup: "+err.Error(), http.StatusInternalServerError)
			return
		}
	case "computePubshares":
		err = a.ComputePubshares()
		if err != nil {
			http.Error(w, "failed to compute pubshares: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
