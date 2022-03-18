package minocontroller

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/internal/testing/fake"
	"go.dedis.ch/dela/mino/minogrpc"
	"go.dedis.ch/dela/mino/minogrpc/certs"
)

func TestCertAction_Execute(t *testing.T) {
	action := certAction{}

	out := new(bytes.Buffer)
	req := node.Context{
		Out:      out,
		Injector: node.NewInjector(),
	}

	cert := fake.MakeCertificate(t, 1)

	store := certs.NewInMemoryStore()
	store.Store(fake.NewAddress(0), cert)

	req.Injector.Inject(fakeJoinable{certs: store})

	err := action.Execute(req)
	require.NoError(t, err)
	expected := fmt.Sprintf("Address: fake.Address[0] Certificate: %v\n", cert.Leaf.NotAfter)
	require.Equal(t, expected, out.String())

	req.Injector = node.NewInjector()
	err = action.Execute(req)
	require.EqualError(t, err,
		"couldn't resolve: couldn't find dependency for 'minogrpc.Joinable'")
}

func TestTokenAction_Execute(t *testing.T) {
	action := tokenAction{}

	flags := make(node.FlagSet)
	flags["expiration"] = time.Millisecond

	out := new(bytes.Buffer)
	req := node.Context{
		Out:      out,
		Flags:    flags,
		Injector: node.NewInjector(),
	}

	cert := fake.MakeCertificate(t, 1)

	store := certs.NewInMemoryStore()
	store.Store(fake.NewAddress(0), cert)

	hash, err := store.Hash(cert)
	require.NoError(t, err)

	req.Injector.Inject(fakeJoinable{certs: store})

	err = action.Execute(req)
	require.NoError(t, err)

	expected := fmt.Sprintf("--token abc --cert-hash %s\n",
		base64.StdEncoding.EncodeToString(hash))
	require.Equal(t, expected, out.String())

	req.Injector = node.NewInjector()
	err = action.Execute(req)
	require.EqualError(t, err,
		"couldn't resolve: couldn't find dependency for 'minogrpc.Joinable'")
}

func TestJoinAction_Execute(t *testing.T) {
	action := joinAction{}

	flags := make(node.FlagSet)
	flags["cert-hash"] = "YQ=="

	req := node.Context{
		Flags:    flags,
		Injector: node.NewInjector(),
	}

	req.Injector.Inject(fakeJoinable{})

	err := action.Execute(req)
	require.NoError(t, err)

	flags["cert-hash"] = "a"
	err = action.Execute(req)
	require.EqualError(t, err,
		"couldn't decode digest: illegal base64 data at input byte 0")

	flags["cert-hash"] = "YQ=="
	req.Injector.Inject(fakeJoinable{err: fake.GetError()})
	err = action.Execute(req)
	require.EqualError(t, err, fake.Err("couldn't join"))

	req.Injector = node.NewInjector()
	err = action.Execute(req)
	require.EqualError(t, err,
		"couldn't resolve: couldn't find dependency for 'minogrpc.Joinable'")
}

// -----------------------------------------------------------------------------
// Utility functions

type fakeJoinable struct {
	minogrpc.Joinable
	certs certs.Storage
	err   error
}

func (j fakeJoinable) GetCertificate() *tls.Certificate {
	cert, err := j.certs.Load(fake.NewAddress(0))
	if err != nil {
		panic(err)
	}

	return cert
}

func (j fakeJoinable) GetCertificateStore() certs.Storage {
	return j.certs
}

func (j fakeJoinable) GenerateToken(time.Duration) string {
	return "abc"
}

func (j fakeJoinable) Join(string, string, []byte) error {
	return j.err
}

type fakeContext struct {
	cli.Flags
	duration time.Duration
	str      string
	path     string
	num      int
}

func (ctx fakeContext) Duration(string) time.Duration {
	return ctx.duration
}

func (ctx fakeContext) String(string) string {
	return ctx.str
}

func (ctx fakeContext) Path(string) string {
	return ctx.path
}

func (ctx fakeContext) Int(string) int {
	return ctx.num
}
