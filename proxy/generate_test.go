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
	msg := `{"TargetUserID" : "123456", "PerformingUserID" : "123456"}`

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
curl -XPOST -H "Content-type: application/json" -d '{
"Payload":"eyJUYXJnZXRVc2VySUQiIDogIjEyMzQ1NiIsICJQZXJmb3JtaW5nVXNlcklEIiA6ICIxMjM0NTYifQ==", "Signature":"942501b9ae78b63aafcfd65b582087de93c877c53754fdb6221de0444bd6d56240882042395a79b601e6cce36e10b474c42b96b12d7a0cb5909c725025978f09"}' 'http://172.19.44.254:8080/evoting/addadmin'
*/
