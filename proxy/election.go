package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/proxy/txnmanager"
	ptypes "github.com/dedis/d-voting/proxy/types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
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

	// Compute the ID of the admin list id
	// We need it to filter the send list of form
	h := sha256.New()
	h.Write([]byte(evoting.AdminListId))
	adminListIDBuf := h.Sum(nil)
	adminListID := hex.EncodeToString(adminListIDBuf)

	return &form{
		logger:      logger,
		orderingSvc: srv,
		context:     ctx,
		formFac:     fac,
		adminFac:    types.AdminListFactory{},
		mngr:        txnManaxer,
		pool:        p,
		pk:          pk,
		adminListID: adminListID,
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
	adminFac    serde.Factory
	mngr        txnmanager.Manager
	pool        pool.Pool
	pk          kyber.Point
	adminListID string
}

// NewForm implements proxy.Proxy
func (form *form) NewForm(w http.ResponseWriter, r *http.Request) {
	var req ptypes.CreateFormRequest

	// get the signed request
	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	// get the request and verify the signature
	err = signed.GetAndVerify(form.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	createForm := types.CreateForm{
		Configuration: req.Configuration,
		UserID:        req.UserID,
	}

	// serialize the transaction
	data, err := createForm.Serialize(form.context)
	if err != nil {
		http.Error(w, "failed to marshal CreateFormTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, blockIdx, err := form.mngr.SubmitTxn(r.Context(), evoting.CmdCreateForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// hash the transaction
	hash := sha256.New()
	hash.Write(txnID)
	formID := hash.Sum(nil)

	// create it to get the  token
	transactionClientInfo, err := form.mngr.CreateTransactionResult(txnID, blockIdx, txnmanager.UnknownTransactionStatus)
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
func (form *form) NewFormVote(w http.ResponseWriter, r *http.Request) {
	var req ptypes.CastVoteRequest

	// get the signed request
	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	// get the request and verify the signature
	err = signed.GetAndVerify(form.pk, &req)
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

	elecMD, err := form.getFormsMetadata()
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
		FormID:  formID,
		VoterID: req.VoterID,
		Ballot:  ciphervote,
	}

	// serialize the vote
	data, err := castVote.Serialize(form.context)
	if err != nil {
		http.Error(w, "failed to marshal CastVoteTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := form.mngr.SubmitTxn(r.Context(), evoting.CmdCastVote, evoting.FormArg, data)
	if err != nil {
		form.logger.Err(err).Msg("failed to submit txn")
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's information
	err = form.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
	if err != nil {
		http.Error(w, "couldn't send transaction info: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// EditForm implements proxy.Proxy
func (form *form) EditForm(w http.ResponseWriter, r *http.Request) {
	var req ptypes.UpdateFormRequest

	// get the signed request
	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	// get the request and verify the signature
	err = signed.GetAndVerify(form.pk, &req)
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

	elecMD, err := form.getFormsMetadata()
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
		form.openForm(formID, req.UserID, w, r)
	case "close":
		form.closeForm(formID, req.UserID, w, r)
	case "combineShares":
		form.combineShares(formID, req.UserID, w, r)
	case "cancel":
		form.cancelForm(formID, req.UserID, w, r)
	default:
		BadRequestError(w, r, xerrors.Errorf("invalid action: %s", req.Action), nil)
		return
	}
}

// openForm allows opening a form, which sets the public key based on
// the DKG actor.
func (form *form) openForm(formID string, userID string, w http.ResponseWriter, r *http.Request) {
	openForm := types.OpenForm{
		FormID: formID,
		UserID: userID,
	}

	// serialize the transaction
	data, err := openForm.Serialize(form.context)
	if err != nil {
		http.Error(w, "failed to marshal OpenFormTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := form.mngr.SubmitTxn(r.Context(), evoting.CmdOpenForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's informations
	form.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
}

// closeForm closes a form.
func (form *form) closeForm(formIDHex string, userID string, w http.ResponseWriter, r *http.Request) {

	closeForm := types.CloseForm{
		FormID: formIDHex,
		UserID: userID,
	}

	// serialize the transaction
	data, err := closeForm.Serialize(form.context)
	if err != nil {
		http.Error(w, "failed to marshal CloseFormTransaction: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := form.mngr.SubmitTxn(r.Context(), evoting.CmdCloseForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's informations
	form.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)

}

// combineShares decrypts the shuffled ballots in a form.
func (form *form) combineShares(formIDHex string, userID string, w http.ResponseWriter, r *http.Request) {

	formFromStore, err := types.FormFromStore(form.context, form.formFac, formIDHex, form.orderingSvc.GetStore())
	if err != nil {
		http.Error(w, "failed to get form: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	if formFromStore.Status != types.PubSharesSubmitted {
		http.Error(w, "the submission of public shares must be over!",
			http.StatusUnauthorized)
		return
	}

	decryptBallots := types.CombineShares{
		FormID: formIDHex,
		UserID: userID,
	}

	// serialize the transaction
	data, err := decryptBallots.Serialize(form.context)
	if err != nil {
		http.Error(w, "failed to marshal decryptBallots: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := form.mngr.SubmitTxn(r.Context(), evoting.CmdCombineShares, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's informations
	form.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
}

// cancelForm cancels a form.
func (form *form) cancelForm(formIDHex string, userID string, w http.ResponseWriter, r *http.Request) {

	cancelForm := types.CancelForm{
		FormID: formIDHex,
		UserID: userID,
	}

	// serialize the transaction
	data, err := cancelForm.Serialize(form.context)
	if err != nil {
		http.Error(w, "failed to marshal CancelForm: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := form.mngr.SubmitTxn(r.Context(), evoting.CmdCancelForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's informations
	form.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
}

// Form implements proxy.Proxy. The request should not be signed because it
// is fetching public data.
func (form *form) Form(w http.ResponseWriter, r *http.Request) {
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
	formFromStore, err := types.FormFromStore(form.context, form.formFac, formID, form.orderingSvc.GetStore())
	if err != nil {
		http.Error(w, xerrors.Errorf("failed to get form: %v", err).Error(), http.StatusInternalServerError)
		return
	}

	var pubkeyBuf []byte

	// get the public key
	if formFromStore.Pubkey != nil {
		pubkeyBuf, err = formFromStore.Pubkey.MarshalBinary()
		if err != nil {
			http.Error(w, "failed to marshal pubkey: "+err.Error(),
				http.StatusInternalServerError)
			return
		}
	}

	roster := make([]string, 0, formFromStore.Roster.Len())

	iter := formFromStore.Roster.AddressIterator()
	for iter.HasNext() {
		roster = append(roster, iter.GetNext().String())
	}

	suff, err := formFromStore.Suffragia(form.context, form.orderingSvc.GetStore())
	if err != nil {
		http.Error(w, "couldn't get ballots: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	response := ptypes.GetFormResponse{
		FormID:          string(formFromStore.FormID),
		Configuration:   formFromStore.Configuration,
		Status:          uint16(formFromStore.Status),
		Pubkey:          hex.EncodeToString(pubkeyBuf),
		Result:          formFromStore.DecryptedBallots,
		Roster:          roster,
		ChunksPerBallot: formFromStore.ChunksPerBallot(),
		BallotSize:      formFromStore.BallotSize,
		Voters:          suff.VoterIDs,
	}

	txnmanager.SendResponse(w, response)

}

// Forms implements proxy.Proxy. The request should not be signed because it
// is fecthing public data.
func (form *form) Forms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	elecMD, err := form.getFormsMetadata()
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to get form metadata: %v", err), nil)
		return
	}

	allFormsInfo := make([]ptypes.LightForm, len(elecMD.FormsIDs))

	// get the forms
	for i, id := range elecMD.FormsIDs {
		if id != form.adminListID {
			form, err := types.FormFromStore(form.context, form.formFac, id, form.orderingSvc.GetStore())
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
	}

	response := ptypes.GetFormsResponse{Forms: allFormsInfo}

	txnmanager.SendResponse(w, response)

}

// DeleteForm implements proxy.Proxy
func (form *form) DeleteForm(w http.ResponseWriter, r *http.Request) {
	var req ptypes.UpdateFormRequest

	vars := mux.Vars(r)

	// check if the formID is valid
	if vars == nil || vars["formID"] == "" {
		http.Error(w, fmt.Sprintf("formID not found: %v", vars), http.StatusInternalServerError)
		return
	}

	formID := vars["formID"]

	elecMD, err := form.getFormsMetadata()
	if err != nil {
		http.Error(w, "failed to get form metadata", http.StatusNotFound)
		return
	}

	// check if the form exists
	if elecMD.FormsIDs.Contains(formID) < 0 {
		http.Error(w, "the form does not exist", http.StatusNotFound)
		return
	}

	// TODO double check
	// get the signed request
	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.GetAndVerify(form.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	deleteForm := types.DeleteForm{
		FormID: formID,
		UserID: req.UserID,
	}

	data, err := deleteForm.Serialize(form.context)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to marshal DeleteForm: %v", err), nil)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := form.mngr.SubmitTxn(r.Context(), evoting.CmdDeleteForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's information
	form.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
}

// TODO CHECK CAUSE NEW
// POST /addtoadminlist
func (form *form) AddAdmin(w http.ResponseWriter, r *http.Request) {
	var req ptypes.AdminRequest

	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.GetAndVerify(form.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	targetUserID := req.TargetUserID
	performingUserID := req.PerformingUserID

	addAdmin := types.AddAdmin{
		TargetUserID:     targetUserID,
		PerformingUserID: performingUserID,
	}

	data, err := addAdmin.Serialize(form.context)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to marshal AddAdmin: %v", err), nil)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := form.mngr.SubmitTxn(r.Context(), evoting.CmdAddAdmin, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's information
	form.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)
}

// POST /removetoadminlist
func (form *form) RemoveAdmin(w http.ResponseWriter, r *http.Request) {
	var req ptypes.AdminRequest

	signed, err := ptypes.NewSignedRequest(r.Body)
	if err != nil {
		InternalError(w, r, newSignedErr(err), nil)
		return
	}

	err = signed.GetAndVerify(form.pk, &req)
	if err != nil {
		InternalError(w, r, getSignedErr(err), nil)
		return
	}

	targetUserID := req.TargetUserID
	performingUserID := req.PerformingUserID

	removeAdmin := types.RemoveAdmin{
		TargetUserID:     targetUserID,
		PerformingUserID: performingUserID,
	}

	data, err := removeAdmin.Serialize(form.context)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to marshal RemoveAdmin: %v", err), nil)
		return
	}

	// create the transaction and add it to the pool
	txnID, lastBlock, err := form.mngr.SubmitTxn(r.Context(), evoting.CmdRemoveAdmin, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// send the transaction's information
	form.mngr.SendTransactionInfo(w, txnID, lastBlock, txnmanager.UnknownTransactionStatus)

}

// GET /adminlist
func (form *form) AdminList(w http.ResponseWriter, r *http.Request) {
	println("un test marche parfois\n\n mais parfois pas \n\n")
	adminList, err := types.AdminListFromStore(form.context, form.adminFac, form.orderingSvc.GetStore(), evoting.AdminListId)
	if err != nil {
		InternalError(w, r, xerrors.Errorf("failed to get form: %v", err), nil)
		return
	}

	myAdminList := "{"
	for id := range adminList.AdminList {
		myAdminList += strconv.Itoa(adminList.AdminList[id]) + ", "
	}
	myAdminList += "}"

	println(myAdminList)

	txnmanager.SendResponse(w, myAdminList)
}

// POST /forms/{formID}/addowner
func (form *form) AddOwnerToForm(http.ResponseWriter, *http.Request) {}

// POST /forms/{formID}/removeowner
func (form *form) RemoveOwnerToForm(http.ResponseWriter, *http.Request) {}

// POST /forms/{formID}/addvoter
func (form *form) AddVoterToForm(http.ResponseWriter, *http.Request) {}

// POST /forms/{formID}/removevoter
func (form *form) RemoveVoterToForm(http.ResponseWriter, *http.Request) {}

// ===== HELPER =====

func (form *form) getFormsMetadata() (types.FormsMetadata, error) {
	var md types.FormsMetadata

	store, err := form.orderingSvc.GetStore().Get([]byte(evoting.FormsMetadataKey))
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
