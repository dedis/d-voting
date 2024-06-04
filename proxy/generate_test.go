package proxy

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"testing"
)

func TestGenerateSignature(t *testing.T) {
	msg := `{"TargetUserID" : "654321", "PerformingUserID" : "123456"}`

	payload := base64.URLEncoding.EncodeToString([]byte(msg))

	hash := sha256.New()
	hash.Write([]byte(payload))
	md := hash.Sum(nil)

	pk := "6aadf480d068ac896330b726802abd0da2a5f3824f791fe8dbd4cd555e80b809"
	pkhex, err := hex.DecodeString(pk)
	require.NoError(t, err)

	point := suite.Scalar()
	err = point.UnmarshalBinary(pkhex)
	require.NoError(t, err)

	require.NoError(t, err)

	signature, err := schnorr.Sign(suite, point, md)
	require.NoError(t, err)

	println(hex.EncodeToString(signature))
	println(payload)
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
