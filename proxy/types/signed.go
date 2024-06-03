package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

var suite = suites.MustFind("ed25519")

// NewSignedRequest returns a new initialized signed request
func NewSignedRequest(r io.Reader) (SignedRequest, error) {
	var req SignedRequest
	decoder := json.NewDecoder(r)

	err := decoder.Decode(&req)
	if err != nil {
		return req, xerrors.Errorf("failed to decode signed request: %v", err)
	}

	println(fmt.Sprintf("helloReq: %v", req))

	return req, nil
}

// SignedRequest represents a frontend request signed by the web backend.
type SignedRequest struct {
	Payload   string // url base64 encoded json message
	Signature string // hex encoded signature on sha256(Payload)
}

// GetMessage JSON unmarshals the payload to the given element. The given
// element MUST be a pointer.
func (s SignedRequest) GetMessage(el interface{}) error {
	payloadBuf, err := base64.URLEncoding.DecodeString(s.Payload)
	if err != nil {
		return xerrors.Errorf("failed to decode base64: %v", err)
	}

	err = json.Unmarshal(payloadBuf, el)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal json %q to %T: %v", payloadBuf, el, err)
	}

	return nil
}

// Verify checks the signature. The signature should be on the sha256 of the
// payload.
func (s SignedRequest) Verify(pk kyber.Point) error {
	if len(s.Payload) == 0 {
		return xerrors.Errorf("cannot verify empty payload")
	}

	hash := sha256.New()

	hash.Write([]byte(s.Payload))
	md := hash.Sum(nil)

	sig, err := hex.DecodeString(s.Signature)
	if err != nil {
		return xerrors.Errorf("failed to decode signature: %v", err)
	}

	err = schnorr.Verify(suite, pk, md, sig)
	if err != nil {
		return xerrors.Errorf("invalid signature: %v", err)
	}

	return nil
}

// GetAndVerify is a shorthand function to verify the signed request and extract
// the payload. el MUST be a pointer.
func (s SignedRequest) GetAndVerify(pk kyber.Point, el interface{}) error {
	err := s.Verify(pk)
	if err != nil {
		return xerrors.Errorf("failed to verify: %v", err)
	}

	err = s.GetMessage(el)
	if err != nil {
		return xerrors.Errorf("failed to get message: %v", err)
	}

	return nil
}
