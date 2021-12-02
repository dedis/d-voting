package pedersen

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	// "github.com/dedis/d-voting/services/dkg"
	// "github.com/dedis/d-voting/services/dkg/pedersen/types"
	"github.com/stretchr/testify/require"
	// "go.dedis.ch/dela/crypto"
	// "go.dedis.ch/dela/crypto/ed25519"
	"go.dedis.ch/dela/core/store/kv"
	"go.dedis.ch/dela/mino"
	// "go.dedis.ch/dela/mino/minogrpc"
	// "go.dedis.ch/dela/mino/router/tree"
	// "go.dedis.ch/kyber/v3"
)

// // After initializing a Pedersen when dkgMap is empty, the actors map should be empty
// func TestPedersen_InitEmptyMap(t *testing.T) {

// }

// // After initializing a Pedersen when dkgMap is not empty, the actors map should contain the same information
// // as dkgMap
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

		actor, err := p.GetActor(electionIDBuf)
		require.NoError(t, err)

		actorDataBuf, err := actor.MarshalJSON()
		require.NoError(t, err)

		handlerDataBuf, err := json.Marshal(handlerData)
		require.NoError(t, err)

		// Check that each field is the same
		require.Equal(t, handlerDataBuf, actorDataBuf)
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

// If you get the persistent data from an actor and then recreate an actor from that data,
// the persistent data should be the same in both actors.
func TestActor_MarshalJSON(t *testing.T) {
	p := NewPedersen(fake.Mino{}, fake.Service{}, fake.Factory{})

	// Create new actor
	electionID := "deadbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	actor, err := p.Listen(electionIDBuf)
	require.NoError(t, err)

	// Serialize its persistent data
	actorBuf, err := actor.MarshalJSON()
	require.NoError(t, err)

	// Create a new actor with that data
	electionID = "beefdead"
	electionIDBuf, err = hex.DecodeString(electionID)
	require.NoError(t, err)

	actorData := HandlerData{}
	err = json.Unmarshal(actorBuf, &actorData)
	require.NoError(t, err)

	newActor, err := p.NewActor(electionIDBuf, actorData)
	require.NoError(t, err)

	// Check that the persistent data is the same for
	// both actors
	newActorBuf, err := newActor.MarshalJSON()
	require.NoError(t, err)

	print(newActorBuf)

	require.Equal(t, actorBuf, newActorBuf)
}

// func TestPedersen_Setup(t *testing.T) {
//         actor := Actor{
//                 rpc:      fake.NewBadRPC(),
//                 startRes: &state{},
//         }

//         fakeAuthority := fake.NewAuthority(1, fake.NewSigner)

//         _, err := actor.Setup(fakeAuthority, 0)
//         require.EqualError(t, err, fake.Err("failed to stream"))

//         rpc := fake.NewStreamRPC(fake.NewReceiver(), fake.NewBadSender())
//         actor.rpc = rpc

//         _, err = actor.Setup(fakeAuthority, 0)
//         require.EqualError(t, err, "expected ed25519.PublicKey, got 'fake.PublicKey'")

//         rpc = fake.NewStreamRPC(fake.NewBadReceiver(), fake.Sender{})
//         actor.rpc = rpc

//         fakeAuthority = fake.NewAuthority(2, ed25519.NewSigner)

//         _, err = actor.Setup(fakeAuthority, 1)
//         require.EqualError(t, err, fake.Err("got an error from '%!s(<nil>)' while receiving"))

//         recv := fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), nil))

//         rpc = fake.NewStreamRPC(recv, fake.Sender{})
//         actor.rpc = rpc

//         _, err = actor.Setup(fakeAuthority, 1)
//         require.EqualError(t, err, "expected to receive a Done message, but go the following: <nil>")

//         rpc = fake.NewStreamRPC(fake.NewReceiver(
//                 fake.NewRecvMsg(fake.NewAddress(0), types.NewStartDone(suite.Point())),
//                 fake.NewRecvMsg(fake.NewAddress(0), types.NewStartDone(suite.Point().Pick(suite.RandomStream()))),
//         ), fake.Sender{})
//         actor.rpc = rpc

//         _, err = actor.Setup(fakeAuthority, 1)
//         require.Error(t, err)
//         require.Regexp(t, "^the public keys does not match:", err)
// }

func TestPedersen_GetPublicKey(t *testing.T) {

	electionID := "deadbeef"
	electionIDBuf, err := hex.DecodeString(electionID)
	require.NoError(t, err)

	actor := Actor{
		handler: NewHandler(
			fake.Mino{}.GetAddress(),
			fake.Service{},
			NewHandlerData(),
		),
		service: fake.Service{ElectionID: electionID},
		rosterFac: fake.Factory{},
		rpc: fake.NewRPC(),
	}

	actor.handler.startRes = &state{
		participants: []mino.Address{fake.NewAddress(0)},
		distrKey:     nil,
	}

	_, err = actor.GetPublicKey()
	require.EqualError(t, err, "DKG has not been initialized")

	actor.Setup()

	_, err = actor.GetPublicKey()
	require.NoError(t, err)
}

// func TestPedersen_Decrypt(t *testing.T) {
//         actor := Actor{
//                 rpc:      fake.NewBadRPC(),
//                 startRes: &state{participants: []mino.Address{fake.NewAddress(0)}, distrKey: suite.Point()},
//         }

//         _, err := actor.Decrypt(suite.Point(), suite.Point(), "electionId")
//         require.EqualError(t, err, fake.Err("failed to create stream"))

//         rpc := fake.NewStreamRPC(fake.NewBadReceiver(), fake.NewBadSender())
//         actor.rpc = rpc

//         _, err = actor.Decrypt(suite.Point(), suite.Point(), "electionId")
//         require.EqualError(t, err, fake.Err("failed to send decrypt request"))

//         recv := fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), nil))

//         rpc = fake.NewStreamRPC(recv, fake.Sender{})
//         actor.rpc = rpc

//         _, err = actor.Decrypt(suite.Point(), suite.Point(), "electionId")
//         require.EqualError(t, err, "got unexpected reply, expected types.DecryptReply but got: <nil>")

//         recv = fake.NewReceiver(
//                 fake.NewRecvMsg(fake.NewAddress(0), types.DecryptReply{I: -1, V: suite.Point()}),
//         )

//         rpc = fake.NewStreamRPC(recv, fake.Sender{})
//         actor.rpc = rpc

//         _, err = actor.Decrypt(suite.Point(), suite.Point(), "electionId")
//         require.EqualError(t, err, "failed to recover commit: share: not enough "+
//                 "good public shares to reconstruct secret commitment")

//         recv = fake.NewReceiver(
//                 fake.NewRecvMsg(fake.NewAddress(0), types.DecryptReply{I: 1, V: suite.Point()}),
//         )

//         rpc = fake.NewStreamRPC(recv, fake.Sender{})
//         actor.rpc = rpc

//         _, err = actor.Decrypt(suite.Point(), suite.Point(), "electionId")
//         require.NoError(t, err)
// }

// func TestPedersen_Reshare(t *testing.T) {
//         actor := Actor{}
//         actor.Reshare()
// }

// func TestPedersen_Scenario(t *testing.T) {
//         // Use with MINO_TRAFFIC=log
//         // traffic.LogItems = false
//         // traffic.LogEvent = false
//         // defer func() {
//         // 	traffic.SaveItems("graph.dot", true, false)
//         // 	traffic.SaveEvents("events.dot")
//         // }()

//         n := 5

//         minos := make([]mino.Mino, n)
//         dkgs := make([]dkg.DKG, n)
//         addrs := make([]mino.Address, n)

//         for i := 0; i < n; i++ {
//                 addr := minogrpc.ParseAddress("127.0.0.1", 0)

//                 minogrpc, err := minogrpc.NewMinogrpc(addr, tree.NewRouter(minogrpc.NewAddressFactory()))
//                 require.NoError(t, err)

//                 defer minogrpc.GracefulStop()

//                 minos[i] = minogrpc
//                 addrs[i] = minogrpc.GetAddress()
//         }

//         pubkeys := make([]kyber.Point, len(minos))

//         for i, mino := range minos {
//                 for _, m := range minos {
//                         mino.(*minogrpc.Minogrpc).GetCertificateStore().Store(m.GetAddress(), m.(*minogrpc.Minogrpc).GetCertificate())
//                 }

//                 dkg, pubkey := NewPedersen(mino.(*minogrpc.Minogrpc), false)

//                 dkgs[i] = dkg
//                 pubkeys[i] = pubkey
//         }

//         fakeAuthority := NewAuthority(addrs, pubkeys)

//         message := []byte("Hello world")
//         actors := make([]dkg.Actor, n)
//         for i := 0; i < n; i++ {
//                 actor, err := dkgs[i].Listen()
//                 require.NoError(t, err)

//                 actors[i] = actor
//         }

//         // trying to call a decrypt/encrypt before a setup
//         _, _, _, err := actors[0].Encrypt(message)
//         require.EqualError(t, err, "you must first initialize DKG. Did you call setup() first?")
//         _, err = actors[0].Decrypt(nil, nil, "electionId")
//         require.EqualError(t, err, "you must first initialize DKG. Did you call setup() first?")

//         _, err = actors[0].Setup(fakeAuthority, n)
//         require.NoError(t, err)

//         _, err = actors[0].Setup(fakeAuthority, n)
//         require.EqualError(t, err, "startRes is already done, only one setup call is allowed")

//         // every node should be able to encrypt/decrypt
//         for i := 0; i < n; i++ {
//                 K, C, remainder, err := actors[i].Encrypt(message)
//                 require.NoError(t, err)
//                 require.Len(t, remainder, 0)
//                 decrypted, err := actors[i].Decrypt(K, C, "electionId")
//                 require.NoError(t, err)
//                 require.Equal(t, message, decrypted)
//         }
// }

// // -----------------------------------------------------------------------------
// // Utility functions

// //
// // Collective authority
// //

// // CollectiveAuthority is a fake implementation of the cosi.CollectiveAuthority
// // interface.
// type CollectiveAuthority struct {
//         crypto.CollectiveAuthority
//         addrs   []mino.Address
//         pubkeys []kyber.Point
//         signers []crypto.Signer
// }

// // NewAuthority returns a new collective authority of n members with new signers
// // generated by g.
// func NewAuthority(addrs []mino.Address, pubkeys []kyber.Point) CollectiveAuthority {
//         signers := make([]crypto.Signer, len(pubkeys))
//         for i, pubkey := range pubkeys {
//                 signers[i] = newFakeSigner(pubkey)
//         }

//         return CollectiveAuthority{
//                 pubkeys: pubkeys,
//                 addrs:   addrs,
//                 signers: signers,
//         }
// }

// // GetPublicKey implements cosi.CollectiveAuthority.
// func (ca CollectiveAuthority) GetPublicKey(addr mino.Address) (crypto.PublicKey, int) {

//         for i, address := range ca.addrs {
//                 if address.Equal(addr) {
//                         return ed25519.NewPublicKeyFromPoint(ca.pubkeys[i]), i
//                 }
//         }
//         return nil, -1
// }

// // Len implements mino.Players.
// func (ca CollectiveAuthority) Len() int {
//         return len(ca.pubkeys)
// }

// // AddressIterator implements mino.Players.
// func (ca CollectiveAuthority) AddressIterator() mino.AddressIterator {
//         return fake.NewAddressIterator(ca.addrs)
// }

// func (ca CollectiveAuthority) PublicKeyIterator() crypto.PublicKeyIterator {
//         return fake.NewPublicKeyIterator(ca.signers)
// }

// func newFakeSigner(pubkey kyber.Point) fakeSigner {
//         return fakeSigner{
//                 pubkey: pubkey,
//         }
// }

// // fakeSigner is a fake signer
// //
// // - implements crypto.Signer
// type fakeSigner struct {
//         crypto.Signer
//         pubkey kyber.Point
// }

// // GetPublicKey implements crypto.Signer
// func (s fakeSigner) GetPublicKey() crypto.PublicKey {
//         return ed25519.NewPublicKeyFromPoint(s.pubkey)
// }
