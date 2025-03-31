package txnmanager

import (
	"bytes"
	"context"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.dedis.ch/d-voting/contracts/evoting"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

const (
	maxTimeTransactionCheck = 10 * time.Minute
)

// NewTransactionManager returns a new initialized transaction manager
func NewTransactionManager(mngr txn.Manager, p pool.Pool,
	ctx serde.Context, pk kyber.Point, blocks blockstore.BlockStore, signer crypto.Signer) Manager {

	logger := dela.Logger.With().Timestamp().Str("role", "proxy-txmanager").Logger()

	return &manager{
		logger:  logger,
		context: ctx,
		mngr:    mngr,
		pool:    p,
		pk:      pk,
		blocks:  blocks,
		signer:  signer,
	}
}

// manager defines the HTTP handlers to manage transactions
//
// - implements proxy.Transaction
type manager struct {
	sync.Mutex

	logger  zerolog.Logger
	context serde.Context
	mngr    txn.Manager
	pool    pool.Pool
	pk      kyber.Point
	blocks  blockstore.BlockStore
	signer  crypto.Signer
}

// StatusHandlerGet checks if the transaction is included in the blockchain
func (h *manager) StatusHandlerGet(w http.ResponseWriter, r *http.Request) {
	// get the token from the url
	vars := mux.Vars(r)

	// check if the token is valid
	if vars == nil || vars["token"] == "" {
		http.Error(w, fmt.Sprintf("token not found: %v", vars), http.StatusInternalServerError)
		return
	}

	token := vars["token"]

	// decode the token
	marshall, err := b64.URLEncoding.DecodeString(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode token: %v", err), http.StatusInternalServerError)
		return
	}

	// unmarshall the token to get the json with all the informations
	var content transactionInternalInfo
	err = json.Unmarshal(marshall, &content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshall token: %v", err), http.StatusInternalServerError)
		return
	}

	err = content.validate(h)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid content: %v", err), http.StatusBadRequest)
		return
	}

	// check if if was submited not to long ago
	if time.Now().Unix()-content.Time > int64(maxTimeTransactionCheck) {
		// if it was submited to long ago, we reject the transaction
		err = h.SendTransactionInfo(w, content.TransactionID, 0, RejectedTransaction)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to send transaction info: %v", err), http.StatusInternalServerError)
			return
		}
		return
	}

	// check if the transaction time stamp is possible
	if time.Now().Unix()-content.Time < 0 {
		http.Error(w, "the transaction is from the future", http.StatusInternalServerError)
		return
	}

	// check if the transaction is included in the blockchain
	newStatus, idx := h.checkTxnIncluded(content.TransactionID, content.LastBlockIdx)

	// send the transaction info
	err = h.SendTransactionInfo(w, content.TransactionID, idx, newStatus)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to send transaction info: %v", err), http.StatusInternalServerError)
		return
	}

}

// validate checks if the transaction is valid
func (content transactionInternalInfo) validate(h *manager) error {
	// check if the transaction status is unknown
	// if it is not unknown, it means that the transaction was already checked
	if content.Status != UnknownTransactionStatus {
		return xerrors.Errorf("the transaction status is known")
	}

	// get the signature as a crypto.Signature
	signature, err := h.signer.GetSignatureFactory().SignatureOf(h.context, content.Signature)
	if err != nil {
		return xerrors.Errorf(fmt.Sprintf("failed to get Signature: %v", err))
	}

	// check if the hash is valid
	if !h.checkHash(content.Status, content.TransactionID, content.LastBlockIdx, content.Time, content.Hash) {
		return xerrors.Errorf("invalid hash")
	}

	// check if the signature is valid
	if !h.checkSignature(content.Hash, signature) {
		return xerrors.Errorf("invalid signature")
	}

	return nil
}

// checkHash checks if the hash is valid
func (h *manager) checkHash(status TransactionStatus, transactionID []byte, LastBlockIdx uint64, Time int64, Hash []byte) bool {
	// create the hash
	myHash := hashInfos(status, transactionID, LastBlockIdx, Time)

	// check if the hash is valid
	return bytes.Equal(myHash, Hash)
}

// checkSignature checks if the signature is valid
func (h *manager) checkSignature(Hash []byte, Signature crypto.Signature) bool {
	return h.signer.GetPublicKey().Verify(Hash, Signature) == nil
}

// checkTxnIncluded checks if the transaction is included in the blockchain
func (h *manager) checkTxnIncluded(transactionID []byte, lastBlockIdx uint64) (TransactionStatus, uint64) {
	// we start at the last block index
	// which is the index of the last block that was checked
	// or the last block before the transaction was submited
	idx := lastBlockIdx

	for {
		// first get the block
		blockLink, err := h.blocks.GetByIndex(idx)

		// if we reached the end of the blockchain
		if err != nil {
			return UnknownTransactionStatus, idx - 1
		}

		// check if the transaction is in the block
		transactions := blockLink.GetBlock().GetTransactions()
		for _, txn := range transactions {
			if bytes.Equal(txn.GetID(), transactionID) {
				return IncludedTransaction, blockLink.GetBlock().GetIndex()
			}

		}

		idx++
	}
}

// SubmitTxn submits a transaction
// Returns the transaction ID.
func (h *manager) SubmitTxn(ctx context.Context, cmd evoting.Command,
	cmdArg string, payload []byte) ([]byte, uint64, error) {

	h.Lock()
	defer h.Unlock()

	tx, err := createTransaction(h.mngr, cmd, cmdArg, payload)
	if err != nil {
		return nil, 0, xerrors.Errorf("failed to create transaction: %v", err)
	}

	// get the last block
	lastBlock, err := h.blocks.Last()
	if err != nil {
		return nil, 0, xerrors.Errorf("failed to get last block: %v", err)
	}
	lastBlockIdx := lastBlock.GetBlock().GetIndex()

	err = h.pool.Add(tx)
	if err != nil {
		return nil, 0, xerrors.Errorf("failed to add transaction to the pool: %v", err)
	}

	return tx.GetID(), lastBlockIdx, nil
}

func (h *manager) SendTransactionInfo(w http.ResponseWriter, txnID []byte, lastBlockIdx uint64, status TransactionStatus) error {

	response, err := h.CreateTransactionResult(txnID, lastBlockIdx, status)
	if err != nil {
		return xerrors.Errorf("failed to create transaction info: %v", err)
	}
	return SendResponse(w, response)

}

func (h *manager) CreateTransactionResult(txnID []byte, lastBlockIdx uint64, status TransactionStatus) (TransactionClientInfo, error) {

	time := time.Now().Unix()
	hash := hashInfos(status, txnID, lastBlockIdx, time)
	signature, err := h.signer.Sign(hash)

	if err != nil {
		return TransactionClientInfo{}, xerrors.Errorf("failed to sign transaction info: %v", err)
	}

	// convert signature to []byte
	signatureBin, err := signature.Serialize(h.context)
	if err != nil {
		return TransactionClientInfo{}, xerrors.Errorf("failed to marshal signature: %v", err)
	}

	infos := transactionInternalInfo{
		Status:        status,
		TransactionID: txnID,
		LastBlockIdx:  lastBlockIdx,
		Time:          time,
		Hash:          hash,
		Signature:     signatureBin,
	}

	marshal, err := json.Marshal(infos)
	if err != nil {
		return TransactionClientInfo{}, xerrors.Errorf("failed to marshal transaction info: %v", err)
	}

	// encode the transaction info so that the client just has to send it back
	token := b64.URLEncoding.EncodeToString(marshal)

	response := TransactionClientInfo{
		Status: status,
		Token:  token,
	}
	return response, nil
}

// SendResponse sends a response to the client.
func SendResponse(w http.ResponseWriter, response any) error {

	w.Header().Set("Content-Type", "application/json")

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

func hashInfos(status TransactionStatus, txnID []byte, lastBlockIdx uint64, time int64) []byte {
	hash := sha256.New()

	// create the hash
	hash.Write([]byte{byte(status)})
	hash.Write(txnID)
	hash.Write([]byte(strconv.FormatUint(lastBlockIdx, 10)))
	hash.Write([]byte(strconv.FormatInt(time, 10)))

	return hash.Sum(nil)
}
