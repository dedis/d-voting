package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/sign/schnorr"
)

func TestNewSignedRequest_bad_json(t *testing.T) {
	data := []byte("{invalid json}")
	r := bytes.NewReader(data)

	_, err := NewSignedRequest(r)
	require.EqualError(t, err, "failed to decode signed request: invalid "+
		"character 'i' looking for beginning of object key string")
}

func TestNewSigedRequest_ok(t *testing.T) {
	data := []byte(`{"Payload": "xx", "Signature": "aef123"}`)
	r := bytes.NewReader(data)

	req, err := NewSignedRequest(r)
	require.NoError(t, err)

	expected := SignedRequest{
		Payload:   "xx",
		Signature: "aef123",
	}

	require.Equal(t, expected, req)
}

func TestGetMessage_not_base64(t *testing.T) {
	signed := SignedRequest{
		Payload: "not base 64",
	}

	var req map[string]interface{}

	err := signed.GetMessage(&req)
	require.EqualError(t, err, "failed to decode base64: illegal base64 data at input byte 3")
}

func TestGetMessage_bad_req(t *testing.T) {
	msg := `{invalid json}`
	payload := base64.URLEncoding.EncodeToString([]byte(msg))

	signed := SignedRequest{
		Payload: payload,
	}

	var req map[string]interface{}

	err := signed.GetMessage(&req)
	require.EqualError(t, err, "failed to unmarshal json \"{invalid json}\" "+
		"to *map[string]interface {}: invalid character 'i' looking for "+
		"beginning of object key string")
}

func TestGetMessage_ok(t *testing.T) {
	msg := `{"Foo": "bar"}`
	payload := base64.URLEncoding.EncodeToString([]byte(msg))

	signed := SignedRequest{
		Payload: payload,
	}

	type dummy struct {
		Foo string
	}

	var req dummy

	err := signed.GetMessage(&req)
	require.NoError(t, err)

	expected := dummy{
		Foo: "bar",
	}

	require.Equal(t, expected, req)
}

func TestVerify_bad_signature_hex(t *testing.T) {
	pk := suite.Point()

	signed := SignedRequest{
		Payload:   "yy",
		Signature: "xx",
	}

	err := signed.Verify(pk)
	require.EqualError(t, err, "failed to decode signature: encoding/hex: invalid byte: U+0078 'x'")
}

func TestVerify_bad_signature_invalid(t *testing.T) {
	pk := suite.Point()

	signed := SignedRequest{
		Payload:   "yy",
		Signature: "aef123",
	}

	err := signed.Verify(pk)
	require.EqualError(t, err, "invalid signature: schnorr: signature of invalid length 3 instead of 64")
}

func TestVerify_empty_payload(t *testing.T) {
	pk := suite.Point()

	signed := SignedRequest{
		Signature: "aef123",
	}

	err := signed.Verify(pk)
	require.EqualError(t, err, "cannot verify empty payload")
}

func TestVerify_ok(t *testing.T) {
	secret := suite.Scalar().Pick(suite.RandomStream())
	pk := suite.Point().Mul(secret, nil)

	payload := "xx"

	hash := sha256.New()
	hash.Write([]byte(payload))
	md := hash.Sum(nil)

	signature, err := schnorr.Sign(suite, secret, md)
	require.NoError(t, err)

	signed := SignedRequest{
		Signature: hex.EncodeToString(signature),
		Payload:   payload,
	}

	err = signed.Verify(pk)
	require.NoError(t, err)
}

func TestGetAndVerify_verify_invalid_signature(t *testing.T) {
	pk := suite.Point()

	signed := SignedRequest{
		Signature: "xx",
	}

	var req map[string]interface{}

	err := signed.GetAndVerify(pk, &req)
	require.EqualError(t, err, "failed to verify: cannot verify empty payload")
}

func TestGetAndVerify_wrong_message(t *testing.T) {
	secret := suite.Scalar().Pick(suite.RandomStream())
	pk := suite.Point().Mul(secret, nil)

	msg := `{invalid json}`
	payload := base64.URLEncoding.EncodeToString([]byte(msg))

	hash := sha256.New()
	hash.Write([]byte(payload))
	md := hash.Sum(nil)

	signature, err := schnorr.Sign(suite, secret, md)
	require.NoError(t, err)

	signed := SignedRequest{
		Signature: hex.EncodeToString(signature),
		Payload:   payload,
	}

	var req map[string]interface{}

	err = signed.GetAndVerify(pk, &req)
	require.EqualError(t, err, "failed to get message: failed to unmarshal "+
		"json \"{invalid json}\" to *map[string]interface {}: invalid "+
		"character 'i' looking for beginning of object key string")
}

func TestGetAndVerify_ok(t *testing.T) {
	secret := suite.Scalar().Pick(suite.RandomStream())
	pk := suite.Point().Mul(secret, nil)

	msg := `{"Foo": "bar"}`
	payload := base64.URLEncoding.EncodeToString([]byte(msg))

	hash := sha256.New()
	hash.Write([]byte(payload))
	md := hash.Sum(nil)

	signature, err := schnorr.Sign(suite, secret, md)
	require.NoError(t, err)

	signed := SignedRequest{
		Signature: hex.EncodeToString(signature),
		Payload:   payload,
	}

	type dummy struct {
		Foo string
	}

	var req dummy

	err = signed.GetAndVerify(pk, &req)
	require.NoError(t, err)

	expected := dummy{
		Foo: "bar",
	}

	require.Equal(t, expected, req)
}
