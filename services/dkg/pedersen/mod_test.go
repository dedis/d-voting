package pedersen

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/dkg/pedersen/types"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/core/store/kv"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/mino/minogrpc"
	"go.dedis.ch/dela/mino/router/tree"
)

// If you get the persistent data from an actor and then recreate an actor
// from that data, the persistent data should be the same in both actors.
func TestActor_MarshalJSON(t *testing.T) {
	p := NewPedersen(fake.Mino{}, fake.Service{}, fake.Factory{})

	// Create new actor
	actor1, err := p.NewActor([]byte("deadbeef"), HandlerDataTest())
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

	electionActorMap := map[string]HandlerData{
		"deadbeef51": NewHandlerData(),
		"deadbeef52": NewHandlerData(),
		"deadbeef53": NewHandlerData(),
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

			handlerDataBuf, err := json.Marshal(handlerData)
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

			print(handlerDataBuf)

			handlerData := HandlerData{}
			err = json.Unmarshal(handlerDataBuf, &handlerData)
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
		require.Equal(t, exists, true)

		actorDataBuf, err := actor.MarshalJSON()
		require.NoError(t, err)

		handlerDataBuf, err := json.Marshal(handlerData)
		require.NoError(t, err)

		// Check that each field is the same
		require.Equal(t, handlerDataBuf, actorDataBuf)
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
	p := NewPedersen(fake.Mino{}, fake.Service{}, fake.Factory{})

	electionID := "d3adbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	actor, err := p.Listen(electionIDBuf)
	require.NoError(t, err)

	require.NotNil(t, actor)
}

// If Listen is called twice for the same election, the actor data is unchanged
func TestPedersen_TwoListens(t *testing.T) {
	p := NewPedersen(fake.Mino{}, fake.Service{}, fake.Factory{})

	electionID := "deadbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	actor1, err := p.Listen(electionIDBuf)
	require.NoError(t, err)

	actor2, err := p.Listen(electionIDBuf)
	require.NoError(t, err)

	require.Equal(t, actor1, actor2)
}

func TestPedersen_Setup(t *testing.T) {
	actor := Actor{
		rpc: fake.NewBadRPC(),
		handler: &Handler{
			startRes: &state{},
		},
	}

	_, err := actor.Setup()
	require.EqualError(t, err, fake.Err("failed to stream"))

	rpc := fake.NewStreamRPC(fake.NewReceiver(), fake.NewBadSender())
	actor.rpc = rpc

	_, err = actor.Setup()
	require.EqualError(t, err, "expected ed25519.PublicKey, got 'fake.PublicKey'")

	rpc = fake.NewStreamRPC(fake.NewBadReceiver(), fake.Sender{})
	actor.rpc = rpc

	_, err = actor.Setup()
	require.EqualError(t, err, fake.Err("got an error from '%!s(<nil>)' while receiving"))

	recv := fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), nil))

	rpc = fake.NewStreamRPC(recv, fake.Sender{})
	actor.rpc = rpc

	_, err = actor.Setup()
	require.EqualError(t, err, "expected to receive a Done message, but go the following: <nil>")

	rpc = fake.NewStreamRPC(fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), types.NewStartDone(suite.Point())),
		fake.NewRecvMsg(fake.NewAddress(0), types.NewStartDone(suite.Point().Pick(suite.RandomStream()))),
	), fake.Sender{})
	actor.rpc = rpc

	_, err = actor.Setup()
	require.Error(t, err)
	require.Regexp(t, "^the public keys does not match:", err)
}

func TestPedersen_GetPublicKey(t *testing.T) {

	p := NewPedersen(fake.Mino{}, fake.Service{}, fake.Factory{})

	electionID := "deadbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	actor, err := p.Listen(electionIDBuf)
	require.NoError(t, err)

	_, err = actor.GetPublicKey()
	require.EqualError(t, err, "DKG has not been initialized")

	actor.Setup()

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
	dkgs := make([]dkg.DKG, n)
	addrs := make([]mino.Address, n)

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

		dkg := NewPedersen(mino.(*minogrpc.Minogrpc), fake.Service{}, fake.Factory{})

		dkgs[i] = dkg
	}

	electionIDBuf, err := hex.DecodeString("deadbeef")
	require.NoError(t, err)

	message := []byte("Hello world")
	actors := make([]dkg.Actor, n)
	for i := 0; i < n; i++ {
		actor, err := dkgs[i].Listen(electionIDBuf)
		require.NoError(t, err)

		actors[i] = actor
	}

	// trying to call a decrypt/encrypt before a setup
	_, _, _, err = actors[0].Encrypt(message)
	require.EqualError(t, err, "you must first initialize DKG. Did you call setup() first?")
	_, err = actors[0].Decrypt(nil, nil)
	require.EqualError(t, err, "you must first initialize DKG. Did you call setup() first?")

	_, err = actors[0].Setup()
	require.NoError(t, err)

	_, err = actors[0].Setup()
	require.EqualError(t, err, "startRes is already done, only one setup call is allowed")

	// every node should be able to encrypt/decrypt
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
