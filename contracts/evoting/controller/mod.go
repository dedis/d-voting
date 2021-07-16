package controller

import (
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"golang.org/x/xerrors"
	"strings"
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

	// memcoin --config /tmp/node1 e-voting initHttpServer --portNumber 8080
	sub := cmd.SetSubCommand("initHttpServer")
	sub.SetDescription("initialize the HTTP server")
	sub.SetFlags(cli.StringFlag{
		Name:     "portNumber",
		Usage:    "port number of the HTTP server",
		Required: true,
	})

	sub.SetAction(builder.MakeAction(&initHttpServerAction{
		client: &Client{
			Nonce: 0,
		},
	}))

	sub = cmd.SetSubCommand("scenarioTest")
	sub.SetDescription("evoting scenario test")
	sub.SetFlags(cli.StringSliceFlag{
		Name:     "member",
		Usage:    "nodes participating in SHUFFLE",
		Required: true,
	})
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

// Client fetches the last nonce used and returns nonce + 1
//
// - implements signed.Client
type Client struct {
	Nonce  uint64
	Blocks blockstore.BlockStore
}

// GetNonce implements signed.Client
func (c *Client) GetNonce(access.Identity) (uint64, error) {
	blockLink, err := c.Blocks.Last()
	if err != nil {
		return 0, xerrors.Errorf("failed to fetch last block: %v", err)
	}

	transactionResults := blockLink.GetBlock().GetData().GetTransactionResults()
	nonce := uint64(0)

	for nonce == 0 {
		for _, txResult := range transactionResults {
			_, msg := txResult.GetStatus()
			if !strings.Contains(msg, "nonce") && txResult.GetTransaction().GetNonce() > nonce {
				// && txResult.GetTransaction().GetIdentity().Equal(signer.GetPublicKey())
				nonce = txResult.GetTransaction().GetNonce()
			}
		}

		previousDigest := blockLink.GetFrom()

		previousBlock, err := c.Blocks.Get(previousDigest)
		if err != nil {
			if strings.Contains(err.Error(), "not found: no block") {
				dela.Logger.Info().Msg("FIRST BLOCK")
				break
			} else {
				return 0, xerrors.Errorf("failed to fetch previous block: %v", err)
			}
		} else {
			transactionResults = previousBlock.GetBlock().GetData().GetTransactionResults()
		}
	}
	nonce++
	c.Nonce = nonce

	return nonce, nil
}
