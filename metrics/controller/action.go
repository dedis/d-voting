package controller

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dvoting "go.dedis.ch/d-voting"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli/node"
	"golang.org/x/xerrors"
)

// parameters to wait for the http server to start
const (
	defaultRetry = 10
	waitRetry    = time.Second * 2
)

// StartAction defines the action to start the Prometheus server
type StartAction struct{}

// Execute implements node.ActionTemplate. It registers the Prometheus handler.
func (a StartAction) Execute(ctx node.Context) error {
	listenAddr := ctx.Flags.String("addr")
	path := ctx.Flags.String("path")

	fmt.Fprintf(ctx.Out, "starting metric server on %s %s\n", listenAddr, path)

	// We'll go with the default Prometheus registerer
	reg := prometheus.DefaultRegisterer

	// Register Dela collectors
	for _, c := range dela.PromCollectors {
		err := reg.Register(c)
		if err != nil {
			fmt.Fprintf(ctx.Out, "ERROR: failed to register: %v\n", err)
		}
	}

	// Register D-Voting collectors
	for _, c := range dvoting.PromCollectors {
		err := reg.Register(c)
		if err != nil {
			fmt.Fprintf(ctx.Out, "ERROR: failed to register: %v\n", err)
		}
	}

	// initialize a new HTTP sever and register the Prometheus handler
	srv := newMetricSrv(listenAddr)

	srv.mux.Handle(path, promhttp.Handler())
	fmt.Fprintf(ctx.Out, "registered prometheus service on %q\n", path)

	// start the server and wait a bit to check for potential errors
	errc := make(chan error, 1)

	go func() {
		errc <- srv.listen()
	}()

	for i := 0; i < defaultRetry && srv.getAddr() == nil; i++ {
		select {
		case err := <-errc:
			return xerrors.Errorf("failed to listen: %v", err)
		case <-time.After(waitRetry):
		}
	}

	srvAddr := srv.getAddr()
	if srvAddr == nil {
		return xerrors.Errorf("failed to start metric server")
	}

	dela.Logger.Info().Msgf("prometheus server started at %s", srvAddr)

	// inject the server so we can stop it later
	ctx.Injector.Inject(&srv)

	return nil
}

// metricHTTP defines an interface to uniquely identify the HTTP server that
// serves the Prometheus metrics.
type metricHTTP interface {
	MetricHTTP()
	Stop()
}

func newMetricSrv(listenAddr string) metricSrv {
	mux := http.NewServeMux()

	return metricSrv{
		mux: mux,
		server: &http.Server{
			Addr:    listenAddr,
			Handler: mux,
		},
		quit:       make(chan struct{}),
		listenAddr: listenAddr,
	}
}

// metricSrv wraps the Prometheus HTTP server
//
// - implements MetricHTTP
type metricSrv struct {
	mux        *http.ServeMux
	server     *http.Server
	quit       chan struct{}
	listenAddr string

	ln net.Listener
}

// listen start the HTTP server. This operation is blocking.
func (m *metricSrv) listen() error {
	go func() {
		<-m.quit

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		m.server.SetKeepAlivesEnabled(false)
		m.server.Shutdown(ctx)
	}()

	ln, err := net.Listen("tcp", m.listenAddr)
	if err != nil {
		return xerrors.Errorf("failed to create conn: %v", err)
	}

	m.ln = ln

	err = m.server.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		return xerrors.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (m *metricSrv) getAddr() net.Addr {
	if m.ln == nil {
		return nil
	}

	return m.ln.Addr()
}

func (m *metricSrv) Stop() {
	m.quit <- struct{}{}
}

// MetricHTTP implements MetricHTTP
func (metricSrv) MetricHTTP() {}
