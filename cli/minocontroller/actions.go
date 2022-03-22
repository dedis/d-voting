// This file contains the implementation of the controller actions.
//
// Documentation Last Review: 07.10.2020
//

package minocontroller

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"

	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/mino/minogrpc"
	"golang.org/x/xerrors"
)

// CertAction is an action to list the certificates known by the server.
//
// - implements node.ActionTemplate
type certAction struct{}

// Execute implements node.ActionTemplate. It prints the list of certificates
// known by the server with the address associated and the expiration date.
func (a certAction) Execute(req node.Context) error {
	var m minogrpc.Joinable

	err := req.Injector.Resolve(&m)
	if err != nil {
		return xerrors.Errorf("couldn't resolve: %v", err)
	}

	m.GetCertificateStore().Range(func(addr mino.Address, cert *tls.Certificate) bool {
		fmt.Fprintf(req.Out, "Address: %v Certificate: %v\n", addr, cert.Leaf.NotAfter)
		return true
	})

	return nil
}

// TokenAction is an action to generate a token that will be valid for another
// server to join the network of participants.
//
// - implements node.ActionTemplate
type tokenAction struct{}

// Execute implements node.ActionTemplate. It generates a token that will be
// valid for the amount of time given in the request.
func (a tokenAction) Execute(req node.Context) error {
	exp := req.Flags.Duration("expiration")

	var m minogrpc.Joinable
	err := req.Injector.Resolve(&m)
	if err != nil {
		return xerrors.Errorf("couldn't resolve: %v", err)
	}

	token := m.GenerateToken(exp)

	digest, err := m.GetCertificateStore().Hash(m.GetCertificate())
	if err != nil {
		return xerrors.Errorf("couldn't hash certificate: %v", err)
	}

	fmt.Fprintf(req.Out, "--token %s --cert-hash %s\n",
		token, base64.StdEncoding.EncodeToString(digest))

	return nil
}

// JoinAction is an action to join a network of participants by providing a
// valid token and the certificate hash.
//
// - implements node.ActionTemplate
type joinAction struct{}

// Execute implements node.ActionTemplate. It parses the request and send the
// join request to the distant node.
func (a joinAction) Execute(req node.Context) error {
	token := req.Flags.String("token")
	addr := req.Flags.String("address")
	certHash := req.Flags.String("cert-hash")

	var m minogrpc.Joinable
	err := req.Injector.Resolve(&m)
	if err != nil {
		return xerrors.Errorf("couldn't resolve: %v", err)
	}

	cert, err := base64.StdEncoding.DecodeString(certHash)
	if err != nil {
		return xerrors.Errorf("couldn't decode digest: %v", err)
	}

	err = m.Join(addr, token, cert)
	if err != nil {
		return xerrors.Errorf("couldn't join: %v", err)
	}

	return nil
}
