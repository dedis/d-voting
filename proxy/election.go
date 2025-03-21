package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.dedis.ch/d-voting/contracts/evoting"
	"go.dedis.ch/d-voting/contracts/evoting/types"
	"go.dedis.ch/d-voting/proxy/txnmanager"
	ptypes "go.dedis.ch/d-voting/proxy/types"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"golang.org/x/xerrors"
)

func newSignedErr(err error) error {
	return xerrors.Errorf("failed to created signed request: %v", err)
}

func getSignedErr(err error) error {
	return xerrors.Errorf("failed to get and verify signed request: %v", err)
}

// NewForm returns a new initialized form proxy
func NewForm(srv ordering.Service, p pool.Pool,
	ctx serde.Context, fac serde.Factory, pk kyber.Point, txnManaxer txnmanager.Manager) Form {

	logger := dela.Logger.With().Timestamp().Str("role", "evoting-proxy").Logger()

	return &form{
		logger:      logger,
		orderingSvc: srv,
		context:     ctx,
		formFac:     fac,
		mngr:        txnManaxer,
		pool:        p,
		pk:          pk,
	}
}

// form defines HTTP handlers to manipulate the evoting smart contract
//
// - implements proxy.Form
type form struct {
	sync.Mutex

	orderingSvc ordering.Service
	logger      zerolog.Logger
	context     serde.Context
	formFac     serde.Factory
	mngr        txnmanager.Manager
	pool        pool.Pool
	pk          kyber.Point
}

// NewForm implements proxy.Proxy
func (h *form) NewForm(w http.ResponseWriter, r *http.Request) {
	var req ptypes.CreateFormRequest

	// get the signed request
	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	// get the request and verify the signature
	err = signed.GetAndVerify(h.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	createForm := types.CreateForm{
		Configuration: req.Configuration,
		AdminID:       req.AdminID,
	}

	// serialize the transaction
	data, err := createForm.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal CreateFormTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, blockIdx, err := h.mngr.SubmitTxn(r.Context(), evoting.CmdCreateForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// hash the transaction
	hash := sha256.New()
	hash.Write(txnID)
	formID := hash.Sum(nil)

	// create it to get the  token
	transactionClientInfo, err := h.mngr.CreateTransactionResult(txnID, blockIdx, txnmanager.UnknownTransactionStatus)
	if err != nil {
		http.Error(w, "failed to create transaction info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := ptypes.CreateFormResponse{
		FormID: hex.EncodeToString(formID),
		Token:  transactionClientInfo.Token,
	}

	// send the response json
	err = txnmanager.SendResponse(w, response)
	if err != nil {
		fmt.Printf("Caught unhandled error: %+v", err)
	}
}

// NewFormVote implements proxy.Proxy
func (h *form) NewFormVote(w http.ResponseWriter, r *http.Request) {
	var req ptypes.CastVoteRequest

	// get the signed request
	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	// get the request and verify the signature
	err = signed.GetAndVerify(h.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	vars := mux.Vars(r)

	// check if the formID is valid
	if vars == nil || vars["formID"] == "" {
		http.Error(w, fmt.Sprintf("formID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	formID := vars["formID"]

	elecMD, err := h.getFormsMetadata()
	if err != nil {
		http.Error(w, "failed to get form metadata", http.StatusNotFound)
		return
	}

	// check if the form exist
	if elecMD.FormsIDs.Contains(formID) < 0 {
		http.Error(w, "the form does not exist", http.StatusNotFound)
		return
	}

	ciphervote := make(types.Ciphervote, len(req.Ballot))

	// unmarshal the encrypted ballot
	for i, egpair := range req.Ballot {
		k := suite.Point()

		err = k.UnmarshalBinary(egpair.K)
		if err != nil {
			http.Error(w, "failed to unmarshal K: "+err.Error(), http.StatusInternalServerError)
			return
		}

		c := suite.Point()

		err = c.UnmarshalBinary(egpair.C)
		if err != nil {
			http.Error(w, "failed to unmarshal C: "+err.Error(), http.StatusInternalServerError)
			return
		}

		ciphervote[i] = types.EGPair{
			K: k,
			C: c,
		}
	}

	castVote := types.CastVote{
		FormID: formID,
		UserID: req.UserID,
		Ballot: ciphervote,
	}

	// serialize the vote
	data, err := castVote.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal CastVoteTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := h.mngr.SubmitTxn(r.Context(), evoting.CmdCastVote, evoting.FormArg, data)
	if err != nil {
		h.logger.Err(err).Msg("failed to submit txn")
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's information
	err = h.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
	if err != nil {
		http.Error(w, "couldn't send transaction info: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// EditForm implements proxy.Proxy
func (h *form) EditForm(w http.ResponseWriter, r *http.Request) {
	var req ptypes.UpdateFormRequest

	// get the signed request
	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	// get the request and verify the signature
	err = signed.GetAndVerify(h.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	vars := mux.Vars(r)

	// check if the formID is valid
	if vars == nil || vars["formID"] == "" {
		http.Error(w, fmt.Sprintf("formID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	formID := vars["formID"]

	elecMD, err := h.getFormsMetadata()
	if err != nil {
		http.Error(w, "failed to get form metadata", http.StatusNotFound)
		return
	}

	// check if the form exists
	if elecMD.FormsIDs.Contains(formID) < 0 {
		http.Error(w, "the form does not exist", http.StatusNotFound)
		return
	}

	switch req.Action {
	case "open":
		h.openForm(formID, w, r)
	case "close":
		h.closeForm(formID, w, r)
	case "combineShares":
		h.combineShares(formID, w, r)
	case "cancel":
		h.cancelForm(formID, w, r)
	default:
		BadRequestError(w, r, xerrors.Errorf("invalid action: %s", req.Action), nil)
		return
	}
}

// openForm allows opening a form, which sets the public key based on
// the DKG actor.
func (h *form) openForm(formID string, w http.ResponseWriter, r *http.Request) {
	openForm := types.OpenForm{
		FormID: formID,
	}

	// serialize the transaction
	data, err := openForm.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal OpenFormTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := h.mngr.SubmitTxn(r.Context(), evoting.CmdOpenForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's informations
	h.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
}

// closeForm closes a form.
func (h *form) closeForm(formIDHex string, w http.ResponseWriter, r *http.Request) {

	closeForm := types.CloseForm{
		FormID: formIDHex,
	}

	// serialize the transaction
	data, err := closeForm.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal CloseFormTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := h.mngr.SubmitTxn(r.Context(), evoting.CmdCloseForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's informations
	h.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)

}

// combineShares decrypts the shuffled ballots in a form.
func (h *form) combineShares(formIDHex string, w http.ResponseWriter, r *http.Request) {

	form, err := types.FormFromStore(h.context, h.formFac, formIDHex, h.orderingSvc.GetStore())
	if err != nil {
		http.Error(w, "failed to get form: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	if form.Status != types.PubSharesSubmitted {
		http.Error(w, "the submission of public shares must be over!",
			http.StatusUnauthorized)
		return
	}

	decryptBallots := types.CombineShares{
		FormID: formIDHex,
	}

	// serialize the transaction
	data, err := decryptBallots.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal decryptBallots: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := h.mngr.SubmitTxn(r.Context(), evoting.CmdCombineShares, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's informations
	h.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
}

// cancelForm cancels a form.
func (h *form) cancelForm(formIDHex string, w http.ResponseWriter, r *http.Request) {

	cancelForm := types.CancelForm{
		FormID: formIDHex,
	}

	// serialize the transaction
	data, err := cancelForm.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal CancelForm: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := h.mngr.SubmitTxn(r.Context(), evoting.CmdCancelForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's informations
	h.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
}

// Form implements proxy.Proxy. The request should not be signed because it
// is fetching public data.
func (h *form) Form(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	vars := mux.Vars(r)

	// check if the form exists
	if vars == nil || vars["formID"] == "" {
		http.Error(w, fmt.Sprintf("formID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	formID := vars["formID"]

	// get the form
	form, err := types.FormFromStore(h.context, h.formFac, formID, h.orderingSvc.GetStore())
	if err != nil {
		http.Error(w, xerrors.Errorf("failed to get form: %v", err).Error(), http.StatusInternalServerError)
		return
	}

	var pubkeyBuf []byte

	// get the public key
	if form.Pubkey != nil {
		pubkeyBuf, err = form.Pubkey.MarshalBinary()
		if err != nil {
			http.Error(w, "failed to marshal pubkey: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
	}

	roster := make([]string, 0, form.Roster.Len())

	iter := form.Roster.AddressIterator()
	for iter.HasNext() {
		roster = append(roster, iter.GetNext().String())
	}

	suff, err := form.Suffragia(h.context, h.orderingSvc.GetStore())
	if err != nil {
		http.Error(w, "couldn't get ballots: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	response := ptypes.GetFormResponse{
		FormID:          string(form.FormID),
		Configuration:   form.Configuration,
		Status:          uint16(form.Status),
		Pubkey:          hex.EncodeToString(pubkeyBuf),
		Result:          form.DecryptedBallots,
		Roster:          roster,
		ChunksPerBallot: form.ChunksPerBallot(),
		BallotSize:      form.BallotSize,
		Voters:          suff.UserIDs,
	}

	txnmanager.SendResponse(w, response)

}

// Forms implements proxy.Proxy. The request should not be signed because it
// is fecthing public data.
func (h *form) Forms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	elecMD, err := h.getFormsMetadata()
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to get form metadata: %v", err), nil)
		return
	}

	allFormsInfo := make([]ptypes.LightForm, len(elecMD.FormsIDs))

	// get the forms
	for i, id := range elecMD.FormsIDs {
		form, err := types.FormFromStore(h.context, h.formFac, id, h.orderingSvc.GetStore())
		if err != nil {
			InternalError(w, r, xerrors.Errorf("failed to get form: %v", err), nil)
			return
		}

		var pubkeyBuf []byte

		if form.Pubkey != nil {
			pubkeyBuf, err = form.Pubkey.MarshalBinary()
			if err != nil {
				InternalError(w, r, xerrors.Errorf("failed to marshal pubkey: %v", err), nil)
				return
			}
		}

		info := ptypes.LightForm{
			FormID: string(form.FormID),
			Title:  form.Configuration.Title,
			Status: uint16(form.Status),
			Pubkey: hex.EncodeToString(pubkeyBuf),
		}

		allFormsInfo[i] = info
	}

	response := ptypes.GetFormsResponse{Forms: allFormsInfo}

	txnmanager.SendResponse(w, response)

}

// DeleteForm implements proxy.Proxy
func (h *form) DeleteForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// check if the formID is valid
	if vars == nil || vars["formID"] == "" {
		http.Error(w, fmt.Sprintf("formID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	formID := vars["formID"]

	elecMD, err := h.getFormsMetadata()
	if err != nil {
		http.Error(w, "failed to get form metadata", http.StatusNotFound)
		return
	}

	// check if the form exists
	if elecMD.FormsIDs.Contains(formID) < 0 {
		http.Error(w, "the form does not exist", http.StatusNotFound)
		return
	}

	// auth should contain the hex-encoded signature on the hex-encoded form
	// ID
	auth := r.Header.Get("Authorization")

	signature, err := hex.DecodeString(auth)
	if err != nil {
		BadRequestError(w, r, xerrors.Errorf("failed to decode auth: %v", err), nil)
		return
	}

	// check if the signature is valid
	err = schnorr.Verify(suite, h.pk, []byte(formID), signature)
	if err != nil {
		ForbiddenError(w, r, xerrors.Errorf("signature verification failed: %v", err), nil)
		return
	}

	deleteForm := types.DeleteForm{
		FormID: formID,
	}

	data, err := deleteForm.Serialize(h.context)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to marshal DeleteForm: %v", err), nil)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := h.mngr.SubmitTxn(r.Context(), evoting.CmdDeleteForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's information
	h.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
}

func (h *form) getFormsMetadata() (types.FormsMetadata, error) {
	var md types.FormsMetadata

	store, err := h.orderingSvc.GetStore().Get([]byte(evoting.FormsMetadataKey))
	if err != nil {
		return md, nil
	}

	// if there is no form created yet the metadata will be empty
	if len(store) == 0 {
		return types.FormsMetadata{}, nil
	}

	err = json.Unmarshal(store, &md)
	if err != nil {
		return md, xerrors.Errorf("failed to unmarshal FormMetadata: %v", err)
	}

	return md, nil
}
