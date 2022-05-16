package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	ptypes "github.com/dedis/d-voting/proxy/types"
	"go.dedis.ch/dela/cli"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/schnorr"
	"golang.org/x/xerrors"
)

const contentType = "application/json"

// initializer defines the standard CLI initializer for the dvoting CLI
//
// - implements cli.Initializer
type initializer struct {
}

func (i initializer) SetCommands(provider cli.Provider) {
	cmd := provider.SetCommand("dkg")

	sub := cmd.SetSubCommand("init")
	sub.SetDescription("initialize DKG")
	sub.SetFlags(cli.StringSliceFlag{
		Name:     "proxy",
		Usage:    "proxy addresses",
		Required: true,
	}, cli.StringFlag{
		Name:     "election",
		Usage:    "Election ID, hex encoded",
		Required: true,
	})
	sub.SetAction(dkgInitAction)

	sub = cmd.SetSubCommand("setup")
	sub.SetDescription("setup DKG")
	sub.SetFlags(cli.StringFlag{
		Name:     "proxy",
		Usage:    "proxy address",
		Required: true,
	}, cli.StringFlag{
		Name:     "election",
		Usage:    "Election ID, hex encoded",
		Required: true,
	})
	sub.SetAction(dkgSetupAction)

	cmd = provider.SetCommand("election")

	sub = cmd.SetSubCommand("")
}

func dkgInitAction(flags cli.Flags) error {
	secretkeyHex := os.Getenv("SK")
	if secretkeyHex == "" {
		return xerrors.New("'SK' env not set")
	}

	secretkeyBuf, err := hex.DecodeString(secretkeyHex)
	if err != nil {
		return xerrors.Errorf("failed to decode secretkeyHex: %v", err)
	}

	secret := suite.Scalar()

	err = secret.UnmarshalBinary(secretkeyBuf)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal secret key: %v", err)
	}

	electionIDHex := flags.String("election")
	proxies := flags.StringSlice("proxy")

	for _, proxy := range proxies {
		err = initDKG(secret, proxy, electionIDHex)
		if err != nil {
			fmt.Printf("failed to initDKG on %q for election %q: %v\n", proxy, electionIDHex, err)
		} else {
			fmt.Printf("dkg initialized on %q\n", proxy)
		}
	}

	return nil
}

func initDKG(secret kyber.Scalar, proxyAddr, electionIDHex string) error {
	setupDKG := ptypes.NewDKGRequest{
		ElectionID: electionIDHex,
	}

	signed, err := createSignedRequest(secret, setupDKG)
	if err != nil {
		return createSignedErr(err)
	}

	resp, err := http.Post(proxyAddr+"/evoting/services/dkg/actors", contentType, bytes.NewBuffer(signed))
	if err != nil {
		return xerrors.Errorf("failed to post request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	return nil
}

func dkgSetupAction(flags cli.Flags) error {
	secretkeyHex := os.Getenv("SK")
	if secretkeyHex == "" {
		return xerrors.New("'SK' env not set")
	}

	secretkeyBuf, err := hex.DecodeString(secretkeyHex)
	if err != nil {
		return xerrors.Errorf("failed to decode secretkeyHex: %v", err)
	}

	secret := suite.Scalar()

	err = secret.UnmarshalBinary(secretkeyBuf)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal secret key: %v", err)
	}

	electionIDHex := flags.String("election")
	proxy := flags.String("proxy")

	_, err = updateDKG(secret, proxy, electionIDHex, "setup")
	if err != nil {
		return xerrors.Errorf("failed to setup: %v", err)
	}

	fmt.Printf("DKG setup on %q\n", proxy)

	return nil
}

func updateDKG(secret kyber.Scalar, proxyAddr, electionIDHex, action string) (int, error) {
	msg := ptypes.UpdateDKG{
		Action: action,
	}

	signed, err := createSignedRequest(secret, msg)
	if err != nil {
		return 0, createSignedErr(err)
	}

	req, err := http.NewRequest(http.MethodPut, proxyAddr+"/evoting/services/dkg/actors/"+electionIDHex, bytes.NewBuffer(signed))
	if err != nil {
		return 0, xerrors.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, xerrors.Errorf("failed to execute the query: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return resp.StatusCode, xerrors.Errorf("unexpected status: %s - %s", resp.Status, buf)
	}

	return 0, nil
}

func createSignedRequest(secret kyber.Scalar, msg interface{}) ([]byte, error) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal json: %v", err)
	}

	payload := base64.URLEncoding.EncodeToString(jsonMsg)

	hash := sha256.New()

	hash.Write([]byte(payload))
	md := hash.Sum(nil)

	signature, err := schnorr.Sign(suite, secret, md)
	if err != nil {
		return nil, xerrors.Errorf("failed to sign: %v", err)
	}

	signed := ptypes.SignedRequest{
		Payload:   payload,
		Signature: hex.EncodeToString(signature),
	}

	signedJSON, err := json.Marshal(signed)
	if err != nil {
		return nil, xerrors.Errorf("failed to create json signed: %v", err)
	}

	return signedJSON, nil
}

func createSignedErr(err error) error {
	return xerrors.Errorf("failed to create signed request: %v", err)
}
