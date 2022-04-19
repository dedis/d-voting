package proxy

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/dedis/d-voting/proxy/types"
	shuffleSrv "github.com/dedis/d-voting/services/shuffle"
	"github.com/gorilla/mux"
	"go.dedis.ch/kyber/v3"
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
	actor shuffleSrv.Actor
	pk    kyber.Point
}

// EditShuffle implements proxy.Shuffle
func (s shuffle) EditShuffle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars == nil || vars["electionID"] == "" {
		http.Error(w, fmt.Sprintf("electionID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	electionID := vars["electionID"]

	buff, err := hex.DecodeString(electionID)
	if err != nil {
		http.Error(w, "failed to decode electionID: "+electionID, http.StatusInternalServerError)
		return
	}

	var req types.UpdateShuffle

	signed, err := types.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.GetAndVerify(s.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	switch req.Action {
	case "shuffle":
		err = s.actor.Shuffle(buff)
		if err != nil {
			http.Error(w, "failed to shuffle: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
