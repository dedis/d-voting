// Package main implements the dvoting backend
//
// Unix example:
//
//	# Expect GOPATH to be correctly set to have dvoting available.
//	go install
//
//	dvoting --config /tmp/node1 start --port 2001 &
//	dvoting --config /tmp/node2 start --port 2002 &
//	dvoting --config /tmp/node3 start --port 2003 &
//
//	# Share the different certificates among the participants.
//	dvoting --config /tmp/node2 minogrpc join --address 127.0.0.1:2001\
//	  $(dvoting --config /tmp/node1 minogrpc token)
//	dvoting --config /tmp/node3 minogrpc join --address 127.0.0.1:2001\
//	  $(dvoting --config /tmp/node1 minogrpc token)
//
//	# Create a chain with two members.
//	dvoting --config /tmp/node1 ordering setup\
//	  --member $(dvoting --config /tmp/node1 ordering export)\
//	  --member $(dvoting --config /tmp/node2 ordering export)
//
//	# Add the third after the chain is set up.
//	dvoting --config /tmp/node1 ordering roster add\
//	  --member $(dvoting --config /tmp/node3 ordering export)
package main

import (
	"fmt"
	"io"
	"os"

	dkg "github.com/c4dt/d-voting/services/dkg/pedersen/controller"
	"github.com/c4dt/d-voting/services/dkg/pedersen/json"
	shuffle "github.com/c4dt/d-voting/services/shuffle/neff/controller"

	cosipbft "github.com/c4dt/d-voting/cli/cosipbftcontroller"
	"github.com/c4dt/d-voting/cli/postinstall"
	evoting "github.com/c4dt/d-voting/contracts/evoting/controller"
	metrics "github.com/c4dt/d-voting/metrics/controller"
	"github.com/c4dt/dela/cli/node"
	access "github.com/c4dt/dela/contracts/access/controller"
	db "github.com/c4dt/dela/core/store/kv/controller"
	pool "github.com/c4dt/dela/core/txn/pool/controller"
	signed "github.com/c4dt/dela/core/txn/signed/controller"
	mino "github.com/c4dt/dela/mino/minogrpc/controller"
	proxy "github.com/c4dt/dela/mino/proxy/http/controller"

	_ "github.com/c4dt/d-voting/services/shuffle/neff/json"

	gapi "github.com/c4dt/dela-apps/gapi/controller"
)

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}

func run(args []string) error {
	return runWithCfg(args, config{Writer: os.Stdout})
}

type config struct {
	Channel chan os.Signal
	Writer  io.Writer
}

func runWithCfg(args []string, cfg config) error {
	json.Register()

	builder := node.NewBuilderWithCfg(
		cfg.Channel,
		cfg.Writer,
		db.NewController(),
		mino.NewController(),
		cosipbft.NewController(),
		dkg.NewController(),
		signed.NewManagerController(),
		pool.NewController(),
		access.NewController(),
		proxy.NewController(),
		shuffle.NewController(),
		evoting.NewController(),
		gapi.NewController(),
		metrics.NewController(),
		postinstall.NewController(),
	)

	app := builder.Build()

	err := app.Run(args)
	if err != nil {
		return err
	}

	return nil
}
