package controller

import (
	"encoding/hex"
	"net/http"

	"github.com/dedis/d-voting/services/shuffle"
	"github.com/gorilla/mux"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/core/validation"
	"go.dedis.ch/dela/mino/proxy"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"

	eproxy "github.com/dedis/d-voting/proxy"
)

var suite = suites.MustFind("ed25519")

// InitAction is an action to initialize the shuffle protocol
//
// - implements node.ActionTemplate
type InitAction struct {
}

// Execute implements node.ActionTemplate. It creates an actor from
// the neffShuffle instance
func (a *InitAction) Execute(ctx node.Context) error {
	var neffShuffle shuffle.Shuffle

	err := ctx.Injector.Resolve(&neffShuffle)
	if err != nil {
		return xerrors.Errorf("failed to resolve shuffle: %v", err)
	}

	keyPath := ctx.Flags.String("signer")

	signer, err := getSigner(keyPath)
	if err != nil {
		return xerrors.Errorf("failed to get signer: %v", err)
	}

	client, err := makeClient(ctx)
	if err != nil {
		return xerrors.Errorf("failed to make client: %v", err)
	}

	actor, err := neffShuffle.Listen(signed.NewManager(signer, &client))

	if err != nil {
		return xerrors.Errorf("failed to initialize the neff shuffle	protocol: %v", err)
	}

	ctx.Injector.Inject(actor)
	dela.Logger.Info().Msg("The shuffle protocol has been initialized successfully")

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

	var actor shuffle.Actor
	err = ctx.Injector.Resolve(&actor)
	if err != nil {
		return xerrors.Errorf("failed to resolve dkg.DKG: %v", err)
	}

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

	ep := eproxy.NewShuffle(actor, proxykey)

	router.HandleFunc("/evoting/services/shuffle/{electionID}", ep.EditShuffle).Methods("PUT")

	router.NotFoundHandler = http.HandlerFunc(eproxy.NotFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(eproxy.NotAllowedHandler)

	proxy.RegisterHandler("/evoting/services/shuffle/", router.ServeHTTP)

	dela.Logger.Info().Msg("DKG handler registered")

	return nil
}

func makeClient(ctx node.Context) (client, error) {
	var service ordering.Service
	err := ctx.Injector.Resolve(&service)
	if err != nil {
		return client{}, xerrors.Errorf("failed to resolve ordering.Service: %v", err)
	}

	var vs validation.Service
	err = ctx.Injector.Resolve(&vs)
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
