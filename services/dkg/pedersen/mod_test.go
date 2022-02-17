package pedersen

import (
	"encoding/hex"
	"encoding/json"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/core/validation"
	"golang.org/x/xerrors"
	"strconv"
	"testing"
	"time"

	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/store/kv"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/mino/minogrpc"
	"go.dedis.ch/dela/mino/router/tree"
	"go.dedis.ch/dela/serde"
	sjson "go.dedis.ch/dela/serde/json"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/util/random"
)

var serdecontext serde.Context
var electionFac serde.Factory
var transactionFac serde.Factory

func init() {
	serdecontext = sjson.NewContext()

	ciphervoteFac := etypes.CiphervoteFactory{}
	electionFac = etypes.NewElectionFactory(ciphervoteFac, fake.Factory{})
	transactionFac = etypes.NewTransactionFactory(ciphervoteFac)
}

// If you get the persistent data from an actor and then recreate an actor
// from that data, the persistent data should be the same in both actors.
func TestActor_MarshalJSON(t *testing.T) {
	p := NewPedersen(fake.Mino{}, fake.Service{}, &fake.Pool{}, fake.Factory{}, fake.Signer{})

	// Create new actor
	actor1, err := p.NewActor([]byte("deadbeef"), &fake.Pool{},
		signed.NewManager(fake.Signer{}, &client{}), NewHandlerData())
	require.NoError(t, err)

	// Serialize its persistent data
	actor1Buf, err := actor1.MarshalJSON()
	require.NoError(t, err)

	// Create a new actor with that data
	handlerData := HandlerData{}
	err = handlerData.UnmarshalJSON(actor1Buf)
	require.NoError(t, err)

	actor2, err := p.NewActor([]byte("beefdead"), &fake.Pool{},
		signed.NewManager(fake.Signer{}, &client{}), handlerData)
	require.NoError(t, err)

	// Check that the persistent data is the same for both actors
	requireActorsEqual(t, actor1, actor2)
}

// After initializing a Pedersen when dkgMap is not empty, the actors map should
// contain the same information as dkgMap
func TestPedersen_InitNonEmptyMap(t *testing.T) {
	// Create a new DKG map and fill it with data
	dkgMap := fake.NewInMemoryDB()

	distKey := suite.Point().Mul(fake.NewScalar(), nil)
	privKey := fake.NewScalar()
	pubKey := suite.Point().Mul(privKey, nil)

	hd := HandlerData{
		StartRes: &state{
			distKey:      distKey,
			participants: []mino.Address{fake.NewAddress(0), fake.NewAddress(1)},
		},
		PrivShare: &share.PriShare{
			I: 1,
			V: fake.NewScalar(),
		},
		PubKey:  pubKey,
		PrivKey: privKey,
	}
	electionActorMap := map[string]HandlerData{
		"deadbeef51": hd,
		"deadbeef52": NewHandlerData(),
	}

	err := dkgMap.Update(func(tx kv.WritableTx) error {
		bucket, err := tx.GetBucketOrCreate([]byte("dkgmap"))
		if err != nil {
			return err
		}

		for electionID, handlerData := range electionActorMap {

			electionIDBuf, err := hex.DecodeString(electionID)
			if err != nil {
				return err
			}

			handlerDataBuf, err := handlerData.MarshalJSON()
			if err != nil {
				return err
			}

			err = bucket.Set(electionIDBuf, handlerDataBuf)
			if err != nil {
				return err
			}
		}

		return nil
	})
	require.NoError(t, err)

	// Initialize a Pedersen
	p := NewPedersen(fake.Mino{}, fake.Service{}, &fake.Pool{}, fake.Factory{}, fake.Signer{})

	err = dkgMap.View(func(tx kv.ReadableTx) error {
		bucket := tx.GetBucket([]byte("dkgmap"))
		require.NotNil(t, bucket)

		return bucket.ForEach(func(electionIDBuf, handlerDataBuf []byte) error {

			handlerData := HandlerData{}
			err = handlerData.UnmarshalJSON(handlerDataBuf)
			if err != nil {
				return err
			}

			_, err = p.NewActor(electionIDBuf, &fake.Pool{},
				signed.NewManager(fake.Signer{}, &client{}), handlerData)
			if err != nil {
				return err
			}

			return nil
		})
	})
	require.NoError(t, err)

	// Check that the data was used properly

	// Check the number of elements is the same
	require.Equal(t, len(electionActorMap), len(p.actors))

	// Check equality pair by pair
	for electionID, handlerData := range electionActorMap {

		electionIDBuf, err := hex.DecodeString(electionID)
		require.NoError(t, err)

		actor, exists := p.GetActor(electionIDBuf)
		require.True(t, exists)

		otherActor := Actor{
			handler: NewHandler(fake.NewAddress(0), fake.Service{}, &fake.Pool{},
				signed.NewManager(fake.Signer{}, &client{}), fake.Signer{},
				handlerData, serdecontext, electionFac),
		}

		requireActorsEqual(t, actor, &otherActor)
	}
}

// When a new actor is created, its information is safely stored in the dkgMap.
func TestPedersen_SyncDB(t *testing.T) {
	electionID1 := "deadbeef51"
	electionID2 := "deadbeef52"

	// Start some elections
	fake.NewElection(electionID1)
	fake.NewElection(electionID2)

	// Initialize a Pedersen
	p := NewPedersen(fake.Mino{}, fake.Service{}, &fake.Pool{}, fake.Factory{}, fake.Signer{})

	// Create actors
	a1, err := p.NewActor([]byte(electionID1), &fake.Pool{},
		signed.NewManager(fake.Signer{}, &client{}), NewHandlerData())
	require.NoError(t, err)
	_, err = p.NewActor([]byte(electionID2), &fake.Pool{},
		signed.NewManager(fake.Signer{}, &client{}), NewHandlerData())
	require.NoError(t, err)

	// Only Setup the first actor
	a1.Setup()

	// Create a new DKG map and fill it with data
	dkgMap := fake.NewInMemoryDB()

	// Store them in the map
	err = dkgMap.Update(func(tx kv.WritableTx) error {
		bucket, err := tx.GetBucketOrCreate([]byte("dkgmap"))
		if err != nil {
			return err
		}

		for electionID, actor := range p.actors {

			electionIDBuf, err := hex.DecodeString(electionID)
			if err != nil {
				return err
			}

			handlerDataBuf, err := actor.MarshalJSON()
			if err != nil {
				return err
			}

			err = bucket.Set(electionIDBuf, handlerDataBuf)
			if err != nil {
				return err
			}
		}

		return nil
	})
	require.NoError(t, err)

	// Recover them from the map
	q := NewPedersen(fake.Mino{}, fake.Service{}, &fake.Pool{}, fake.Factory{}, fake.Signer{})

	err = dkgMap.View(func(tx kv.ReadableTx) error {
		bucket := tx.GetBucket([]byte("dkgmap"))
		require.NotNil(t, bucket)

		return bucket.ForEach(func(electionIDBuf, handlerDataBuf []byte) error {

			handlerData := HandlerData{}
			err = json.Unmarshal(handlerDataBuf, &handlerData)
			require.NoError(t, err)

			_, err = q.NewActor(electionIDBuf, &fake.Pool{},
				signed.NewManager(fake.Signer{}, &client{}), handlerData)
			require.NoError(t, err)

			return nil
		})
	})
	require.NoError(t, err)

	// Check the information is conserved

	require.Equal(t, len(q.actors), len(p.actors))

	// Check equality of actor data
	for electionID, actorQ := range q.actors {

		electionIDBuf, err := hex.DecodeString(electionID)
		require.NoError(t, err)

		actorP, exists := p.GetActor(electionIDBuf)
		require.True(t, exists)

		requireActorsEqual(t, actorP, actorQ)
	}
}

func TestPedersen_Listen(t *testing.T) {
	electionID := "d3adbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	p := NewPedersen(fake.Mino{}, fake.NewService(electionID,
		etypes.Election{Roster: fake.Authority{}}, serdecontext), &fake.Pool{},
		fake.Factory{}, fake.Signer{})

	actor, err := p.Listen(electionIDBuf, signed.NewManager(fake.Signer{}, &client{}))
	require.NoError(t, err)

	require.NotNil(t, actor)
}

// If Listen is called twice for the same election, the actor data is unchanged
func TestPedersen_TwoListens(t *testing.T) {
	electionID := "deadbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	p := NewPedersen(fake.Mino{}, fake.NewService(electionID,
		etypes.Election{Roster: fake.Authority{}}, serdecontext), &fake.Pool{},
		fake.Factory{}, fake.Signer{})

	actor1, err := p.Listen(electionIDBuf, signed.NewManager(fake.Signer{}, &client{}))
	require.NoError(t, err)

	actor2, err := p.Listen(electionIDBuf, signed.NewManager(fake.Signer{}, &client{}))
	require.Error(t, err, "actor already exists for electionID deadbeef")

	require.Equal(t, actor1, actor2)
}

func TestPedersen_Setup(t *testing.T) {
	electionID := "d3adbeef"

	actor := Actor{
		rpc:     nil,
		factory: nil,
		service: fake.NewService(electionID, etypes.Election{
			ElectionID: electionID,
			Roster:     fake.Authority{},
		}, serdecontext),
		handler: &Handler{
			startRes: &state{},
		},
		context:     serdecontext,
		electionFac: electionFac,
	}

	// Wrong electionID
	wrongElectionID := "beefdead"
	actor.electionID = wrongElectionID

	_, err := actor.Setup()
	require.EqualError(t, err, "failed to get election: election does not exist: <nil>")

	actor.electionID = electionID

	// RPC is bogus 1
	actor.rpc = fake.NewBadRPC()

	_, err = actor.Setup()
	require.EqualError(t, err, fake.Err("failed to stream"))

	// RPC is bogus 2
	actor.rpc = fake.NewRPC()

	_, err = actor.Setup()
	require.EqualError(t, err, "the list of addresses is empty")

	// RPC is bogus 3
	actor.rpc = fake.NewStreamRPC(fake.NewReceiver(), fake.NewBadSender())

	_, err = actor.Setup()
	require.EqualError(t, err, fake.Err("failed to send getPeerKey message"))

	// The public keys do not match
	// Create the roster
	rosterLen := 2
	roster := authority.FromAuthority(fake.NewAuthority(rosterLen, fake.NewSigner))

	addrs := make([]mino.Address, 0, rosterLen)
	addrsIter := roster.AddressIterator()
	for addrsIter.HasNext() {
		addrs = append(addrs, addrsIter.GetNext())
	}

	// This fake RosterFac always returns roster upon Deserialize
	fac := etypes.NewElectionFactory(etypes.CiphervoteFactory{}, fake.NewRosterFac(roster))
	actor.electionFac = fac

	actor.service = fake.NewService(
		electionID,
		etypes.Election{
			ElectionID: electionID,
			Roster:     roster,
		},
		serdecontext,
	)
	pubKey1 := suite.Point().Pick(suite.RandomStream())
	pubKey2 := suite.Point().Pick(suite.RandomStream())

	actor.rpc = fake.NewStreamRPC(fake.NewReceiver(
		fake.NewRecvMsg(addrs[0], types.NewGetPeerPubKeyResp(pubKey1)),
		fake.NewRecvMsg(addrs[1], types.NewGetPeerPubKeyResp(pubKey2)),
		fake.NewRecvMsg(addrs[0], types.NewStartDone(pubKey1)),
		fake.NewRecvMsg(addrs[1], types.NewStartDone(pubKey2)),
	), fake.Sender{})

	_, err = actor.Setup()
	require.Regexp(t, "^the public keys do not match:", err)

	// Everything works now
	actor.rpc = fake.NewStreamRPC(fake.NewReceiver(
		fake.NewRecvMsg(addrs[0], types.NewGetPeerPubKeyResp(pubKey2)),
		fake.NewRecvMsg(addrs[1], types.NewGetPeerPubKeyResp(pubKey2)),
		fake.NewRecvMsg(addrs[0], types.NewStartDone(pubKey2)),
		fake.NewRecvMsg(addrs[1], types.NewStartDone(pubKey2)),
	), fake.Sender{})

	// This will not change startRes since the responses are all
	// simulated, so running setup() several times will work.
	// We test that particular behaviour later.
	_, err = actor.Setup()
	require.NoError(t, err)
}

func TestPedersen_Decrypt(t *testing.T) {
	//
	//actor := Actor{
	//	rpc: fake.NewBadRPC(),
	//	handler: &Handler{
	//		startRes: &state{participants: []mino.Address{fake.NewAddress(0)},
	//			distKey: suite.Point()},
	//	},
	//	context:     serdecontext,
	//	electionFac: electionFac,
	//}
	//
	//_, err := actor.Decrypt(suite.Point(), suite.Point())
	//require.EqualError(t, err, fake.Err("failed to create stream"))
	//rpc := fake.NewStreamRPC(fake.NewBadReceiver(), fake.NewBadSender())
	//actor.rpc = rpc
	//
	//_, err = actor.Decrypt(suite.Point(), suite.Point())
	//require.EqualError(t, err, fake.Err("failed to send decrypt request"))
	//
	//recv := fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), nil))
	//
	//rpc = fake.NewStreamRPC(recv, fake.Sender{})
	//actor.rpc = rpc
	//
	//_, err = actor.Decrypt(suite.Point(), suite.Point())
	//require.EqualError(t, err, "got unexpected reply,
	// 			expected types.DecryptReply but got: <nil>")
	//
	///*
	//recv = fake.NewReceiver(
	//	fake.NewRecvMsg(fake.NewAddress(0),
	//			types.DecryptReply{I: -1, V: suite.Point()}),
	//)*/
	//
	//rpc = fake.NewStreamRPC(recv, fake.Sender{})
	//actor.rpc = rpc
	//
	//_, err = actor.Decrypt(suite.Point(), suite.Point())
	//require.EqualError(t, err, "failed to recover commit: share: not enough "+
	//	"good public shares to reconstruct secret commitment")
	//
	///*
	//recv = fake.NewReceiver(
	//	fake.NewRecvMsg(fake.NewAddress(0),
	//			types.DecryptReply{I: 1, V: suite.Point()}),
	//)*/
	//
	//rpc = fake.NewStreamRPC(recv, fake.Sender{})
	//actor.rpc = rpc
	//
	//_, err = actor.Decrypt(suite.Point(), suite.Point())
	//require.NoError(t, err)
}

func TestPedersen_GetPublicKey(t *testing.T) {

	actor := Actor{handler: &Handler{startRes: &state{}}}

	// GetPublicKey requires Setup to have been run
	_, err := actor.GetPublicKey()
	require.EqualError(t, err, "dkg has not been initialized")

	actor.handler.startRes = &state{participants: []mino.Address{fake.NewAddress(0)}, distKey: suite.Point()}

	_, err = actor.GetPublicKey()
	require.NoError(t, err)
}

func TestPedersen_Scenario(t *testing.T) {
	n := 5

	minos := make([]mino.Mino, n)
	addrs := make([]mino.Address, n)
	dkgs := make([]dkg.DKG, n)
	actors := make([]dkg.Actor, n)

	electionID := "deadbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	for i := 0; i < n; i++ {
		addr := minogrpc.ParseAddress("127.0.0.1", 0)

		minogrpc, err := minogrpc.NewMinogrpc(addr, tree.NewRouter(minogrpc.NewAddressFactory()))
		require.NoError(t, err)

		defer minogrpc.GracefulStop()

		minos[i] = minogrpc
		addrs[i] = minogrpc.GetAddress()
	}

	for _, mino := range minos {
		// share the certificates
		joinable, ok := mino.(minogrpc.Joinable)
		require.True(t, ok)

		addrStr := mino.GetAddress().String()
		token := joinable.GenerateToken(time.Hour)

		certHash, err := joinable.GetCertificateStore().Hash(joinable.GetCertificate())
		require.NoError(t, err)

		for _, n := range minos {
			otherJoinable, ok := n.(minogrpc.Joinable)
			require.True(t, ok)

			err = otherJoinable.Join(addrStr, token, certHash)
			require.NoError(t, err)
		}
	}

	roster := authority.FromAuthority(fake.NewAuthorityFromMino(fake.NewSigner, minos...))

	election := fake.NewElection(electionID)
	election.Roster = roster

	service := fake.NewService(electionID, election, serdecontext)

	for i, mino := range minos {
		fac := etypes.NewElectionFactory(etypes.CiphervoteFactory{}, fake.NewRosterFac(roster))

		dkg := NewPedersen(mino, service, &fake.Pool{}, fac, fake.Signer{})

		actor, err := dkg.Listen(electionIDBuf, signed.NewManager(fake.Signer{}, &client{}))
		require.NoError(t, err)

		dkgs[i] = dkg
		actors[i] = actor
	}

	// trying to call a decrypt/encrypt before a setup
	_, _, _, err = actors[0].Encrypt(nil)
	require.EqualError(t, err, "setup() was not called")

	//_, err = actors[0].Decrypt(nil, nil)
	//require.EqualError(t, err, "setup() was not called")

	pubKey, err := actors[0].Setup()
	require.NoError(t, err)

	// number of votes
	k := 1

	message := "Hello world"

	Ks, Cs, _ := fakeKCPoints(k, message, pubKey)

	for i := 0; i < k; i++ {
		ballot := etypes.Ciphervote{etypes.EGPair{
			K: Ks[i],
			C: Cs[i],
		}}
		election.Suffragia.CastVote("dummyUser"+strconv.Itoa(i), ballot)
	}

	shuffledBallots := election.Suffragia.Ciphervotes
	shuffleInstance := etypes.ShuffleInstance{ShuffledBallots: shuffledBallots}
	election.ShuffleInstances = append(election.ShuffleInstances, shuffleInstance)

	election.ShuffleThreshold = 1

	service.Elections[electionID] = election

	_, err = actors[0].Setup()
	require.EqualError(t, err, "setup() was already called, only one call is allowed")

	// every node should be able to decrypt

	//ks, cs := shuffledBallots[0].GetElGPairs()
	//
	//for _, actor := range actors {
	//	decrypted, err := actor.Decrypt(ks[0], cs[0])
	//	require.NoError(t, err)
	//	require.Equal(t, message, string(decrypted))
	//}
}

// -----------------------------------------------------------------------------
// Utility functions

// actorsEqual checks that two actors hold the same data
func requireActorsEqual(t require.TestingT, actor1, actor2 dkg.Actor) {
	actor1Data, err := actor1.MarshalJSON()
	require.NoError(t, err)
	actor2Data, err := actor2.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, actor1Data, actor2Data)
}

func fakeKCPoints(k int, msg string, pubKey kyber.Point) ([]kyber.Point, []kyber.Point, kyber.Point) {
	Ks := make([]kyber.Point, 0, k)
	Cs := make([]kyber.Point, 0, k)

	for i := 0; i < k; i++ {
		// Embed the message into a curve point
		M := suite.Point().Embed([]byte(msg), random.New())

		// ElGamal-encrypt the point to produce ciphertext (K,C).
		k := suite.Scalar().Pick(random.New()) // ephemeral private key
		K := suite.Point().Mul(k, nil)         // ephemeral DH public key
		S := suite.Point().Mul(k, pubKey)      // ephemeral DH shared secret
		C := S.Add(S, M)                       // message blinded with secret

		Ks = append(Ks, K)
		Cs = append(Cs, C)
	}
	return Ks, Cs, pubKey
}

// client fetches the last nonce used by the client
//
// - implements signed.Client
type client struct {
	srvc ordering.Service
	vs   validation.Service
}

// GetNonce implements signed.Client. It uses the validation service to get the
// last nonce.
func (c *client) GetNonce(id access.Identity) (uint64, error) {
	store := c.srvc.GetStore()

	nonce, err := c.vs.GetNonce(store, id)
	if err != nil {
		return 0, xerrors.Errorf("failed to get nonce from validation: %v", err)
	}

	return nonce, nil
}
