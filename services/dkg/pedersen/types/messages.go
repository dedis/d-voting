package types

import (
	"github.com/c4dt/dela/mino"
	"github.com/c4dt/dela/serde"
	"github.com/c4dt/dela/serde/registry"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

var msgFormats = registry.NewSimpleRegistry()

// RegisterMessageFormat register the engine for the provided format.
func RegisterMessageFormat(c serde.Format, f serde.FormatEngine) {
	msgFormats.Register(c, f)
}

// Start is the message the initiator of the DKG protocol should send to all the
// nodes.
//
// - implements serde.Message
type Start struct {
	// the full list of addresses that will participate in the DKG
	addresses []mino.Address
	// the corresponding kyber.Point pub keys of the addresses
	pubkeys []kyber.Point
}

// NewStart creates a new start message.
func NewStart(addrs []mino.Address, pubkeys []kyber.Point) Start {
	return Start{
		addresses: addrs,
		pubkeys:   pubkeys,
	}
}

// GetAddresses returns the list of addresses.
func (s Start) GetAddresses() []mino.Address {
	return append([]mino.Address{}, s.addresses...)
}

// GetPublicKeys returns the list of public keys.
func (s Start) GetPublicKeys() []kyber.Point {
	return append([]kyber.Point{}, s.pubkeys...)
}

// Serialize implements serde.Message. It looks up the format and returns the
// serialized data for the start message.
func (s Start) Serialize(ctx serde.Context) ([]byte, error) {
	format := msgFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, s)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode message: %v", err)
	}

	return data, nil
}

// EncryptedDeal contains the different parameters and data of an encrypted
// deal.
type EncryptedDeal struct {
	dhkey     []byte
	signature []byte
	nonce     []byte
	cipher    []byte
}

// NewEncryptedDeal creates a new encrypted deal message.
func NewEncryptedDeal(dhkey, sig, nonce, cipher []byte) EncryptedDeal {
	return EncryptedDeal{
		dhkey:     dhkey,
		signature: sig,
		nonce:     nonce,
		cipher:    cipher,
	}
}

// GetDHKey returns the Diffie-Helmann key in bytes.
func (d EncryptedDeal) GetDHKey() []byte {
	return append([]byte{}, d.dhkey...)
}

// GetSignature returns the signatures in bytes.
func (d EncryptedDeal) GetSignature() []byte {
	return append([]byte{}, d.signature...)
}

// GetNonce returns the nonce in bytes.
func (d EncryptedDeal) GetNonce() []byte {
	return append([]byte{}, d.nonce...)
}

// GetCipher returns the cipher in bytes.
func (d EncryptedDeal) GetCipher() []byte {
	return append([]byte{}, d.cipher...)
}

// Deal matches the attributes defined in kyber dkg.Deal.
//
// - implements serde.Message
type Deal struct {
	index     uint32
	signature []byte

	encryptedDeal EncryptedDeal
}

// NewDeal creates a new deal.
func NewDeal(index uint32, sig []byte, e EncryptedDeal) Deal {
	return Deal{
		index:         index,
		signature:     sig,
		encryptedDeal: e,
	}
}

// GetIndex returns the index.
func (d Deal) GetIndex() uint32 {
	return d.index
}

// GetSignature returns the signature in bytes.
func (d Deal) GetSignature() []byte {
	return append([]byte{}, d.signature...)
}

// GetEncryptedDeal returns the encrypted deal.
func (d Deal) GetEncryptedDeal() EncryptedDeal {
	return d.encryptedDeal
}

// Serialize implements serde.Message.
func (d Deal) Serialize(ctx serde.Context) ([]byte, error) {
	format := msgFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, d)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode deal: %v", err)
	}

	return data, nil
}

// DealerResponse is a response of a single dealer.
type DealerResponse struct {
	sessionID []byte
	// Indexes of the verifier issuing this Response from the new set of
	// nodes.
	index     uint32
	status    bool
	signature []byte
}

// NewDealerResponse creates a new dealer response.
func NewDealerResponse(index uint32, status bool, sessionID, sig []byte) DealerResponse {
	return DealerResponse{
		sessionID: sessionID,
		index:     index,
		status:    status,
		signature: sig,
	}
}

// GetSessionID returns the session ID in bytes.
func (dresp DealerResponse) GetSessionID() []byte {
	return append([]byte{}, dresp.sessionID...)
}

// GetIndex returns the index.
func (dresp DealerResponse) GetIndex() uint32 {
	return dresp.index
}

// GetStatus returns the status.
func (dresp DealerResponse) GetStatus() bool {
	return dresp.status
}

// GetSignature returns the signature in bytes.
func (dresp DealerResponse) GetSignature() []byte {
	return append([]byte{}, dresp.signature...)
}

// Response matches the attributes defined in kyber pedersen.Response.
//
// - implements serde.Message
type Response struct {
	// Indexes of the Dealer this response is for.
	index    uint32
	response DealerResponse
}

// NewResponse creates a new response.
func NewResponse(index uint32, r DealerResponse) Response {
	return Response{
		index:    index,
		response: r,
	}
}

// GetIndex returns the index.
func (r Response) GetIndex() uint32 {
	return r.index
}

// GetResponse returns the dealer response.
func (r Response) GetResponse() DealerResponse {
	return r.response
}

// Serialize implements serde.Message.
func (r Response) Serialize(ctx serde.Context) ([]byte, error) {
	format := msgFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, r)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode response: %v", err)
	}

	return data, nil
}

// StartDone should be sent by all the nodes to the initiator of the DKG when
// the DKG setup is done.
//
// - implements serde.Message
type StartDone struct {
	pubkey kyber.Point
}

// NewStartDone creates a new start done message.
func NewStartDone(pubkey kyber.Point) StartDone {
	return StartDone{
		pubkey: pubkey,
	}
}

// GetPublicKey returns the public key of the LTS.
func (s StartDone) GetPublicKey() kyber.Point {
	return s.pubkey
}

// Serialize implements serde.Message.
func (s StartDone) Serialize(ctx serde.Context) ([]byte, error) {
	format := msgFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, s)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode ack: %v", err)
	}

	return data, nil
}

// DecryptRequest is a message sent to request a decryption.
//
// - implements serde.Message
type DecryptRequest struct {
	formId string
}

// NewDecryptRequest creates a new decryption request.
func NewDecryptRequest(formId string) DecryptRequest {
	return DecryptRequest{
		formId: formId,
	}
}

// GetFormId returns formId.
func (req DecryptRequest) GetFormId() string {
	return req.formId
}

// Serialize implements serde.Message.
func (req DecryptRequest) Serialize(ctx serde.Context) ([]byte, error) {
	format := msgFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, req)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode decrypt request: %v", err)
	}

	return data, nil
}

// AddrKey is the key for the address factory.
type AddrKey struct{}

// MessageFactory is a message factory for the different DKG messages.
//
// - implements serde.Factory
type MessageFactory struct {
	addrFactory mino.AddressFactory
}

// NewMessageFactory returns a message factory for the DKG.
func NewMessageFactory(f mino.AddressFactory) MessageFactory {
	return MessageFactory{
		addrFactory: f,
	}
}

// Deserialize implements serde.Factory.
func (f MessageFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := msgFormats.Get(ctx.GetFormat())

	ctx = serde.WithFactory(ctx, AddrKey{}, f.addrFactory)

	msg, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("couldn't decode message: %v", err)
	}

	return msg, nil
}

// GetPeerPubKey ...
//
// - implements serde.Message
type GetPeerPubKey struct {
}

// NewGetPeerPubKey creates a new get peer pubkey message.
func NewGetPeerPubKey() GetPeerPubKey {
	return GetPeerPubKey{}
}

// Serialize implements serde.Message.
func (s GetPeerPubKey) Serialize(ctx serde.Context) ([]byte, error) {
	format := msgFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, s)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode GetPeerPubKey: %v", err)
	}

	return data, nil
}

// GetPeerPubKeyResp ...
//
// - implements serde.Message
type GetPeerPubKeyResp struct {
	pubkey kyber.Point
}

// NewGetPeerPubKeyResp creates a new get peer pubkey message.
func NewGetPeerPubKeyResp(pubkey kyber.Point) GetPeerPubKeyResp {
	return GetPeerPubKeyResp{
		pubkey: pubkey,
	}
}

// Serialize implements serde.Message.
func (s GetPeerPubKeyResp) Serialize(ctx serde.Context) ([]byte, error) {
	format := msgFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, s)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode GetPeerPubKeyResp: %v", err)
	}

	return data, nil
}

// GetPublicKey returns the public key of the LTS.
func (s GetPeerPubKeyResp) GetPublicKey() kyber.Point {
	return s.pubkey
}
