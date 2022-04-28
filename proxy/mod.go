// Package proxy defines and implement the public APIs.
//
// Function names follow the convention used in URL helpers on rails:
// https://guides.rubyonrails.org/routing.html#path-and-url-helpers
//
// For the API specification look at /docs/api.md.
package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dedis/d-voting/proxy/types"
	"go.dedis.ch/kyber/v3/suites"
)

const inclusionTimeout = 2 * time.Second

var suite = suites.MustFind("ed25519")

// Election defines the public HTTP API for the election smart contract
type Election interface {
	// POST /elections
	NewElection(http.ResponseWriter, *http.Request)
	// POST /elections/{electionID}/vote
	NewElectionVote(http.ResponseWriter, *http.Request)
	// PUT /elections/{electionID}
	EditElection(http.ResponseWriter, *http.Request)
	// GET /elections
	Elections(http.ResponseWriter, *http.Request)
	// GET /elections/{electionID}
	Election(http.ResponseWriter, *http.Request)
}

// DKG defines the public HTTP API of the DKG service
type DKG interface {
	// POST /services/dkg
	NewDKGActor(http.ResponseWriter, *http.Request)
	// PUT /services/dkg/{electionID}
	EditDKGActor(http.ResponseWriter, *http.Request)
}

// Shuffle defines the public HTTP API of the shuffling service
type Shuffle interface {
	// PUT /services/shuffle/{electionID}
	EditShuffle(http.ResponseWriter, *http.Request)
}

// NotFoundHandler defines a generic handler for 404
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	err := types.HTTPError{
		Title:   "Not found",
		Code:    http.StatusNotFound,
		Message: "The requested endpoint was not found",
		Args: map[string]interface{}{
			"url":    r.URL.String(),
			"method": r.Method,
		},
	}

	buf, _ := json.MarshalIndent(&err, "", "  ")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, string(buf))
}

// NotAllowedHandler degines a generic handler for 405
func NotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	err := types.HTTPError{
		Title:   "Not allowed",
		Code:    http.StatusMethodNotAllowed,
		Message: "The requested endpoint was not allowed",
		Args: map[string]interface{}{
			"url":    r.URL.String(),
			"method": r.Method,
		},
	}

	buf, _ := json.MarshalIndent(&err, "", "  ")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(w, string(buf))
}

// InternalError set an internal server error
func InternalError(w http.ResponseWriter, r *http.Request, err error, args map[string]interface{}) {
	if args == nil {
		args = make(map[string]interface{})
	}

	args["error"] = err.Error()
	args["url"] = r.URL.String()
	args["method"] = r.Method

	errMsg := types.HTTPError{
		Title:   "Internal server error",
		Code:    http.StatusInternalServerError,
		Message: "A problem occurred on the proxy",
		Args:    args,
	}

	buf, _ := json.MarshalIndent(&errMsg, "", "  ")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, string(buf))
}

// AllowCORS defines a basic handler that adds wide Access Control Allow origin
// headers.
func AllowCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}
