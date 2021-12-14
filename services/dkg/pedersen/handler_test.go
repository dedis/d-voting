package pedersen

import (
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	pedersen "go.dedis.ch/kyber/v3/share/dkg/pedersen"
	vss "go.dedis.ch/kyber/v3/share/vss/pedersen"
)

func TestHandler_Stream(t *testing.T) {
	h := Handler{startRes: &state{}}
	receiver := fake.NewBadReceiver()
	err := h.Stream(fake.Sender{}, receiver)
	require.EqualError(t, err, fake.Err("failed to receive"))

	receiver = fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), types.Deal{}),
		fake.NewRecvMsg(fake.NewAddress(0), types.DecryptRequest{}),
	)
	err = h.Stream(fake.Sender{}, receiver)
	require.EqualError(t, err, "you must first initialize DKG. Did you call setup() first?")

	h.startRes.distKey = suite.Point()
	h.startRes.participants = []mino.Address{fake.NewAddress(0)}
	h.privShare = &share.PriShare{I: 0, V: suite.Scalar()}
	receiver = fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), types.DecryptRequest{C: suite.Point()}),
	)
	err = h.Stream(fake.NewBadSender(), receiver)
	require.EqualError(t, err, fake.Err("got an error while sending the decrypt reply"))

	receiver = fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), fake.Message{}),
	)
	err = h.Stream(fake.Sender{}, receiver)
	require.EqualError(t, err, "expected Start message, decrypt request or Deal as first message, got: fake.Message")
}

func TestHandler_Start(t *testing.T) {
	privKey := suite.Scalar().Pick(suite.RandomStream())
	pubKey := suite.Point().Mul(privKey, nil)

	h := Handler{
		startRes: &state{},
		privKey:  privKey,
	}
	start := types.NewStart(
		[]mino.Address{fake.NewAddress(0)},
		[]kyber.Point{},
	)
	err := h.start(start, []types.Deal{}, []*pedersen.Response{}, nil, nil, nil)
	require.EqualError(t, err, "there should be as many players as pubKey: 1 := 0")

	start = types.NewStart(
		[]mino.Address{fake.NewAddress(0), fake.NewAddress(1)},
		[]kyber.Point{pubKey, suite.Point()},
	)
	receiver := fake.NewBadReceiver()
	err = h.start(start, []types.Deal{}, []*pedersen.Response{}, nil, fake.Sender{}, receiver)
	require.EqualError(t, err, fake.Err("failed to receive after sending deals"))

	receiver = fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), types.Deal{}),
		fake.NewRecvMsg(fake.NewAddress(0), nil),
	)
	err = h.start(start, []types.Deal{}, []*pedersen.Response{}, nil, fake.Sender{}, receiver)
	require.EqualError(t, err, "failed to handle deal from 'fake.Address[0]': failed to process deal from %!s(<nil>): schnorr: signature of invalid length 0 instead of 64")

	err = h.start(start, []types.Deal{}, []*pedersen.Response{}, nil, fake.Sender{}, &fake.Receiver{})
	require.EqualError(t, err, "unexpected message: <nil>")

	// We check when there is already something in the slice if Deals
	err = h.start(start, []types.Deal{{}}, []*pedersen.Response{}, nil, fake.NewBadSender(), &fake.Receiver{})
	require.EqualError(t, err, "failed to certify: expected a response, got: <nil>")
}

func TestHandler_Certify(t *testing.T) {
	privKey := suite.Scalar().Pick(suite.RandomStream())
	pubKey := suite.Point().Mul(privKey, nil)

	dkg, err := pedersen.NewDistKeyGenerator(suite, privKey, []kyber.Point{pubKey, suite.Point()}, 2)
	require.NoError(t, err)

	h := Handler{
		startRes: &state{},
		dkg:      dkg,
	}
	receiver := fake.NewBadReceiver()
	responses := []*pedersen.Response{{Response: &vss.Response{}}}

	err = h.certify(responses, fake.Sender{}, receiver, nil)
	require.EqualError(t, err, fake.Err("failed to receive after sending deals"))

	dkg = getCertified(t)
	h.dkg = dkg
	err = h.certify(responses, fake.NewBadSender(), &fake.Receiver{}, nil)
	require.EqualError(t, err, fake.Err("got an error while sending pub key"))
}

func TestHandler_HandleDeal(t *testing.T) {
	privKey1 := suite.Scalar().Pick(suite.RandomStream())
	pubKey1 := suite.Point().Mul(privKey1, nil)
	privKey2 := suite.Scalar().Pick(suite.RandomStream())
	pubKey2 := suite.Point().Mul(privKey2, nil)

	dkg1, err := pedersen.NewDistKeyGenerator(suite, privKey1, []kyber.Point{pubKey1, pubKey2}, 2)
	require.NoError(t, err)

	dkg2, err := pedersen.NewDistKeyGenerator(suite, privKey2, []kyber.Point{pubKey1, pubKey2}, 2)
	require.NoError(t, err)

	deals, err := dkg2.Deals()
	require.Len(t, deals, 1)
	require.NoError(t, err)

	var deal *pedersen.Deal
	for _, d := range deals {
		deal = d
	}

	dealMsg := types.NewDeal(
		deal.Index,
		deal.Signature,
		types.NewEncryptedDeal(
			deal.Deal.DHKey,
			deal.Deal.Signature,
			deal.Deal.Nonce,
			deal.Deal.Cipher,
		),
	)

	h := Handler{
		dkg: dkg1,
	}
	err = h.handleDeal(dealMsg, nil, []mino.Address{fake.NewAddress(0)}, fake.NewBadSender())
	require.EqualError(t, err, fake.Err("failed to send response to 'fake.Address[0]'"))
}

func TestHandlerData_MarshalJSON(t *testing.T) {
	hd := HandlerDataTest()

	data, err := hd.MarshalJSON()
	require.NoError(t, err)

	newHd := &HandlerData{}
	err = newHd.UnmarshalJSON(data)
	require.NoError(t, err)

	require.True(t, newHd.PrivKey.Equal(hd.PrivKey))
	require.True(t, newHd.PubKey.Equal(hd.PubKey))
	requireStatesEqual(t, newHd.StartRes, hd.StartRes)
	require.Equal(t, newHd.PrivShare, hd.PrivShare)
	// requirePriShareEqual(t, newHd.PrivShare, hd.PrivShare)
}

func TestState_MarshalJSON(t *testing.T) {
	distKey := suite.Point().Pick(suite.RandomStream())
	participants := []mino.Address{fake.NewAddress(0), fake.NewAddress(1)}

	s1 := &state{}
	s1.SetDistKey(distKey)
	s1.SetParticipants(participants)

	data, err := s1.MarshalJSON()
	require.NoError(t, err)

	s2 := &state{}
	err = s2.UnmarshalJSON(data)
	require.NoError(t, err)

	requireStatesEqual(t, s1, s2)
}

// -----------------------------------------------------------------------------
// Utility functions

func getCertified(t *testing.T) *pedersen.DistKeyGenerator {
	privKey1 := suite.Scalar().Pick(suite.RandomStream())
	pubKey1 := suite.Point().Mul(privKey1, nil)
	privKey2 := suite.Scalar().Pick(suite.RandomStream())
	pubKey2 := suite.Point().Mul(privKey2, nil)

	dkg1, err := pedersen.NewDistKeyGenerator(suite, privKey1, []kyber.Point{pubKey1, pubKey2}, 2)
	require.NoError(t, err)
	dkg2, err := pedersen.NewDistKeyGenerator(suite, privKey2, []kyber.Point{pubKey1, pubKey2}, 2)
	require.NoError(t, err)

	deals1, err := dkg1.Deals()
	require.NoError(t, err)
	require.Len(t, deals1, 1)

	deals2, err := dkg2.Deals()
	require.NoError(t, err)
	require.Len(t, deals2, 1)

	var resp1 *pedersen.Response
	var resp2 *pedersen.Response

	for _, deal := range deals2 {
		resp1, err = dkg1.ProcessDeal(deal)
		require.NoError(t, err)
	}
	for _, deal := range deals1 {
		resp2, err = dkg2.ProcessDeal(deal)
		require.NoError(t, err)
	}

	_, err = dkg1.ProcessResponse(resp2)
	require.NoError(t, err)
	_, err = dkg2.ProcessResponse(resp1)
	require.NoError(t, err)

	require.True(t, dkg1.Certified())
	require.True(t, dkg2.Certified())

	return dkg1
}

func HandlerDataTest() HandlerData {
	hd := NewHandlerData()

	// Set StartRes
	distKey := suite.Point().Pick(suite.RandomStream())
	participants := []mino.Address{fake.NewAddress(0), fake.NewAddress(1)}

	hd.StartRes.SetDistKey(distKey)
	hd.StartRes.SetParticipants(participants)

	// Set PrivShare
	hd.PrivShare = &share.PriShare{
		I: 0,
		V: suite.Scalar().Pick(suite.RandomStream()),
	}

	return hd
}

func requireStatesEqual(t *testing.T, s1, s2 *state) {
	require.True(t, s2.GetDistKey().Equal(s1.GetDistKey()))
	require.Equal(t, s2.GetParticipants(), s1.GetParticipants())
}
