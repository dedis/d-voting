package proxy

import (
	"encoding/hex"
	"fmt"
	"net/http"

	ptypes "github.com/dedis/d-voting/proxy/types"
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

	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.Verify(s.pk)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to verify signed: %v", err), nil)
		return
	}

	err = s.actor.Shuffle(buff)
	if err != nil {
		http.Error(w, "failed to shuffle: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
