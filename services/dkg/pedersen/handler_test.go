package pedersen

import (
	"container/list"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"

	formTypes "github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/mino/minogrpc/session"
	"go.dedis.ch/dela/serde/json"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	pedersen "go.dedis.ch/kyber/v3/share/dkg/pedersen"
)

func TestHandler_Stream(t *testing.T) {
	h := Handler{startRes: &state{}, service: &fake.Service{Forms: make(map[string]formTypes.Form),
		BallotSnap: fake.NewSnapshot()}}
	receiver := fake.NewBadReceiver()
	err := h.Stream(fake.Sender{}, receiver)
	require.EqualError(t, err, fake.Err("failed to receive"))

	receiver = fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), types.Deal{}),
		fake.NewRecvMsg(fake.NewAddress(0), types.DecryptRequest{}),
	)
	err = h.Stream(fake.Sender{}, receiver)
	require.EqualError(t, err, "you must first initialize DKG."+
		" Did you call setup() first?")

	h.startRes.distKey = suite.Point()
	h.startRes.participants = []mino.Address{fake.NewAddress(0)}
	h.privShare = &share.PriShare{I: 0, V: suite.Scalar()}
	receiver = fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), types.DecryptRequest{}),
	)
	err = h.Stream(fake.NewBadSender(), receiver)
	require.EqualError(t, err, "could not send pubShares: failed to check if the shuffle is over: "+
		"could not get the form: while getting data for form: this key doesn't exist")

	formIDHex := hex.EncodeToString([]byte("form"))

	receiver = fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), types.NewDecryptRequest(formIDHex)),
	)

	units := formTypes.PubsharesUnits{
		Pubshares: make([]formTypes.PubsharesUnit, 0),
		PubKeys:   make([][]byte, 0),
		Indexes:   make([]int, 0),
	}

	form := formTypes.Form{
		Configuration:    formTypes.Configuration{},
		FormID:           formIDHex,
		Status:           formTypes.ShuffledBallots,
		Pubkey:           nil,
		BallotSize:       0,
		ShuffleInstances: make([]formTypes.ShuffleInstance, 1),
		ShuffleThreshold: 0,
		PubsharesUnits:   units,
		DecryptedBallots: nil,
		Roster:           fake.Authority{},
	}

	Forms := make(map[string]formTypes.Form)
	Forms[formIDHex] = form

	h.formFac = formTypes.NewFormFactory(formTypes.CiphervoteFactory{}, fake.RosterFac{})

	h.service = &fake.Service{
		Err:        nil,
		Forms:      Forms,
		Pool:       nil,
		Status:     false,
		Channel:    nil,
		Context:    json.NewContext(),
		BallotSnap: fake.NewSnapshot(),
	}

	h.context = json.NewContext()
	h.pubSharesSigner = fake.NewSigner()
	h.txmnger = fake.Manager{}

	err = h.Stream(fake.NewBadSender(), receiver)
	require.NoError(t, err) // Threshold = 0 => no submission required

	receiver = fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), fake.Message{}),
	)
	err = h.Stream(fake.Sender{}, receiver)
	require.EqualError(t, err, "expected Start message, decrypt request or"+
		" Deal as first message, got: fake.Message")
}

func TestHandler_Start(t *testing.T) {
	privKey := suite.Scalar().Pick(suite.RandomStream())
	pubKey := suite.Point().Mul(privKey, nil)

	h := Handler{
		startRes: &state{},
		privKey:  privKey,
		status:   &dkg.Status{},
	}
	start := types.NewStart(
		[]mino.Address{fake.NewAddress(0)},
		[]kyber.Point{},
	)
	err := h.start(start, list.New(), list.New(), nil, nil)
	require.EqualError(t, err, "there should be as many players as pubKey: 1 := 0")

	start = types.NewStart(
		[]mino.Address{fake.NewAddress(0), fake.NewAddress(1)},
		[]kyber.Point{pubKey, suite.Point()},
	)

	err = h.start(start, list.New(), list.New(), nil, fake.Sender{})
	require.NoError(t, err)
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

	responses := list.New()

	dkg = getCertified(t)
	h.dkg = dkg
	err = h.certify(responses, fake.NewBadSender())
	require.NoError(t, err)
}

func TestHandler_HandleDeal_Fail(t *testing.T) {
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
		startRes: &state{
			participants: []mino.Address{fake.NewAddress(0)},
		},
	}
	err = h.handleDeal(dealMsg, fake.NewBadSender())
	require.EqualError(t, err, fake.Err("failed to send response to 'fake.Address[0]'"))

	err = h.handleDeal(dealMsg, fake.Sender{})
	require.True(t, strings.Contains(err.Error(), "failed to process deal"))
}

func TestHandler_HandleDeal_Ok(t *testing.T) {
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
		startRes: &state{
			participants: []mino.Address{fake.NewAddress(0)},
		},
	}

	err = h.handleDeal(dealMsg, fake.Sender{})
	require.NoError(t, err)
}

func TestHandlerData_MarshalJSON(t *testing.T) {
	hd := NewHandlerData()

	data, err := hd.MarshalJSON()
	require.NoError(t, err)

	newHd := &HandlerData{}
	err = newHd.UnmarshalJSON(data)
	require.NoError(t, err)

	require.True(t, newHd.PrivKey.Equal(hd.PrivKey))
	require.True(t, newHd.PubKey.Equal(hd.PubKey))
	requireStatesEqual(t, newHd.StartRes, hd.StartRes)
	require.Equal(t, newHd.PrivShare, hd.PrivShare)
}

func TestState_MarshalJSON(t *testing.T) {
	s1 := &state{}

	// Try with no data
	data, err := s1.MarshalJSON()
	require.NoError(t, err)

	s2 := &state{}
	err = s2.UnmarshalJSON(data)
	require.NoError(t, err)

	requireStatesEqual(t, s1, s2)

	// Try with some data
	distKey := suite.Point().Pick(suite.RandomStream())
	// TODO: use AddressFactory here
	participants := []mino.Address{session.NewAddress("grpcs://localhost:12345"), session.NewAddress("grpcs://localhost:1234")}

	s1.SetDistKey(distKey)
	s1.SetParticipants(participants)

	data, err = s1.MarshalJSON()
	require.NoError(t, err)

	s2 = &state{}
	err = s2.UnmarshalJSON(data)
	require.NoError(t, err)

	requireStatesEqual(t, s1, s2)
}

func TestHandler_HandlerDecryptRequest(t *testing.T) {
	formIDHex := hex.EncodeToString([]byte("form"))

	units := formTypes.PubsharesUnits{
		Pubshares: make([]formTypes.PubsharesUnit, 0),
		PubKeys:   make([][]byte, 0),
		Indexes:   make([]int, 0),
	}

	form := formTypes.Form{
		Configuration:    formTypes.Configuration{},
		FormID:           formIDHex,
		Status:           formTypes.ShuffledBallots,
		Pubkey:           nil,
		BallotSize:       0,
		ShuffleInstances: make([]formTypes.ShuffleInstance, 1),
		ShuffleThreshold: 1,
		PubsharesUnits:   units,
		DecryptedBallots: nil,
		Roster:           fake.Authority{},
	}

	Forms := make(map[string]formTypes.Form)
	Forms[formIDHex] = form

	h := Handler{}

	h.privShare = &share.PriShare{I: 0, V: suite.Scalar()}

	h.formFac = formTypes.NewFormFactory(formTypes.CiphervoteFactory{}, fake.RosterFac{})

	service := fake.Service{
		Err:        nil,
		Forms:      Forms,
		Pool:       nil,
		Status:     false,
		Channel:    nil,
		Context:    json.NewContext(),
		BallotSnap: fake.NewSnapshot(),
	}

	h.context = json.NewContext()
	h.pubSharesSigner = fake.NewSigner()

	pool := fake.Pool{
		Err:         nil,
		Transaction: fake.Transaction{},
		Service:     &service,
	}
	service.Pool = &pool

	h.service = &service
	h.pool = &pool

	// Bad manager:
	h.txmnger = fake.Manager{}

	err := h.handleDecryptRequest(formIDHex)
	require.EqualError(t, err, fake.Err("failed to make tx: failed to use manager"))

	h.txmnger = signed.NewManager(fake.NewSigner(), fakeClient{})

	// All good:

	err = h.handleDecryptRequest(formIDHex)
	require.NoError(t, err)

	// With PubsharesUnit to compute:

	// number of votes
	k := 1

	message := "Hello world"

	Ks, Cs, _ := fakeKCPoints(k, message, suite.Point())

	snap := fake.NewSnapshot()
	for i := 0; i < k; i++ {
		ballot := formTypes.Ciphervote{formTypes.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		form.CastVote(service.Context, snap, "dummyUser"+strconv.Itoa(i), ballot)
	}

	shuffledBallots, err := form.Suffragia(service.Context, snap)
	require.NoError(t, err)
	shuffleInstance := formTypes.ShuffleInstance{ShuffledBallots: shuffledBallots.Ciphervotes}
	form.ShuffleInstances = append(form.ShuffleInstances, shuffleInstance)

	Forms[formIDHex] = form

	err = h.handleDecryptRequest(formIDHex)
	require.NoError(t, err)

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

// NewHandlerDataFull extends NewHandlerData which does not
// initialize all fields
func NewHandlerDataFull() HandlerData {
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
	DistKey1 := s1.GetDistKey()
	DistKey2 := s2.GetDistKey()
	if DistKey1 == nil {
		require.Nil(t, DistKey2)
	} else {
		require.True(t, DistKey2.Equal(DistKey1))
	}
	require.Equal(t, s2.GetParticipants(), s1.GetParticipants())
}

type fakeClient struct{}

func (fakeClient) GetNonce(access.Identity) (uint64, error) {
	return 0, nil
}
