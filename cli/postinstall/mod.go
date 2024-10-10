package postinstall

import (
	"os"
	"path/filepath"
	"time"

	evoting "github.com/dedis/d-voting/contracts/evoting/controller"
	prom "github.com/dedis/d-voting/metrics/controller"
	dkg "github.com/dedis/d-voting/services/dkg/pedersen/controller"
	neff "github.com/dedis/d-voting/services/shuffle/neff/controller"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/mino/proxy"
	"go.dedis.ch/dela/mino/proxy/http"
	"golang.org/x/xerrors"
)

var defaultRetry = 10
var proxyFac func(string) proxy.Proxy = http.NewHTTP

const defaultProxyAddr = "127.0.0.1:0"
const defaultPromAddr = "127.0.0.1:0"

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
	builder.SetStartFlags(
		cli.StringFlag{
			Name:     "proxyaddr",
			Usage:    "the proxy address",
			Required: false,
			Value:    defaultProxyAddr,
		},
		cli.StringFlag{
			Name:     "promaddr",
			Usage:    "the Prometheus address",
			Required: false,
			Value:    defaultPromAddr,
		},
		cli.BoolFlag{
			Name:     "postinstall",
			Usage:    "run the postinstall functions",
			Required: false,
		},
		cli.StringFlag{
			Name:     "proxykey",
			Usage:    "the frontend public key that signs requests, hex encoded",
			Required: false,
		},
	)
}

// OnStart implements node.Initializer. It creates and registers a pedersen DKG.
func (m controller) OnStart(ctx cli.Flags, inj node.Injector) error {
	if !ctx.Bool("postinstall") {
		dela.Logger.Info().Msg("not using postinstall")
		return nil
	}

	dela.Logger.Info().Msg("using postinstalls")

	//
	// Init the shuffle
	//

	sinit := neff.InitAction{}

	err := sinit.Execute(node.Context{
		Injector: inj,
		Flags: node.FlagSet{
			"signer": filepath.Join(ctx.Path("config"), "private.key"),
		},
		Out: os.Stdout,
	})

	if err != nil {
		return xerrors.Errorf("failed to auto init shuffle: %v", err)
	}

	//
	// Start the proxy server
	//

	proxyAddr := ctx.String("proxyaddr")

	proxyhttp := proxyFac(proxyAddr)

	inj.Inject(proxyhttp)

	go proxyhttp.Listen()

	for i := 0; i < defaultRetry && proxyhttp.GetAddr() == nil; i++ {
		time.Sleep(time.Second)
	}

	if proxyhttp.GetAddr() == nil {
		return xerrors.Errorf("failed to start proxy server")
	}

	// We assume the listen worked proprely, however it might not be the case.
	// The log should inform the user about that.
	dela.Logger.Info().Msgf("started proxy server on %s", proxyhttp.GetAddr().String())

	//
	// Register the d-voting proxy handlers
	//

	eregister := evoting.RegisterAction{}
	err = eregister.Execute(node.Context{
		Injector: inj,
		Flags: node.FlagSet{
			"signer":   filepath.Join(ctx.Path("config"), "private.key"),
			"proxykey": ctx.String("proxykey"),
		},
		Out: os.Stdout,
	})

	if err != nil {
		return xerrors.Errorf("failed to register evoting handlers: %v", err)
	}

	//
	// Register the DKG proxy handlers
	//

	dregister := dkg.RegisterHandlersAction{}
	err = dregister.Execute(node.Context{
		Injector: inj,
		Flags:    ctx,
		Out:      os.Stdout,
	})

	if err != nil {
		return xerrors.Errorf("failed to register dkg handlers: %v", err)
	}

	//
	// Register the Shuffle proxy handlers
	//

	nregister := neff.RegisterHandlersAction{}
	err = nregister.Execute(node.Context{
		Injector: inj,
		Flags:    ctx,
		Out:      os.Stdout,
	})

	if err != nil {
		return xerrors.Errorf("failed to register neff handlers: %v", err)
	}

	//
	// Start the Prometheus server
	//

	pstart := prom.StartAction{}
	pstart.Execute(node.Context{
		Injector: inj,
		Flags: node.FlagSet{
			"addr": ctx.String("promaddr"),
			"path": "/metrics",
		},
		Out: os.Stdout,
	})

	return nil
}

// OnStop implements node.Initializer.
func (controller) OnStop(inj node.Injector) error {
	return nil
}
