package controller

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/dkg/pedersen"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/store/kv"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/ed25519"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/mino/proxy"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

const separator = ":"

var suite = suites.MustFind("Ed25519")

// initAction is an action to initialize the DKG protocol
//
// - implements node.ActionTemplate
type initAction struct {
}

// Execute implements node.ActionTemplate. It creates an actor from
// the dkgPedersen instance and links it to an election.
func (a *initAction) Execute(ctx node.Context) error {
	electionIDBuf, err := hex.DecodeString(ctx.Flags.String("electionID"))
	if err != nil {
		return xerrors.Errorf("failed to decode electionID: %v", err)
	}
	// TODO: Check that election ID corresponds to a real election
	// ElectionMetadataKey has a list of election ids (types transaction.go)
	// electionMetadata.ElectionIDs.contains(electionId)

	// Initialize the actor
	var dkgPedersen dkg.DKG
	err = ctx.Injector.Resolve(&dkgPedersen)
	if err != nil {
		return xerrors.Errorf("failed to resolve dkg: %v", err)
	}

	actor, err := dkgPedersen.Listen(electionIDBuf)
	if err != nil {
		return xerrors.Errorf("failed to start the RPC: %v", err)
	}

	ctx.Injector.Inject(actor)
	dela.Logger.Info().Msg("DKG has been initialized successfully")

	var dkgMap kv.DB
	err = ctx.Injector.Resolve(&dkgMap)
	if err != nil {
		return xerrors.Errorf("failed to resolve dkgMap: %v", err)
	}

	// TODO: Will need actor to implement serde
	// actorData, err := actor.Serialize(json.NewContext())
	// Dummy for now
	actorData, err := hex.DecodeString(ctx.Flags.String("actor"))
	if err != nil {
		xerrors.Errorf("actor serialization failed: %v", err)
	}
	// actorBuf := serde.Serialize(actor.context)
	err = dkgMap.Update(func(tx kv.WritableTx) error {
		// TODO: What name should this have
		// It should be put in one place
		bucket, err := tx.GetBucketOrCreate([]byte("dkgmap"))
		if err != nil {
			return err
		}

		err = bucket.Set(electionIDBuf, actorData)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		xerrors.Errorf("database write failed: %v", err)
	}

	// TODO Needed?
	ctx.Injector.Inject(dkgMap)
	dela.Logger.Info().Msgf("DKG was successfully linked to election %v", electionIDBuf)
	return nil
}

// setupAction is an action to setup the DKG protocol and generate a collective
// public key
//
// - implements node.ActionTemplate
type setupAction struct {
}

// Execute implements node.ActionTemplate. It reads the list of members and
// request the setup.
func (a *setupAction) Execute(ctx node.Context) error {
	// TODO: This should now be obtained from dkgMap
	var actor dkg.Actor
	err := ctx.Injector.Resolve(&actor)
	if err != nil {
		return xerrors.Errorf("failed to resolve actor: %v", err)
	}

	electionIDBuf, err := hex.DecodeString(ctx.Flags.String("electionID"))
	if err != nil {
		return xerrors.Errorf("failed to decode electionID: %v", err)
	}

	pubkey, err := actor.Setup(electionIDBuf)
	if err != nil {
		return xerrors.Errorf("failed to setup DKG: %v", err)
	}

	pubkeyBuf, err := pubkey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to encode pubkey: %v", err)
	}

	dela.Logger.Info().
		Hex("DKG public key", pubkeyBuf).
		Msg("DKG public key")

	return nil
}

func (a setupAction) readMembers(ctx node.Context) (authority.Authority, error) {
	members := ctx.Flags.StringSlice("member")

	addrs := make([]mino.Address, len(members))
	pubkeys := make([]crypto.PublicKey, len(members))

	for i, member := range members {
		addr, pubkey, err := decodeMember(ctx, member)
		if err != nil {
			return nil, xerrors.Errorf("failed to decode: %v", err)
		}

		addrs[i] = addr
		pubkeys[i] = pubkey
	}

	return authority.New(addrs, pubkeys), nil
}

func decodeMember(ctx node.Context, str string) (mino.Address, crypto.PublicKey, error) {
	parts := strings.Split(str, separator)
	if len(parts) != 2 {
		return nil, nil, xerrors.Errorf("invalid member base64 string '%s'", str)
	}

	// 1. Deserialize the address.
	var m mino.Mino
	err := ctx.Injector.Resolve(&m)
	if err != nil {
		return nil, nil, xerrors.Errorf("injector: %v", err)
	}

	addrBuf, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, xerrors.Errorf("base64 address: %v", err)
	}

	addr := m.GetAddressFactory().FromText(addrBuf)

	// 2. Deserialize the public key.
	publicKeyFactory := ed25519.NewPublicKeyFactory()

	pubkeyBuf, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, xerrors.Errorf("base64 public key: %v", err)
	}

	pubkey, err := publicKeyFactory.FromBytes(pubkeyBuf)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to decode public key: %v", err)
	}

	return addr, pubkey, nil
}

// exportInfoAction is an action to display a base64 string describing the node.
// It can be used to transmit the identity of a node to another one.
//
// - implements node.ActionTemplate
type exportInfoAction struct {
}

// Execute implements node.ActionTemplate. It looks for the node address and
// public key and prints "$ADDR_BASE64:$PUBLIC_KEY_BASE64".
func (a *exportInfoAction) Execute(ctx node.Context) error {
	var m mino.Mino
	err := ctx.Injector.Resolve(&m)
	if err != nil {
		return xerrors.Errorf("injector: %v", err)
	}

	addr, err := m.GetAddress().MarshalText()
	if err != nil {
		return xerrors.Errorf("failed to marshal address: %v", err)
	}
	var pubkey kyber.Point

	err = ctx.Injector.Resolve(&pubkey)
	if err != nil {
		return xerrors.Errorf("injector: %v", err)
	}

	pubkeyMarshalled, err := pubkey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to marshal public key: %v", err)
	}

	desc := base64.StdEncoding.EncodeToString(addr) + separator + base64.StdEncoding.EncodeToString(pubkeyMarshalled)

	fmt.Fprint(ctx.Out, desc)

	return nil
}

// Ciphertext wraps the ciphertext pairs
type Ciphertext struct {
	K []byte
	C []byte
}

// getPublicKeyAction is an action that prints the collective public key
//
// - implements node.ActionTemplate
type getPublicKeyAction struct {
}

// Execute implements node.ActionTemplate. It retrieves the collective
// public key from the DKG service and prints it.
func (a *getPublicKeyAction) Execute(ctx node.Context) error {
	var actor dkg.Actor
	err := ctx.Injector.Resolve(&actor)
	if err != nil {
		return xerrors.Errorf("failed to resolve actor: %v", err)
	}

	pubkey, err := actor.GetPublicKey()
	if err != nil {
		return xerrors.Errorf("failed to retrieve the public key: %v", err)
	}

	pubkeyBuf, err := pubkey.MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to encode pubkey: %v", err)
	}

	dela.Logger.Info().
		Hex("DKG public key", pubkeyBuf).
		Msg("DKG public key")

	return nil
}

// registerHandlersAction is an action that registers the proxy handlers
//
// - implements node.ActionTemplate
type registerHandlersAction struct {
}

// Execute implements node.ActionTemplate. It registers the proxy
// handlers to set up elections
func (a *registerHandlersAction) Execute(ctx node.Context) error {
	var proxy proxy.Proxy
	err := ctx.Injector.Resolve(&proxy)
	if err != nil {
		return xerrors.Errorf("failed to resolve proxy: %v", err)
	}

	var pedersen pedersen.Pedersen
	err = ctx.Injector.Resolve(pedersen)
	if err != nil {
		return xerrors.Errorf("failed to resolve dkg.DKG: %v", err)
	}

	proxy.RegisterHandler("/evoting/dkg", getHandler(pedersen))

	dela.Logger.Info().Msg("DKG handler registered")

	return nil
}

func getHandler(pedersen pedersen.Pedersen) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// hex-encoded string obtained from the URL
		electionID := r.FormValue("electionID")

		if electionID == "" {
			http.Error(w, "electionID was not found in request", http.StatusBadRequest)
			return
		}

		electionIDBuf, err := hex.DecodeString(electionID)
		if err != nil {
			http.Error(w, "failed to decode electionID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		a, err := pedersen.GetActor(electionIDBuf)
		if err != nil {
			http.Error(w, "failed to get actor: "+err.Error(), http.StatusInternalServerError)
			return
		}

		pubKey, err := a.Setup(electionIDBuf)
		if err != nil {
			http.Error(w, "failed to setup: "+err.Error(), http.StatusInternalServerError)
			return
		}

		pubKeyBuf, err := pubKey.MarshalBinary()
		if err != nil {
			http.Error(w, "failed to marshal the pubKey: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(pubKeyBuf)
	}
}
