package proxy

import (
	"bytes"
	"context"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	//"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	ptypes "github.com/dedis/d-voting/proxy/types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"golang.org/x/xerrors"
)

const (
	maxTimeTransactionCheck = 10 * time.Minute
)

func newSignedErr(err error) error {
	return xerrors.Errorf("failed to created signed request: %v", err)
}

func getSignedErr(err error) error {
	return xerrors.Errorf("failed to get and verify signed request: %v", err)
}

// NewForm returns a new initialized form proxy
func NewForm(srv ordering.Service, mngr txn.Manager, p pool.Pool,
	ctx serde.Context, fac serde.Factory, pk kyber.Point, blocks blockstore.BlockStore, signer crypto.Signer) Form {

	logger := dela.Logger.With().Timestamp().Str("role", "evoting-proxy").Logger()

	return &form{
		logger:      logger,
		orderingSvc: srv,
		context:     ctx,
		formFac:     fac,
		mngr:        mngr,
		pool:        p,
		pk:          pk,
		blocks:      blocks,
		signer:      signer,
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
	mngr        txn.Manager
	pool        pool.Pool
	pk          kyber.Point
	blocks      blockstore.BlockStore
	signer      crypto.Signer
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
	txnID, blockIdx, err := h.submitTxn(r.Context(), evoting.CmdCreateForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: wait for the transaction to be included in a block??

	// hash the transaction
	hash := sha256.New()
	hash.Write(txnID)
	formID := hash.Sum(nil)

	transactionInfoToSend, err := h.CreateTransactionInfoToSend(txnID, blockIdx, ptypes.UnknownTransactionStatus)
	if err != nil {
		http.Error(w, "failed to create transaction info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// return the formID
	response := ptypes.CreateFormResponse{
		FormID: hex.EncodeToString(formID),
		Token:  transactionInfoToSend.Token,
	}

	// sign the response
	sendResponse(w, response)
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

	h.logger.Info().Msg(fmt.Sprintf("NewFormVote: %v", req))

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

	// encrypt the vote
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
	txnID, lastBlock, err := h.submitTxn(r.Context(), evoting.CmdCastVote, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//send the transaction
	h.sendTransactionInfo(w, txnID, lastBlock, ptypes.UnknownTransactionStatus)

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
	txnID, lastBlock, err := h.submitTxn(r.Context(), evoting.CmdOpenForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//send the transaction
	h.sendTransactionInfo(w, txnID, lastBlock, ptypes.UnknownTransactionStatus)
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
	txnID, lastBlock, err := h.submitTxn(r.Context(), evoting.CmdCloseForm, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//send the transaction
	h.sendTransactionInfo(w, txnID, lastBlock, ptypes.UnknownTransactionStatus)

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
	txnID, lastBlock, err := h.submitTxn(r.Context(), evoting.CmdCombineShares, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//send the transaction
	h.sendTransactionInfo(w, txnID, lastBlock, ptypes.UnknownTransactionStatus)
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
	txnID, lastBlock, err := h.submitTxn(r.Context(), evoting.CmdCombineShares, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//send the transaction
	h.sendTransactionInfo(w, txnID, lastBlock, ptypes.UnknownTransactionStatus)
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
		FormID:          string(form.FormID),
		Configuration:   form.Configuration,
		Status:          uint16(form.Status),
		Pubkey:          hex.EncodeToString(pubkeyBuf),
		Result:          form.DecryptedBallots,
		Roster:          roster,
		ChunksPerBallot: form.ChunksPerBallot(),
		BallotSize:      form.BallotSize,
		Voters:          form.Suffragia.UserIDs,
	}

	sendResponse(w, response)

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
			Title:  form.Configuration.MainTitle,
			Status: uint16(form.Status),
			Pubkey: hex.EncodeToString(pubkeyBuf),
		}

		allFormsInfo[i] = info
	}

	response := ptypes.GetFormsResponse{Forms: allFormsInfo}

	sendResponse(w, response)

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
	txnID, lastBlock, err := h.submitTxn(r.Context(), evoting.CmdCombineShares, evoting.FormArg, data)
	if err != nil {
		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//send the transaction
	h.sendTransactionInfo(w, txnID, lastBlock, ptypes.UnknownTransactionStatus)
}

// IsTxnIncluded
func (h *form) IsTxnIncluded(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// check if the formID is valid
	if vars == nil || vars["token"] == "" {
		http.Error(w, fmt.Sprintf("token not found: %v", vars), http.StatusInternalServerError)
		return
	}

	token := vars["token"]

	marshall, err := b64.URLEncoding.DecodeString(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode token: %v", err), http.StatusInternalServerError)
		return
	}

	var content ptypes.TransactionInfo
	json.Unmarshal(marshall, &content)

	//h.logger.Info().Msg(fmt.Sprintf("Transaction infos: %+v", content))

	// get the status of the transaction as byte
	if content.Status != ptypes.UnknownTransactionStatus {
		http.Error(w, "the transaction status is known", http.StatusBadRequest)
		return
	}

	// get the signature as a crypto.Signature
	signature, err := h.signer.GetSignatureFactory().SignatureOf(h.context, content.Signature)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get Signature: %v", err), http.StatusInternalServerError)
		return
	}

	// check if the hash is valid
	if !h.checkHash(content.Status, content.TransactionID, content.LastBlockIdx, content.Time, content.Hash) {
		http.Error(w, "invalid hash", http.StatusInternalServerError)
		return
	}

	// check if the signature is valid
	if !h.checkSignature(content.Hash, signature) {
		http.Error(w, "invalid signature", http.StatusInternalServerError)
		return
	}

	// check if if was submited not to long ago
	if time.Now().Unix()-content.Time > int64(maxTimeTransactionCheck) {
		http.Error(w, "the transaction is too old", http.StatusInternalServerError)
		return
	}

	if time.Now().Unix()-content.Time < 0 {
		http.Error(w, "the transaction is from the future", http.StatusInternalServerError)
		return
	}

	// check if the transaction is included in the blockchain
	newStatus, idx := h.checkTxnIncluded(content.TransactionID, content.LastBlockIdx)

	err = h.sendTransactionInfo(w, content.TransactionID, idx, newStatus)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to send transaction info: %v", err), http.StatusInternalServerError)
		return
	}

}

// checkHash checks if the hash is valid
func (h *form) checkHash(status ptypes.TransactionStatus, transactionID []byte, LastBlockIdx uint64, Time int64, Hash []byte) bool {
	// create the hash
	hash := sha256.New()
	hash.Write([]byte{byte(status)})
	hash.Write(transactionID)
	hash.Write([]byte(strconv.FormatUint(LastBlockIdx, 10)))
	hash.Write([]byte(strconv.FormatInt(Time, 10)))

	// check if the hash is valid
	return bytes.Equal(hash.Sum(nil), Hash)
}

// checkSignature checks if the signature is valid
func (h *form) checkSignature(Hash []byte, Signature crypto.Signature) bool {
	// check if the signature is valid

	return h.signer.GetPublicKey().Verify(Hash, Signature) == nil
}

// checkTxnIncluded checks if the transaction is included in the blockchain
func (h *form) checkTxnIncluded(transactionID []byte, lastBlockIdx uint64) (ptypes.TransactionStatus, uint64) {
	// first get the block
	idx := lastBlockIdx

	for {

		blockLink, err := h.blocks.GetByIndex(idx)
		// if we reached the end of the blockchain
		if err != nil {
			return ptypes.UnknownTransactionStatus, idx - 1
		}

		transactions := blockLink.GetBlock().GetTransactions()
		for _, txn := range transactions {
			if bytes.Equal(txn.GetID(), transactionID) {
				return ptypes.IncludedTransaction, blockLink.GetBlock().GetIndex()
			}

		}

		idx++
	}
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

// submitTxn submits a transaction
// Returns the transaction ID.
func (h *form) submitTxn(ctx context.Context, cmd evoting.Command,
	cmdArg string, payload []byte) ([]byte, uint64, error) {

	h.Lock()
	defer h.Unlock()

	err := h.mngr.Sync()
	if err != nil {
		return nil, 0, xerrors.Errorf("failed to sync manager: %v", err)
	}

	tx, err := createTransaction(h.mngr, cmd, cmdArg, payload)
	if err != nil {
		return nil, 0, xerrors.Errorf("failed to create transaction: %v", err)
	}

	//watchCtx, cancel := context.WithTimeout(ctx, inclusionTimeout) //plus besoin de timeout
	//defer cancel()

	// events := h.orderingSvc.Watch(watchCtx) il faudra implementer ca lorsque l'on devra appeler checkTxnIncluded

	lastBlock, err := h.blocks.Last()
	if err != nil {
		return nil, 0, xerrors.Errorf("failed to get last block: %v", err)
	}
	lastBlockIdx := lastBlock.GetBlock().GetIndex()

	err = h.pool.Add(tx) //dans l'idee, on ajoute la transaction au pool et on sauvegarde le bloc qui debute,
	// ensuite on dit au frontend que ca a bien ete added en lui transmettant le txnID
	// le frontend peut alors lui meme verifier si la transaction est bien incluse dans le bloc
	// en passant par le proxy et sa fonction checkTxnIncluded

	if err != nil {
		return nil, 0, xerrors.Errorf("failed to add transaction to the pool: %v", err)
	}
	/*
		err = h.waitForTxnID(events, tx.GetID())
		if err != nil {
			return nil, xerrors.Errorf("failed to wait for transaction: %v", err)
		}
	*/

	return tx.GetID(), lastBlockIdx, nil
}

// A function that checks if a transaction is included in a block
/*func (h *form) checkTxnIncluded(events <-chan ordering.Event, ID []byte) (bool, error) {
	for event := range events {
		for _, res := range event.Transactions {
			if !bytes.Equal(res.GetTransaction().GetID(), ID) {
				continue
			}

			ok, msg := res.GetStatus()
			if !ok {
				return false, xerrors.Errorf("transaction %x denied : %s", ID, msg)
			}

			return true, nil
		}
	}

	return false, nil
}*/

func (h *form) sendTransactionInfo(w http.ResponseWriter, txnID []byte, lastBlockIdx uint64, status ptypes.TransactionStatus) error {

	response, err := h.CreateTransactionInfoToSend(txnID, lastBlockIdx, status)
	if err != nil {
		return xerrors.Errorf("failed to create transaction info: %v", err)
	}
	return sendResponse(w, response)

}

func (h *form) CreateTransactionInfoToSend(txnID []byte, lastBlockIdx uint64, status ptypes.TransactionStatus) (ptypes.TransactionInfoToSend, error) {

	time := time.Now().Unix()
	hash := sha256.New()

	// write status which is a byte to the hash as a []byte
	hash.Write([]byte{byte(status)})
	hash.Write(txnID)
	hash.Write([]byte(strconv.FormatUint(lastBlockIdx, 10)))
	hash.Write([]byte(strconv.FormatInt(time, 10)))

	finalHash := hash.Sum(nil)

	signature, err := h.signer.Sign(finalHash)

	if err != nil {
		return ptypes.TransactionInfoToSend{}, xerrors.Errorf("failed to sign transaction info: %v", err)
	}
	//convert signature to []byte
	signatureBin, err := signature.Serialize(h.context)
	if err != nil {
		return ptypes.TransactionInfoToSend{}, xerrors.Errorf("failed to marshal signature: %v", err)
	}

	infos := ptypes.TransactionInfo{
		Status:        status,
		TransactionID: txnID,
		LastBlockIdx:  lastBlockIdx,
		Time:          time,
		Hash:          finalHash,
		Signature:     signatureBin,
	}
	marshal, err := json.Marshal(infos)
	if err != nil {
		return ptypes.TransactionInfoToSend{}, xerrors.Errorf("failed to marshal transaction info: %v", err)
	}

	token := b64.URLEncoding.EncodeToString(marshal)

	response := ptypes.TransactionInfoToSend{
		Status: status,
		Token:  token,
	}
	h.logger.Info().Msg(fmt.Sprintf("Transaction info: %v", response))
	return response, nil
}

func sendResponse(w http.ResponseWriter, response any) error {

	w.Header().Set("Content-Type", "application/json")

	// Status et token

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to write in ResponseWriter: "+err.Error(),
			http.StatusInternalServerError)
		return nil
	}

	return nil
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
