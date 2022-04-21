package proxy

import (
	"encoding/hex"
	"fmt"
	"net/http"

	shuffleSrv "github.com/dedis/d-voting/services/shuffle"
	"github.com/gorilla/mux"
)

// NewShuffle returns a new initialized shuffle
func NewShuffle(actor shuffleSrv.Actor) Shuffle {
	return shuffle{
		actor: actor,
	}
}

// shuffle defines the proxy handlers for the shuffling service
//
// - implements proxy.Shuffle
type shuffle struct {
	actor shuffleSrv.Actor
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

	err = s.actor.Shuffle(buff)
	if err != nil {
		http.Error(w, "failed to shuffle: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
