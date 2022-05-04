package controller

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/core/validation"

	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/dkg/pedersen"
	"github.com/gorilla/mux"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/store/kv"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/mino/proxy"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"

	eproxy "github.com/dedis/d-voting/proxy"
)

var suite = suites.MustFind("Ed25519")

// initAction is an action to initialize the DKG protocol
//
// - implements node.ActionTemplate
type initAction struct {
}

// Execute implements node.ActionTemplate. It creates an actor from
// the dkgPedersen instance and links it to an election.
func (a *initAction) Execute(ctx node.Context) error {

	electionID := ctx.Flags.String("electionID")

	electionIDBuf, err := hex.DecodeString(electionID)
	if err != nil {
		return xerrors.Errorf("failed to decode electionID: %v", err)
	}

	// Initialize the actor
	var dkg dkg.DKG
	err = ctx.Injector.Resolve(&dkg)
	if err != nil {
		return xerrors.Errorf("failed to resolve DKG: %v", err)
	}

	_, exists := dkg.GetActor(electionIDBuf)
	if exists {
		return xerrors.Errorf("DKG was already initialized for electionID %s", electionID)
	}

	signer, err := getSigner(ctx.Flags)
	if err != nil {
		return xerrors.Errorf("failed to get signer: %v", err)
	}

	client, err := makeClient(ctx.Injector)
	if err != nil {
		return xerrors.Errorf("failed to make client: %v", err)
	}

	actor, err := dkg.Listen(electionIDBuf, signed.NewManager(signer, &client))
	if err != nil {
		return xerrors.Errorf("failed to start the RPC: %v", err)
	}

	dela.Logger.Info().Msg("DKG has been initialized successfully")

	err = updateDKGStore(ctx.Injector, func(tx kv.WritableTx) error {
		bucket, err := tx.GetBucketOrCreate([]byte(BucketName))
		if err != nil {
			return err
		}

		actorBuf, err := actor.MarshalJSON()
		if err != nil {
			return err
		}

		return bucket.Set(electionIDBuf, actorBuf)
	})
	if err != nil {
		return xerrors.Errorf("failed to update DKG store: %v", err)
	}

	dela.Logger.Info().Msgf("DKG was successfully linked to election %v", electionIDBuf)

	return nil
}

// setupAction is an action to setup the DKG protocol and generate a collective
// public key
//
// - implements node.ActionTemplate
type setupAction struct {
}

// Execute implements node.ActionTemplate. It reads the list of members and
// request the setup.
func (a *setupAction) Execute(ctx node.Context) error {

	electionIDBuf, err := hex.DecodeString(ctx.Flags.String("electionID"))
	if err != nil {
		return xerrors.Errorf("failed to decode electionID: %v", err)
	}

	var dkg dkg.DKG
	err = ctx.Injector.Resolve(&dkg)
	if err != nil {
		return xerrors.Errorf("failed to resolve DKG: %v", err)
	}

	actor, exists := dkg.GetActor(electionIDBuf)
	if !exists {
		return xerrors.Errorf("failed to get actor: %v", err)
	}

	pubkey, err := actor.Setup()
	if err != nil {
		return xerrors.Errorf("failed to setup DKG: %v", err)
	}

	pubkeyBuf, err := pubkey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to encode pubkey: %v", err)
	}

	dela.Logger.Info().
		Hex("DKG public key", pubkeyBuf).
		Msg("DKG public key")

	err = updateDKGStore(ctx.Injector, func(tx kv.WritableTx) error {
		bucket, err := tx.GetBucketOrCreate([]byte(BucketName))
		if err != nil {
			return err
		}

		actorBuf, err := actor.MarshalJSON()
		if err != nil {
			return err
		}

		return bucket.Set(electionIDBuf, actorBuf)
	})
	if err != nil {
		return xerrors.Errorf("failed to update DKG store: %v", err)
	}

	return nil
}

// exportInfoAction is an action to display a base64 string describing the node.
// It can be used to transmit the identity of a node to another one.
//
// - implements node.ActionTemplate
type exportInfoAction struct {
}

// Execute implements node.ActionTemplate. It looks for the node address and
// public key and prints "$ADDR_BASE64:$PUBLIC_KEY_BASE64".
func (a *exportInfoAction) Execute(ctx node.Context) error {
	var m mino.Mino
	err := ctx.Injector.Resolve(&m)
	if err != nil {
		return xerrors.Errorf("injector: %v", err)
	}

	addr, err := m.GetAddress().MarshalText()
	if err != nil {
		return xerrors.Errorf("failed to marshal address: %v", err)
	}

	desc := base64.StdEncoding.EncodeToString(addr)

	// Print address
	fmt.Fprint(ctx.Out, desc)

	var db kv.DB
	err = ctx.Injector.Resolve(&db)
	if err != nil {
		return xerrors.Errorf("injector: %v", err)
	}

	err = db.View(func(tx kv.ReadableTx) error {
		bucket := tx.GetBucket([]byte(BucketName))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(electionIDBuf, handlerDataBuf []byte) error {

			handlerData := pedersen.HandlerData{}
			err = json.Unmarshal(handlerDataBuf, &handlerData)
			if err != nil {
				return err
			}

			// Print electionID and actor data
			fmt.Fprint(ctx.Out, hex.EncodeToString(electionIDBuf))
			fmt.Fprint(ctx.Out, handlerData)

			return nil
		})
	})
	if err != nil {
		return xerrors.Errorf("database read failed: %v", err)
	}

	return nil
}

// Ciphertext wraps the ciphertext pairs
type Ciphertext struct {
	K []byte
	C []byte
}

// getPublicKeyAction is an action that prints the collective public key
//
// - implements node.ActionTemplate
type getPublicKeyAction struct {
}

// Execute implements node.ActionTemplate. It retrieves the collective
// public key from the DKG service and prints it.
func (a *getPublicKeyAction) Execute(ctx node.Context) error {
	electionIDBuf, err := hex.DecodeString(ctx.Flags.String("electionID"))
	if err != nil {
		return xerrors.Errorf("failed to decode electionID: %v", err)
	}

	var dkgPedersen dkg.DKG
	err = ctx.Injector.Resolve(&dkgPedersen)
	if err != nil {
		return xerrors.Errorf("failed to resolve dkg: %v", err)
	}

	actor, exists := dkgPedersen.GetActor(electionIDBuf)
	if !exists {
		return xerrors.Errorf("failed to get actor: %v", err)
	}

	pubkey, err := actor.GetPublicKey()
	if err != nil {
		return xerrors.Errorf("failed to retrieve the public key: %v", err)
	}

	pubkeyBuf, err := pubkey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to encode pubkey: %v", err)
	}

	dela.Logger.Info().
		Hex("DKG public key", pubkeyBuf).
		Msg("DKG public key")

	return nil
}

// RegisterHandlersAction is an action that registers the proxy handlers
//
// - implements node.ActionTemplate
type RegisterHandlersAction struct {
}

// Execute implements node.ActionTemplate. It registers the proxy
// handlers to set up elections
func (a *RegisterHandlersAction) Execute(ctx node.Context) error {
	var proxy proxy.Proxy
	err := ctx.Injector.Resolve(&proxy)
	if err != nil {
		return xerrors.Errorf("failed to resolve proxy: %v", err)
	}

	var dkg dkg.DKG
	err = ctx.Injector.Resolve(&dkg)
	if err != nil {
		return xerrors.Errorf("failed to resolve dkg.DKG: %v", err)
	}

	signer, err := getSigner(ctx.Flags)
	if err != nil {
		return xerrors.Errorf("failed to get signer for txmngr : %v", err)
	}

	client, err := makeClient(ctx.Injector)
	if err != nil {
		return xerrors.Errorf("failed to make client: %v", err)
	}

	mngr := signed.NewManager(signer, &client)

	proxykeyHex := ctx.Flags.String("proxykey")

	proxykeyBuf, err := hex.DecodeString(proxykeyHex)
	if err != nil {
		return xerrors.Errorf("failed to decode proxykeyHex: %v", err)
	}

	proxykey := suite.Point()

	err = proxykey.UnmarshalBinary(proxykeyBuf)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal proxy key: %v", err)
	}

	router := mux.NewRouter()

	ep := eproxy.NewDKG(mngr, dkg, proxykey)

	router.HandleFunc("/evoting/services/dkg/actors", ep.NewDKGActor).Methods("POST")
	router.HandleFunc("/evoting/services/dkg/actors/{electionID}", ep.Actor).Methods("GET")
	router.HandleFunc("/evoting/services/dkg/actors/{electionID}", ep.EditDKGActor).Methods("PUT")
	router.HandleFunc("/evoting/services/dkg/actors/{electionID}", eproxy.AllowCORS).Methods("OPTIONS")

	router.NotFoundHandler = http.HandlerFunc(eproxy.NotFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(eproxy.NotAllowedHandler)

	proxy.RegisterHandler("/evoting/services/dkg/", router.ServeHTTP)

	dela.Logger.Info().Msg("DKG handler registered")

	return nil
}

func updateDKGStore(inj node.Injector, fn func(kv.WritableTx) error) error {
	var db kv.DB
	err := inj.Resolve(&db)
	if err != nil {
		return xerrors.Errorf("failed to resolve db: %v", err)
	}

	err = db.Update(fn)
	if err != nil {
		return xerrors.Errorf("failed to update: %v", err)
	}

	return nil
}

func makeClient(inj node.Injector) (client, error) {
	var service ordering.Service
	err := inj.Resolve(&service)
	if err != nil {
		return client{}, xerrors.Errorf("failed to resolve ordering.Service: %v", err)
	}

	var vs validation.Service
	err = inj.Resolve(&vs)
	if err != nil {
		return client{}, xerrors.Errorf("failed to resolve validation.Service: %v", err)
	}

	client := client{
		srvc: service,
		vs:   vs,
	}

	return client, nil
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
