package controller

import (
	"github.com/dedis/d-voting/services/shuffle"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/core/validation"
	"golang.org/x/xerrors"
)

// initAction is an action to initialize the shuffle protocol
//
// - implements node.ActionTemplate
type initAction struct {
}

// Execute implements node.ActionTemplate. It creates an actor from
// the neffShuffle instance
func (a *initAction) Execute(ctx node.Context) error {
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
		mgr:  vs,
	}

	return client, nil
}

// client fetches the last nonce used by the client
//
// - implements signed.Client
type client struct {
	srvc ordering.Service
	mgr  validation.Service
}

// GetNonce implements signed.Client. It uses the validation service to get the
// last nonce.
func (c *client) GetNonce(id access.Identity) (uint64, error) {
	store := c.srvc.GetStore()

	nonce, err := c.mgr.GetNonce(store, id)
	if err != nil {
		return 0, xerrors.Errorf("failed to get nonce from validation: %v", err)
	}

	return nonce, nil
}
