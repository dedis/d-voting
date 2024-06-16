package proxy

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"testing"
)

/*
TestGenerateSignatureAndB64Payload is a code snippet that help developer
to generate Signature and Base64 Payload to create cURL request for debug environment.

How to use it:
- Feel free to change the INPUT following docs/api.md
- Run the test and use the provided result as a cURL request.
*/
func TestGenerateSignatureAndB64Payload(t *testing.T) {
	// #### INPUT ####
	// rawPayload must be built following docs/api.md
	rawPayload := `{"TargetUserID" : "654321", "PerformingUserID" : "123456"}`

	// pk must be set according to the public key used to run the system.
	pk := "6aadf480d068ac896330b726802abd0da2a5f3824f791fe8dbd4cd555e80b809"
	// #### END INPUT ####

	// #### DO NOT MODIFY BELOW ####
	payload := base64.URLEncoding.EncodeToString([]byte(rawPayload))

	hash := sha256.New()
	hash.Write([]byte(payload))
	md := hash.Sum(nil)

	pkhex, err := hex.DecodeString(pk)
	require.NoError(t, err)

	point := suite.Scalar()
	err = point.UnmarshalBinary(pkhex)
	require.NoError(t, err)

	require.NoError(t, err)

	signature, err := schnorr.Sign(suite, point, md)
	require.NoError(t, err)

	println("Signature: " + hex.EncodeToString(signature))
	println("Payload: " + payload)
}
