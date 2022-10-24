package proxy

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/dedis/d-voting/proxy/types"
	shuffleSrv "github.com/dedis/d-voting/services/shuffle"
	"github.com/gorilla/mux"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

// NewShuffle returns a new initialized shuffle
func NewShuffle(actor shuffleSrv.Actor, pk kyber.Point) Shuffle {
	return shuffle{
		actor: actor,
		pk:    pk,
	}
}

// shuffle defines the proxy handlers for the shuffling service
//
// - implements proxy.Shuffle
type shuffle struct {
	// actor - the shuffle actor
	actor shuffleSrv.Actor
	// pk - the public key of the proxy
	pk kyber.Point
}

// EditShuffle implements proxy.Shuffle
func (s shuffle) EditShuffle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// check if the formID is present
	if vars == nil || vars["formID"] == "" {
		http.Error(w, fmt.Sprintf("formID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	formID := vars["formID"]

	// decode the formID
	buff, err := hex.DecodeString(formID)
	if err != nil {
		http.Error(w, "failed to decode formID: "+formID, http.StatusInternalServerError)
		return
	}

	var req types.UpdateShuffle

	// Read the request
	signed, err := types.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	// Verify the signature and get the request
	err = signed.GetAndVerify(s.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	switch req.Action {
	// shuffle the ballots
	case "shuffle":
		err = s.actor.Shuffle(buff)
		if err != nil {
			http.Error(w, "failed to shuffle: "+err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		BadRequestError(w, r, xerrors.Errorf("invalid action: %s", req.Action), nil)
		return
	}
}
