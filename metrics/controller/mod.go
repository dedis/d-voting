package controller

import (
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
)

// NewController returns a new controller initializer
func NewController() node.Initializer {
	return controller{}
}

// controller is an initializer with a set of commands.
//
// - implements node.Initializer
type controller struct{}

// Build implements node.Initializer.
func (m controller) SetCommands(builder node.Builder) {

	cmd := builder.SetCommand("metrics")
	cmd.SetDescription("handle Prometheus functions")

	sub := cmd.SetSubCommand("start")
	sub.SetDescription("start an http server to serve Prometheus metrics")
	sub.SetFlags(
		cli.StringFlag{
			Name:     "addr",
			Usage:    "The server address",
			Required: false,
			Value:    "127.0.0.1:9100",
		},
		cli.StringFlag{
			Name:     "path",
			Usage:    "The path to fetch the metrics",
			Required: false,
			Value:    "/metrics",
		},
	)
	sub.SetAction(builder.MakeAction(&startAction{}))
}

// OnStart implements node.Initializer. It creates and registers a pedersen DKG.
func (m controller) OnStart(ctx cli.Flags, inj node.Injector) error {
	return nil
}

// OnStop implements node.Initializer.
func (controller) OnStop(inj node.Injector) error {
	var srv metricHTTP

	err := inj.Resolve(&srv)
	if err == nil {
		srv.Stop()
	}

	return nil
}
