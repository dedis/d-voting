package proxy

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	dkgSrv "github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"github.com/gorilla/mux"
	"go.dedis.ch/dela/core/txn"
)

// NewDKG returns a new initialized DKG proxy
func NewDKG(mngr txn.Manager, d dkgSrv.DKG) DKG {
	return dkg{
		mngr: mngr,
		d:    d,
	}
}

// dkg defines the DKG handlers
//
// - implements proxy.DKG
type dkg struct {
	mngr txn.Manager
	d    dkgSrv.DKG
}

// NewDKGActor implements proxy.DKG
func (d dkg) NewDKGActor(w http.ResponseWriter, r *http.Request) {
	// Receive the hex-encoded electionID
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	electionID := string(data)

	// sanity check
	electionIDBuf, err := hex.DecodeString(electionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+electionID,
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

	var input types.UpdateDKG

	decoder := json.NewDecoder(r.Body)

	err = decoder.Decode(&input)
	if err != nil {
		http.Error(w, "failed to decode input: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch input.Action {
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
