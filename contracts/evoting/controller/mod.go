package controller

import (
	"github.com/c4dt/dela/cli"
	"github.com/c4dt/dela/cli/node"
	"github.com/c4dt/dela/core/access"
	"github.com/c4dt/dela/core/ordering"
	"github.com/c4dt/dela/core/validation"
)

// NewController returns a new controller initializer
func NewController() node.Initializer {
	return controller{}
}

// controller is an initializer with a set of commands.
//
// - implements node.Initializer
type controller struct {
}

// Build implements node.Initializer.
func (m controller) SetCommands(builder node.Builder) {

	cmd := builder.SetCommand("e-voting")
	cmd.SetDescription("interact with the evoting service")

	// dvoting --config /tmp/node1 e-voting \
	//   registerHandlers --signer private.key
	sub := cmd.SetSubCommand("registerHandlers")
	sub.SetDescription("register the e-voting handlers on the default proxy")
	sub.SetFlags(
		cli.StringFlag{
			Name:     "signer",
			Usage:    "Path to signer's private key",
			Required: true,
		},
	)
	sub.SetAction(builder.MakeAction(&RegisterAction{}))

	// dvoting --config /tmp/node1 e-voting scenarioTest
	sub = cmd.SetSubCommand("scenarioTest")
	sub.SetDescription("evoting scenario test")
	sub.SetFlags(
		cli.StringFlag{
			Name:     "secretkey",
			Usage:    "the proxy secret key to sign requests, hex encoded",
			Required: true,
		},
		cli.StringFlag{
			Name:  "proxy-addr1",
			Usage: "base address of the proxy for node 1",
			Value: "http://localhost:9080",
		},
		cli.StringFlag{
			Name:  "proxy-addr2",
			Usage: "base address of the proxy for node 2",
			Value: "http://localhost:9081",
		},
		cli.StringFlag{
			Name:  "proxy-addr3",
			Usage: "base address of the proxy for node 3",
			Value: "http://localhost:9082",
		},
	)
	sub.SetAction(builder.MakeAction(&scenarioTestAction{}))
}

// OnStart implements node.Initializer. It creates and registers a pedersen DKG.
func (m controller) OnStart(ctx cli.Flags, inj node.Injector) error {
	return nil
}

// OnStop implements node.Initializer.
func (controller) OnStop(node.Injector) error {
	return nil
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
func (c client) GetNonce(ident access.Identity) (uint64, error) {
	store := c.srvc.GetStore()

	nonce, err := c.mgr.GetNonce(store, ident)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}
