package json

import (
        "fmt"
        "testing"

        "github.com/dedis/d-voting/internal/testing/fake"
        "github.com/dedis/d-voting/services/dkg/pedersen/types"
        "github.com/stretchr/testify/require"
        "go.dedis.ch/dela/mino"
        "go.dedis.ch/dela/serde"
        "go.dedis.ch/kyber/v3"
        "go.dedis.ch/kyber/v3/suites"
)

// suite is the Kyber suite for Pedersen.
var suite = suites.MustFind("Ed25519")

func TestMessageFormat_Start_Encode(t *testing.T) {
        start := types.NewStart([]mino.Address{fake.NewAddress(0)}, []kyber.Point{suite.Point()})

        format := newMsgFormat()
        ctx := serde.NewContext(fake.ContextEngine{})

        data, err := format.Encode(ctx, start)
        require.NoError(t, err)
        regexp := `{"Start":{"Threshold":1,"Addresses":\["AAAAAA=="\],"PublicKeys":\["[^"]+"\]}}`
        require.Regexp(t, regexp, string(data))

        start = types.NewStart([]mino.Address{fake.NewBadAddress()}, nil)
        _, err = format.Encode(ctx, start)
        require.EqualError(t, err, fake.Err("couldn't marshal address"))

        start = types.NewStart(nil, []kyber.Point{badPoint{}})
        _, err = format.Encode(ctx, start)
        require.EqualError(t, err, fake.Err("couldn't marshal public key"))

        _, err = format.Encode(fake.NewBadContext(), types.Start{})
        require.EqualError(t, err, fake.Err("couldn't marshal"))

        _, err = format.Encode(ctx, fake.Message{})
        require.EqualError(t, err, "unsupported message of type 'fake.Message'")
}

func TestMessageFormat_Deal_Encode(t *testing.T) {
        deal := types.NewDeal(1, []byte{1}, types.EncryptedDeal{})

        format := newMsgFormat()
        ctx := serde.NewContext(fake.ContextEngine{})

        data, err := format.Encode(ctx, deal)
        require.NoError(t, err)
        expected := `{"Deal":{"Index":1,"Signature":"AQ==","EncryptedDeal":{"DHKey":"","Signature":"","Nonce":"","Cipher":""}}}`
        require.Equal(t, expected, string(data))
}

func TestMessageFormat_Response_Encode(t *testing.T) {
        resp := types.NewResponse(1, types.DealerResponse{})

        format := newMsgFormat()
        ctx := serde.NewContext(fake.ContextEngine{})

        data, err := format.Encode(ctx, resp)
        require.NoError(t, err)
        expected := `{"Response":{"Index":1,"Response":{"SessionID":"","Index":0,"Status":false,"Signature":""}}}`
        require.Equal(t, expected, string(data))
}

func TestMessageFormat_StartDone_Encode(t *testing.T) {
        done := types.NewStartDone(suite.Point())

        format := newMsgFormat()
        ctx := serde.NewContext(fake.ContextEngine{})

        data, err := format.Encode(ctx, done)
        require.NoError(t, err)
        require.Regexp(t, `{(("StartDone":{"PublicKey":"[^"]+"}|"\w+":null),?)+}`, string(data))

        done = types.NewStartDone(badPoint{})
        _, err = format.Encode(ctx, done)
        require.EqualError(t, err, fake.Err("couldn't marshal public key"))
}

func TestMessageFormat_DecryptRequest_Encode(t *testing.T) {
        req := types.NewDecryptRequest(suite.Point(), suite.Point(), "electionId")

        format := newMsgFormat()
        ctx := serde.NewContext(fake.ContextEngine{})

        data, err := format.Encode(ctx, req)
        require.NoError(t, err)
        require.Regexp(t, `{(("DecryptRequest":{"K":"[^"]+","C":"[^"]+","ElectionId":"electionId"}|"\w+":null),?)+}`, string(data))

        req.K = badPoint{}
        _, err = format.Encode(ctx, req)
        require.EqualError(t, err, fake.Err("couldn't marshal K"))

        req.K = suite.Point()
        req.C = badPoint{}
        _, err = format.Encode(ctx, req)
        require.EqualError(t, err, fake.Err("couldn't marshal C"))
}

func TestMessageFormat_DecryptReply_Encode(t *testing.T) {
        resp := types.NewDecryptReply(5, suite.Point())

        format := newMsgFormat()
        ctx := serde.NewContext(fake.ContextEngine{})

        data, err := format.Encode(ctx, resp)
        require.NoError(t, err)
        require.Regexp(t, `{(("DecryptReply":{"V":"[^"]+","I":5}|"\w+":null),?)+}`, string(data))

        resp.V = badPoint{}
        _, err = format.Encode(ctx, resp)
        require.EqualError(t, err, fake.Err("couldn't marshal V"))
}

func TestMessageFormat_Decode(t *testing.T) {
        format := newMsgFormat()
        ctx := serde.NewContext(fake.ContextEngine{})
        ctx = serde.WithFactory(ctx, types.AddrKey{}, fake.AddressFactory{})

        // Decode start messages.
        expected := types.NewStart(
                []mino.Address{fake.NewAddress(0)},
                []kyber.Point{suite.Point()},
        )

        data, err := format.Encode(ctx, expected)
        require.NoError(t, err)

        start, err := format.Decode(ctx, data)
        require.NoError(t, err)
        require.Len(t, start.(types.Start).GetAddresses(), len(expected.GetAddresses()))
        require.Len(t, start.(types.Start).GetPublicKeys(), len(expected.GetPublicKeys()))

        _, err = format.Decode(ctx, []byte(`{"Start":{"PublicKeys":[[]]}}`))
        require.EqualError(t, err,
                "couldn't unmarshal public key: invalid Ed25519 curve point")

        badCtx := serde.WithFactory(ctx, types.AddrKey{}, nil)
        _, err = format.Decode(badCtx, []byte(`{"Start":{}}`))
        require.EqualError(t, err, "invalid factory of type '<nil>'")

        // Decode deal messages.
        deal, err := format.Decode(ctx, []byte(`{"Deal":{}}`))
        require.NoError(t, err)
        require.Equal(t, types.NewDeal(0, nil, types.EncryptedDeal{}), deal)

        // Decode response messages.
        resp, err := format.Decode(ctx, []byte(`{"Response":{}}`))
        require.NoError(t, err)
        require.Equal(t, types.NewResponse(0, types.DealerResponse{}), resp)

        // Decode start done messages.
        data = []byte(fmt.Sprintf(`{"StartDone":{"PublicKey":"%s"}}`, testPoint))
        done, err := format.Decode(ctx, data)
        require.NoError(t, err)
        require.IsType(t, types.StartDone{}, done)

        data = []byte(`{"StartDone":{"PublicKey":[]}}`)
        _, err = format.Decode(ctx, data)
        require.EqualError(t, err,
                "couldn't unmarshal public key: invalid Ed25519 curve point")

        // Decode decryption request messages.
        data = []byte(fmt.Sprintf(`{"DecryptRequest":{"K":"%s","C":"%s"}}`, testPoint, testPoint))
        req, err := format.Decode(ctx, data)
        require.NoError(t, err)
        require.IsType(t, types.DecryptRequest{}, req)

        data = []byte(fmt.Sprintf(`{"DecryptRequest":{"K":[],"C":"%s"}}`, testPoint))
        _, err = format.Decode(ctx, data)
        require.EqualError(t, err,
                "couldn't unmarshal K: invalid Ed25519 curve point")

        data = []byte(fmt.Sprintf(`{"DecryptRequest":{"K":"%s","C":[]}}`, testPoint))
        _, err = format.Decode(ctx, data)
        require.EqualError(t, err,
                "couldn't unmarshal C: invalid Ed25519 curve point")

        // Decode decryption reply messages.
        data = []byte(fmt.Sprintf(`{"DecryptReply":{"I":4,"V":"%s"}}`, testPoint))
        resp, err = format.Decode(ctx, data)
        require.NoError(t, err)
        require.IsType(t, types.DecryptReply{}, resp)

        data = []byte(`{"DecryptReply":{"V":[]}}`)
        _, err = format.Decode(ctx, data)
        require.EqualError(t, err,
                "couldn't unmarshal V: invalid Ed25519 curve point")

        _, err = format.Decode(fake.NewBadContext(), []byte(`{}`))
        require.EqualError(t, err, fake.Err("couldn't deserialize message"))

        _, err = format.Decode(ctx, []byte(`{}`))
        require.EqualError(t, err, "message is empty")
}

// -----------------------------------------------------------------------------
// Utility functions

const testPoint = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

type badPoint struct {
        kyber.Point
}

func (p badPoint) MarshalBinary() ([]byte, error) {
        return nil, fake.GetError()
}
