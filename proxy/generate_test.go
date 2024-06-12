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

/*
CURL to add 123456 as an admin
--
curl -XPOST -H "Content-type: application/json" -d '{
"Payload":"eyJUYXJnZXRVc2VySUQiIDogIjEyMzQ1NiIsICJQZXJmb3JtaW5nVXNlcklEIiA6ICIxMjM0NTYifQ==", "Signature":"942501b9ae78b63aafcfd65b582087de93c877c53754fdb6221de0444bd6d56240882042395a79b601e6cce36e10b474c42b96b12d7a0cb5909c725025978f09"}' 'http://172.19.44.254:8080/evoting/addadmin'
*/

/*
CURL to add 654321 as an admin
--
curl -XPOST -H "Content-type: application/json" -d '{
"Payload":"eyJUYXJnZXRVc2VySUQiIDogIjY1NDMyMSIsICJQZXJmb3JtaW5nVXNlcklEIiA6ICIxMjM0NTYifQ==", "Signature":"b02557ce9133fe945dc3dd47e53883e007f06c30031b7e42159b38623923d19a5b05f1012edc8ea580fbff233ab819bd1d6b9dc72b881c2d1c9dd9f681d07a0a"}' 'http://172.19.44.254:8080/evoting/addadmin'
*/

/*
CURL to remove 654321 as an admin
--
curl -XPOST -H "Content-type: application/json" -d '{
"Payload":"eyJUYXJnZXRVc2VySUQiIDogIjY1NDMyMSIsICJQZXJmb3JtaW5nVXNlcklEIiA6ICIxMjM0NTYifQ==", "Signature":"b02557ce9133fe945dc3dd47e53883e007f06c30031b7e42159b38623923d19a5b05f1012edc8ea580fbff233ab819bd1d6b9dc72b881c2d1c9dd9f681d07a0a"}' 'http://172.19.44.254:8080/evoting/removeadmin'
*/

/*
CURL to get admin list
--
curl -XGET -H "Content-type: application/json" 'http://172.19.44.254:8080/evoting/adminlist'
*/
