package proxy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	ptypes "github.com/dedis/d-voting/proxy/types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
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
func NewForm(srv ordering.Service, mngr txn.Manager, p pool.Pool,
	ctx serde.Context, fac serde.Factory, pk kyber.Point) Form {

	logger := dela.Logger.With().Timestamp().Str("role", "evoting-proxy").Logger()

	return &form{
		logger:      logger,
		orderingSvc: srv,
		context:     ctx,
		formFac: fac,
		mngr:        mngr,
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
	formFac serde.Factory
	mngr        txn.Manager
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
	txID, err := h.submitAndWaitForTxn(r.Context(), evoting.CmdCreateForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// hash the transaction
	hash := sha256.New()
	hash.Write(txID)
	formID := hash.Sum(nil)

	// return the formID
	response := ptypes.CreateFormResponse{
		FormID: hex.EncodeToString(formID),
	}

	// encode the response and write it in the ResponseWriter
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
}

// NewFormVote implements proxy.Proxy
func (h *form) NewFormVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// check if the formID is valid
	if vars == nil || vars["formID"] == "" {
		http.Error(w, fmt.Sprintf("formID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	formID := vars["formID"]

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

	// unmarshalling the ballots and keys from the request.
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
		UserID:     req.UserID,
		Ballot:     ciphervote,
	}

	// serialize the vote
	data, err := castVote.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal CastVoteTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCastVote, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// EditForm implements proxy.Proxy
func (h *form) EditForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	//check if the formID is valid
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
func (h *form) openForm(elecID string, w http.ResponseWriter, r *http.Request) {
	openForm := types.OpenForm{
		FormID: elecID,
	}

	// serialize the transaction
	data, err := openForm.Serialize(h.context)
	if err != nil {
		http.Error(w, "failed to marshal OpenFormTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdOpenForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
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
	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCloseForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// combineShares decrypts the shuffled ballots in a form.
func (h *form) combineShares(formIDHex string, w http.ResponseWriter, r *http.Request) {

	form, err := getForm(h.context, h.formFac, formIDHex, h.orderingSvc)
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
	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCombineShares, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
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
	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdCancelForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
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
	form, err := getForm(h.context, h.formFac, formID, h.orderingSvc)
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

	response := ptypes.GetFormResponse{
		FormID:      string(form.FormID),
		Configuration:   form.Configuration,
		Status:          uint16(form.Status),
		Pubkey:          hex.EncodeToString(pubkeyBuf),
		Result:          form.DecryptedBallots,
		Roster:          roster,
		ChunksPerBallot: form.ChunksPerBallot(),
		BallotSize:      form.BallotSize,
		Voters:          form.Suffragia.UserIDs,
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
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
		form, err := getForm(h.context, h.formFac, id, h.orderingSvc)
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
			Title:      form.Configuration.MainTitle,
			Status:     uint16(form.Status),
			Pubkey:     hex.EncodeToString(pubkeyBuf),
		}

		allFormsInfo[i] = info
	}

	response := ptypes.GetFormsResponse{Forms: allFormsInfo}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to write response: %v", err), nil)
		return
	}
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

	sig, err := hex.DecodeString(auth)
	if err != nil {
		BadRequestError(w, r, xerrors.Errorf("failed to decode auth: %v", err), nil)
		return
	}

	// check if the signature is valid
	err = schnorr.Verify(suite, h.pk, []byte(formID), sig)
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
	_, err = h.submitAndWaitForTxn(r.Context(), evoting.CmdDeleteForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// waitForTxnID blocks until `ID` is included or `events` is closed.
func (h *form) waitForTxnID(events <-chan ordering.Event, ID []byte) error {
	for event := range events {
		for _, res := range event.Transactions {
			if !bytes.Equal(res.GetTransaction().GetID(), ID) {
				continue
			}

			ok, msg := res.GetStatus()
			if !ok {
				return xerrors.Errorf("transaction %x denied : %s", ID, msg)
			}

			return nil
		}
	}

	return xerrors.New("transaction not found")
}

func (h *form) getFormsMetadata() (types.FormsMetadata, error) {
	var md types.FormsMetadata

	proof, err := h.orderingSvc.GetProof([]byte(evoting.FormsMetadataKey))
	if err != nil {
		// if the proof doesn't exist we assume there is no metadata, thus no
		// forms has been created so far.
		return md, nil
	}

	// if there is not form created yet the metadata will be empty
	if len(proof.GetValue()) == 0 {
		return types.FormsMetadata{}, nil
	}

	err = json.Unmarshal(proof.GetValue(), &md)
	if err != nil {
		return md, xerrors.Errorf("failed to unmarshal FormMetadata: %v", err)
	}

	return md, nil
}

// getForm gets the form from the snap. Returns the form ID NOT hex
// encoded.
func getForm(ctx serde.Context, formFac serde.Factory, formIDHex string,
	srv ordering.Service) (types.Form, error) {

	var form types.Form

	formID, err := hex.DecodeString(formIDHex)
	if err != nil {
		return form, xerrors.Errorf("failed to decode formIDHex: %v", err)
	}

	proof, err := srv.GetProof(formID)
	if err != nil {
		return form, xerrors.Errorf("failed to get proof: %v", err)
	}

	formBuff := proof.GetValue()
	if len(formBuff) == 0 {
		return form, xerrors.Errorf("form does not exist")
	}

	message, err := formFac.Deserialize(ctx, formBuff)
	if err != nil {
		return form, xerrors.Errorf("failed to deserialize Form: %v", err)
	}

	form, ok := message.(types.Form)
	if !ok {
		return form, xerrors.Errorf("wrong message type: %T", message)
	}

	return form, nil
}

// submitAndWaitForTxn submits a transaction and waits for it to be included.
// Returns the transaction ID.
func (h *form) submitAndWaitForTxn(ctx context.Context, cmd evoting.Command,
	cmdArg string, payload []byte) ([]byte, error) {

	h.Lock()
	defer h.Unlock()

	err := h.mngr.Sync()
	if err != nil {
		return nil, xerrors.Errorf("failed to sync manager: %v", err)
	}

	tx, err := createTransaction(h.mngr, cmd, cmdArg, payload)
	if err != nil {
		return nil, xerrors.Errorf("failed to create transaction: %v", err)
	}

	watchCtx, cancel := context.WithTimeout(ctx, inclusionTimeout)
	defer cancel()

	events := h.orderingSvc.Watch(watchCtx)

	err = h.pool.Add(tx)
	if err != nil {
		return nil, xerrors.Errorf("failed to add transaction to the pool: %v", err)
	}

	err = h.waitForTxnID(events, tx.GetID())
	if err != nil {
		return nil, xerrors.Errorf("failed to wait for transaction: %v", err)
	}

	return tx.GetID(), nil
}

// createTransaction creates a transaction with the given command and payload.
func createTransaction(manager txn.Manager, commandType evoting.Command,
	commandArg string, buf []byte) (txn.Transaction, error) {

	args := []txn.Arg{
		{
			Key:   native.ContractArg,
			Value: []byte(evoting.ContractName),
		},
		{
			Key:   evoting.CmdArg,
			Value: []byte(commandType),
		},
		{
			Key:   commandArg,
			Value: buf,
		},
	}

	tx, err := manager.Make(args...)
	if err != nil {
		return nil, xerrors.Errorf("failed to create transaction from manager: %v", err)
	}

	return tx, nil
}
