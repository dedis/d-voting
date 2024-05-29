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

	"github.com/dedis/d-voting/proxy/types"
	"go.dedis.ch/kyber/v3/suites"
)

var suite = suites.MustFind("ed25519")

// Form defines the public HTTP API for the form smart contract
type Form interface {
	// POST /forms
	NewForm(http.ResponseWriter, *http.Request)
	// POST /forms/{formID}/vote
	NewFormVote(http.ResponseWriter, *http.Request)
	// PUT /forms/{formID}
	EditForm(http.ResponseWriter, *http.Request)
	// GET /forms
	Forms(http.ResponseWriter, *http.Request)
	// GET /forms/{formID}
	Form(http.ResponseWriter, *http.Request)
	// DELETE /forms/{formID}
	DeleteForm(http.ResponseWriter, *http.Request)
	// TODO CHECK CAUSE NEW -> modif according to blockchain
	// POST /addadmin
	AddAdmin(http.ResponseWriter, *http.Request)
	// POST /removeadmin
	RemoveAdmin(http.ResponseWriter, *http.Request)
	// POST /forms/{formID}/addowner
	AddOwnerToForm(http.ResponseWriter, *http.Request)
	// POST /forms/{formID}/removeowner
	RemoveOwnerToForm(http.ResponseWriter, *http.Request)
	// POST /forms/{formID}/addvoter
	AddVoterToForm(http.ResponseWriter, *http.Request)
	// POST /forms/{formID}/removevoter
	RemoveVoterToForm(http.ResponseWriter, *http.Request)
}

// DKG defines the public HTTP API of the DKG service
type DKG interface {
	// POST /services/dkg
	NewDKGActor(http.ResponseWriter, *http.Request)
	// GET /services/dkg/{formID}
	Actor(http.ResponseWriter, *http.Request)
	// PUT /services/dkg/{formID}
	EditDKGActor(http.ResponseWriter, *http.Request)
}

// Shuffle defines the public HTTP API of the shuffling service
type Shuffle interface {
	// PUT /services/shuffle/{formID}
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

// InternalError sets an internal server error
func InternalError(w http.ResponseWriter, r *http.Request, err error, args map[string]interface{}) {
	httpErr(w, r, err, http.StatusInternalServerError, "Internal server error", args)
}

// BadRequestError sets an bad request error
func BadRequestError(w http.ResponseWriter, r *http.Request, err error, args map[string]interface{}) {
	httpErr(w, r, err, http.StatusBadRequest, "bad request", args)
}

// ForbiddenError sets a forbidden error error
func ForbiddenError(w http.ResponseWriter, r *http.Request, err error, args map[string]interface{}) {
	httpErr(w, r, err, http.StatusForbidden, "not authorized / forbidden", args)
}

// NotFoundErr sets a not found error
func NotFoundErr(w http.ResponseWriter, r *http.Request, err error, args map[string]interface{}) {
	httpErr(w, r, err, http.StatusNotFound, "not found", args)
}

func httpErr(w http.ResponseWriter, r *http.Request, err error, code uint, title string, args map[string]interface{}) {
	if args == nil {
		args = make(map[string]interface{})
	}

	args["error"] = err.Error()
	args["url"] = r.URL.String()
	args["method"] = r.Method

	errMsg := types.HTTPError{
		Title:   title,
		Code:    code,
		Message: "A problem occurred on the proxy",
		Args:    args,
	}

	buf, _ := json.MarshalIndent(&errMsg, "", "  ")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(int(code))
	fmt.Fprintln(w, string(buf))
}

// AllowCORS defines a basic handler that adds wide Access Control Allow origin
// headers.
func AllowCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}
