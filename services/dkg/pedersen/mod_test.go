package pedersen

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	electionTypes "github.com/dedis/d-voting/contracts/evoting/types"
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
	"go.dedis.ch/kyber/v3/share"
)

// If you get the persistent data from an actor and then recreate an actor
// from that data, the persistent data should be the same in both actors.
func TestActor_MarshalJSON(t *testing.T) {
	p := NewPedersen(fake.Mino{}, fake.Service{}, fake.Factory{})

	// Create new actor
	actor1, err := p.NewActor([]byte("deadbeef"), NewHandlerData())
	require.NoError(t, err)

	// Serialize its persistent data
	actor1Buf, err := actor1.MarshalJSON()
	require.NoError(t, err)

	// Create a new actor with that data
	handlerData := HandlerData{}
	err = handlerData.UnmarshalJSON(actor1Buf)
	require.NoError(t, err)

	actor2, err := p.NewActor([]byte("beefdead"), handlerData)
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
	p := NewPedersen(fake.Mino{}, fake.Service{}, fake.Factory{})

	err = dkgMap.View(func(tx kv.ReadableTx) error {
		bucket := tx.GetBucket([]byte("dkgmap"))
		require.NotNil(t, bucket)

		return bucket.ForEach(func(electionIDBuf, handlerDataBuf []byte) error {

			handlerData := HandlerData{}
			err = handlerData.UnmarshalJSON(handlerDataBuf)
			if err != nil {
				return err
			}

			_, err = p.NewActor(electionIDBuf, handlerData)
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
			handler: NewHandler(fake.NewAddress(0), fake.Service{}, handlerData),
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
	p := NewPedersen(fake.Mino{}, fake.Service{}, fake.Factory{})

	// Create actors
	a1, err := p.NewActor([]byte(electionID1), NewHandlerData())
	require.NoError(t, err)
	_, err = p.NewActor([]byte(electionID2), NewHandlerData())
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
	q := NewPedersen(fake.Mino{}, fake.Service{}, fake.Factory{})

	err = dkgMap.View(func(tx kv.ReadableTx) error {
		bucket := tx.GetBucket([]byte("dkgmap"))
		require.NotNil(t, bucket)

		return bucket.ForEach(func(electionIDBuf, handlerDataBuf []byte) error {

			handlerData := HandlerData{}
			err = json.Unmarshal(handlerDataBuf, &handlerData)
			require.NoError(t, err)

			_, err = q.NewActor(electionIDBuf, handlerData)
			require.NoError(t, err)

			return nil
		})
	})
	require.NoError(t, err)

	// Check the information is conserved

	require.Equal(t, len(q.actors), len(p.actors))

	// Check equality of actor data
	for electionID, actor_q := range q.actors {

		electionIDBuf, err := hex.DecodeString(electionID)
		require.NoError(t, err)

		actor_p, exists := p.GetActor(electionIDBuf)
		require.True(t, exists)

		requireActorsEqual(t, actor_p, actor_q)
	}
}

func TestPedersen_Listen(t *testing.T) {
	electionID := "d3adbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	p := NewPedersen(fake.Mino{}, fake.NewService(electionID, electionTypes.Election{}), fake.Factory{})

	actor, err := p.Listen(electionIDBuf)
	require.NoError(t, err)

	require.NotNil(t, actor)
}

// If Listen is called twice for the same election, the actor data is unchanged
func TestPedersen_TwoListens(t *testing.T) {
	electionID := "deadbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	p := NewPedersen(fake.Mino{}, fake.NewService(electionID, electionTypes.Election{}), fake.Factory{})

	actor1, err := p.Listen(electionIDBuf)
	require.NoError(t, err)

	actor2, err := p.Listen(electionIDBuf)
	require.Error(t, err, "actor already exists for electionID deadbeef")

	require.Equal(t, actor1, actor2)
}

func TestPedersen_Setup(t *testing.T) {
	electionID := "d3adbeef"

	actor := Actor{
		rpc:       nil,
		factory:   nil,
		service:   fake.NewService(electionID, electionTypes.Election{}),
		rosterFac: fake.Factory{},
		handler: &Handler{
			startRes: &state{},
		},
	}

	// Wrong electionID
	wrongElectionID := "beefdead"
	actor.electionID = wrongElectionID

	_, err := actor.Setup()
	require.EqualError(t, err, fmt.Sprintf("election %s was not found", wrongElectionID))

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
	actor.rosterFac = fake.NewRosterFac(roster)

	rosterBuf, err := roster.Serialize(fake.NewContextWithFormat(serde.Format("JSON")))
	require.NoError(t, err)

	actor.service = fake.NewService(
		electionID,
		electionTypes.Election{
			RosterBuf: rosterBuf,
		},
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

func TestPedersen_GetPublicKey(t *testing.T) {

	actor := Actor{handler: &Handler{startRes: &state{}}}

	// GetPublicKey requires Setup to have been run
	_, err := actor.GetPublicKey()
	require.EqualError(t, err, "dkg has not been initialized")

	actor.handler.startRes = &state{participants: []mino.Address{fake.NewAddress(0)}, distKey: suite.Point()}

	_, err = actor.GetPublicKey()
	require.NoError(t, err)
}

func TestPedersen_Reshare(t *testing.T) {
	actor := Actor{}
	actor.Reshare()
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
	electionService := fake.NewService(electionID, electionTypes.Election{})

	for i := 0; i < n; i++ {
		addr := minogrpc.ParseAddress("127.0.0.1", 0)

		minogrpc, err := minogrpc.NewMinogrpc(addr, tree.NewRouter(minogrpc.NewAddressFactory()))
		require.NoError(t, err)

		defer minogrpc.GracefulStop()

		minos[i] = minogrpc
		addrs[i] = minogrpc.GetAddress()
	}

	for i, mino := range minos {
		for _, m := range minos {
			mino.(*minogrpc.Minogrpc).GetCertificateStore().Store(m.GetAddress(), m.(*minogrpc.Minogrpc).GetCertificate())
		}

		dkgs[i] = NewPedersen(mino, electionService, fake.Factory{})
	}

	for i, dkg := range dkgs {
		actor, err := dkg.Listen(electionIDBuf)
		require.NoError(t, err)

		actors[i] = actor
	}

	// trying to call a decrypt/encrypt before a setup
	_, _, _, err = actors[0].Encrypt(nil)
	require.EqualError(t, err, "setup() was not called")

	_, err = actors[0].Decrypt(nil, nil)
	require.EqualError(t, err, "setup() was not called")

	_, err = actors[0].Setup()
	require.NoError(t, err)

	_, err = actors[0].Setup()
	require.EqualError(t, err, "startRes is already done, only one setup call is allowed")

	// every node should be able to encrypt/decrypt
	message := []byte("Hello world")
	for i := 0; i < n; i++ {
		K, C, remainder, err := actors[i].Encrypt(message)
		require.NoError(t, err)
		require.Len(t, remainder, 0)
		decrypted, err := actors[i].Decrypt(K, C)
		require.NoError(t, err)
		require.Equal(t, message, decrypted)
	}
}

// actorsEqual checks that two actors hold the same data
func requireActorsEqual(t require.TestingT, actor1, actor2 dkg.Actor) {
	actor1Data, err := actor1.MarshalJSON()
	require.NoError(t, err)
	actor2Data, err := actor2.MarshalJSON()
	require.NoError(t, err)

	require.Equal(t, actor1Data, actor2Data)
}
