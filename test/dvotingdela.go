package integration

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/services/dkg"
	"github.com/dedis/d-voting/services/dkg/pedersen"
	"github.com/dedis/d-voting/services/shuffle"
	"github.com/dedis/d-voting/services/shuffle/neff"
	"github.com/stretchr/testify/require"
	accessContract "go.dedis.ch/dela/contracts/access"
	"go.dedis.ch/dela/contracts/value"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/access/darc"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	"go.dedis.ch/dela/core/ordering/cosipbft/types"
	"go.dedis.ch/dela/core/store/hashtree"
	"go.dedis.ch/dela/core/store/hashtree/binprefix"
	"go.dedis.ch/dela/core/store/kv"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	poolimpl "go.dedis.ch/dela/core/txn/pool/gossip"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/core/validation"
	"go.dedis.ch/dela/core/validation/simple"
	"go.dedis.ch/dela/cosi/threshold"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/crypto/bls"
	"go.dedis.ch/dela/crypto/loader"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/mino/gossip"
	"go.dedis.ch/dela/mino/minogrpc"
	"go.dedis.ch/dela/mino/minogrpc/certs"
	"go.dedis.ch/dela/mino/minogrpc/session"
	"go.dedis.ch/dela/mino/router/tree"
	"go.dedis.ch/dela/serde/json"
	"golang.org/x/xerrors"
)

const certKeyName = "cert.key"
const privateKeyFile = "private.key"

var aKey = [32]byte{1}
var valueAccessKey = [32]byte{2}

// evotingAccessKey is the access key used for the evoting contract.
var evotingAccessKey = [32]byte{3}

// dela defines the common interface for a Dela node.
type dela interface {
	Setup(...dela)
	GetMino() mino.Mino
	GetOrdering() ordering.Service
	GetTxManager() txn.Manager
	GetAccessService() access.Service
}

// dVotingCosiDela defines the interface needed to use a Dela node using cosi.
type dVotingCosiDela interface {
	dela

	GetPublicKey() crypto.PublicKey
	GetPool() pool.Pool
	GetAccessStore() accessstore
	GetTree() hashtree.Tree
	GetDkg() dkg.DKG
	GetShuffle() shuffle.Shuffle
	GetShuffleSigner() crypto.AggregateSigner
	GetValidationSrv() validation.Service
}

// dVotingNode represents a Dela node using cosi pbft aimed to execute d-voting
//
// - implements dVotingCosiDela
type dVotingNode struct {
	t             *testing.T
	onet          mino.Mino
	ordering      ordering.Service
	cosi          *threshold.Threshold
	txManager     txn.Manager
	pool          pool.Pool
	accessService access.Service
	accessStore   accessstore
	tree          hashtree.Tree
	dkg           *pedersen.Pedersen
	shuffle       *neff.NeffShuffle
	shuffleSigner crypto.AggregateSigner
	vs            validation.Service
}

// Creates n dela nodes using tempDir as root to file path and returns an array
// of nodes or error
func setupDVotingNodes(t *testing.T, numberOfNodes int, tempDir string) []dVotingCosiDela {

	wait := sync.WaitGroup{}

	wait.Add(numberOfNodes)

	nodes := make(chan dVotingCosiDela, numberOfNodes)

	randSource := rand.NewSource(int64(0))

	for n := 0; n < numberOfNodes; n++ {
		go func(i int) {
			defer wait.Done()
			filePath := filepath.Join(tempDir, "node", strconv.Itoa(i))
			nodes <- newDVotingNode(t, filePath, randSource)
		}(n)
	}

	wait.Wait()
	close(nodes)

	delaNodes := make([]dela, 0, numberOfNodes)
	dVotingNodes := make([]dVotingCosiDela, 0, numberOfNodes)

	for node := range nodes {
		delaNodes = append(delaNodes, node)
		dVotingNodes = append(dVotingNodes, node)
	}
	delaNodes[0].Setup(delaNodes[1:]...)

	return dVotingNodes
}

// Creates a single dVotingCosiDela node
func newDVotingNode(t *testing.T, path string, randSource rand.Source) dVotingCosiDela {
	err := os.MkdirAll(path, 0700)
	require.NoError(t, err)

	os.Setenv("LLVL", "info")

	// store
	db, err := kv.New(filepath.Join(path, "dela.db"))
	require.NoError(t, err)

	// mino
	router := tree.NewRouter(minogrpc.NewAddressFactory())
	addr := minogrpc.ParseAddress("127.0.0.1", uint16(0))

	certs := certs.NewDiskStore(db, session.AddressFactory{})

	fload := loader.NewFileLoader(filepath.Join(path, certKeyName))

	keydata, err := fload.LoadOrCreate(newCertGenerator(rand.New(randSource), elliptic.P521()))
	require.NoError(t, err)

	key, err := x509.ParseECPrivateKey(keydata)
	require.NoError(t, err)

	opts := []minogrpc.Option{
		minogrpc.WithStorage(certs),
		minogrpc.WithCertificateKey(key, key.Public()),
	}

	onet, err := minogrpc.NewMinogrpc(addr, router, opts...)
	require.NoError(t, err)

	// ordering + validation + execution
	fload = loader.NewFileLoader(filepath.Join(path, privateKeyFile))

	signerdata, err := fload.LoadOrCreate(newKeyGenerator())
	require.NoError(t, err)

	signer, err := bls.NewSignerFromBytes(signerdata)
	require.NoError(t, err)

	cosi := threshold.NewThreshold(onet.WithSegment("cosi"), signer)
	cosi.SetThreshold(threshold.ByzantineThreshold)

	exec := native.NewExecution()
	accessService := darc.NewService(json.NewContext())

	rosterFac := authority.NewFactory(onet.GetAddressFactory(), cosi.GetPublicKeyFactory())
	cosipbft.RegisterRosterContract(exec, rosterFac, accessService)

	value.RegisterContract(exec, value.NewContract(valueAccessKey[:], accessService))

	txFac := signed.NewTransactionFactory()
	vs := simple.NewService(exec, txFac)

	pool, err := poolimpl.NewPool(gossip.NewFlat(onet.WithSegment("pool"), txFac))
	require.NoError(t, err)

	tree := binprefix.NewMerkleTree(db, binprefix.Nonce{})

	param := cosipbft.ServiceParam{
		Mino:       onet,
		Cosi:       cosi,
		Validation: vs,
		Access:     accessService,
		Pool:       pool,
		DB:         db,
		Tree:       tree,
	}

	err = tree.Load()
	require.NoError(t, err)

	genstore := blockstore.NewGenesisDiskStore(db, types.NewGenesisFactory(rosterFac))

	err = genstore.Load()
	require.NoError(t, err)

	blockFac := types.NewBlockFactory(vs.GetFactory())
	csFac := authority.NewChangeSetFactory(onet.GetAddressFactory(), cosi.GetPublicKeyFactory())
	linkFac := types.NewLinkFactory(blockFac, cosi.GetSignatureFactory(), csFac)

	blocks := blockstore.NewDiskStore(db, linkFac)

	err = blocks.Load()
	require.NoError(t, err)

	srvc, err := cosipbft.NewService(param, cosipbft.WithGenesisStore(genstore), cosipbft.WithBlockStore(blocks))
	require.NoError(t, err)

	// tx
	mgr := signed.NewManager(cosi.GetSigner(), client{
		srvc: srvc,
		mgr:  vs,
	})

	// access
	accessStore := newAccessStore()
	contract := accessContract.NewContract(aKey[:], accessService, accessStore)
	accessContract.RegisterContract(exec, contract)

	dkg, _ := pedersen.NewPedersen(onet, true, srvc, rosterFac)

	rosterKey := [32]byte{}
	evoting.RegisterContract(exec, evoting.NewContract(evotingAccessKey[:], rosterKey[:],
		accessService, dkg, rosterFac))

	neffShuffle := neff.NewNeffShuffle(onet, srvc, pool, blocks, rosterFac, signer)

	// Neff shuffle signer
	l := loader.NewFileLoader(filepath.Join(path, "private_neff.key"))

	neffSignerdata, err := l.LoadOrCreate(newKeyGenerator())
	require.NoError(t, err)

	neffSigner, err := bls.NewSignerFromBytes(neffSignerdata)
	require.NoError(t, err)

	return dVotingNode{
		t:             t,
		onet:          onet,
		ordering:      srvc,
		cosi:          cosi,
		txManager:     mgr,
		pool:          pool,
		accessService: accessService,
		accessStore:   accessStore,
		tree:          tree,
		dkg:           dkg,
		shuffle:       neffShuffle,
		shuffleSigner: neffSigner,
		vs:            vs,
	}
}

// Creates an access on all dVotingCosiDela node given
func createDVotingAccess(t *testing.T, nodes []dVotingCosiDela, dir string) crypto.AggregateSigner {
	l := loader.NewFileLoader(filepath.Join(dir, "private.key"))

	signerdata, err := l.LoadOrCreate(newKeyGenerator())
	require.NoError(t, err)

	signer, err := bls.NewSignerFromBytes(signerdata)
	require.NoError(t, err)

	pubKey := signer.GetPublicKey()
	cred := accessContract.NewCreds(aKey[:])

	for _, node := range nodes {
		n := node.(dVotingNode)
		n.GetAccessService().Grant(n.GetAccessStore(), cred, pubKey)
	}

	return signer
}

// Setup implements dela. It creates the roster, shares the certificate, and
// create an new chain.
func (c dVotingNode) Setup(nodes ...dela) {
	// share the certificates
	joinable, ok := c.onet.(minogrpc.Joinable)
	require.True(c.t, ok)

	addrStr := c.onet.GetAddress().String()
	token := joinable.GenerateToken(time.Hour)

	certHash, err := joinable.GetCertificateStore().Hash(joinable.GetCertificate())
	require.NoError(c.t, err)

	for _, n := range nodes {
		otherJoinable, ok := n.GetMino().(minogrpc.Joinable)
		require.True(c.t, ok)

		err = otherJoinable.Join(addrStr, token, certHash)
		require.NoError(c.t, err)
	}

	type extendedService interface {
		GetRoster() (authority.Authority, error)
		Setup(ctx context.Context, ca crypto.CollectiveAuthority) error
	}

	// make roster
	extended, ok := c.GetOrdering().(extendedService)
	require.True(c.t, ok)

	minoAddrs := make([]mino.Address, len(nodes)+1)
	pubKeys := make([]crypto.PublicKey, len(nodes)+1)

	for i, n := range nodes {
		minoAddr := n.GetMino().GetAddress()

		d, ok := n.(dVotingCosiDela)
		require.True(c.t, ok)

		pubkey := d.GetPublicKey()

		minoAddrs[i+1] = minoAddr
		pubKeys[i+1] = pubkey
	}

	minoAddrs[0] = c.onet.GetAddress()
	pubKeys[0] = c.cosi.GetSigner().GetPublicKey()

	roster := authority.New(minoAddrs, pubKeys)

	// create chain
	err = extended.Setup(context.Background(), roster)
	require.NoError(c.t, err)
}

// GetMino implements dVotingNode
func (c dVotingNode) GetMino() mino.Mino {
	return c.onet
}

// GetOrdering implements dVotingNode
func (c dVotingNode) GetOrdering() ordering.Service {
	return c.ordering
}

// GetTxManager implements dVotingNode
func (c dVotingNode) GetTxManager() txn.Manager {
	return c.txManager
}

// GetAccessService implements dVotingNode
func (c dVotingNode) GetAccessService() access.Service {
	return c.accessService
}

// GetPublicKey  implements dVotingNode
func (c dVotingNode) GetPublicKey() crypto.PublicKey {
	return c.cosi.GetSigner().GetPublicKey()
}

// GetPool implements dVotingNode
func (c dVotingNode) GetPool() pool.Pool {
	return c.pool
}

// GetAccessStore implements dVotingNode
func (c dVotingNode) GetAccessStore() accessstore {
	return c.accessStore
}

// GetTree implements dVotingNode
func (c dVotingNode) GetTree() hashtree.Tree {
	return c.tree
}

// GetDkg implements dVotingNode
func (c dVotingNode) GetDkg() dkg.DKG {
	return c.dkg
}

// GetShuffle implements dVotingNode
func (c dVotingNode) GetShuffle() shuffle.Shuffle {
	return c.shuffle
}

// GetShuffleSigner implements dVotingNode
func (c dVotingNode) GetShuffleSigner() crypto.AggregateSigner {
	return c.shuffleSigner
}

// GetValidationSrv implements dVotingNode
func (c dVotingNode) GetValidationSrv() validation.Service {
	return c.vs
}

// certGenerator can generate a private key compatible with the x509 certificate.
//
// - implements loader.Generator
type certGenerator struct {
	random io.Reader
	curve  elliptic.Curve
}

func newCertGenerator(r io.Reader, c elliptic.Curve) loader.Generator {
	return certGenerator{
		random: r,
		curve:  c,
	}
}

// Generate implements loader.Generator. It returns the serialized data of
// a private key generated from the an elliptic curve. The data is formatted as
// a PEM block "EC PRIVATE KEY".
func (g certGenerator) Generate() ([]byte, error) {
	priv, err := ecdsa.GenerateKey(g.curve, g.random)
	if err != nil {
		return nil, xerrors.Errorf("ecdsa: %v", err)
	}

	data, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, xerrors.Errorf("while marshaling: %v", err)
	}

	return data, nil
}

func newKeyGenerator() loader.Generator {
	return keyGenerator{}
}

// keyGenerator is an implementation to generate a private key.
//
// - implements loader.Generator
type keyGenerator struct {
}

// Generate implements loader.Generator. It returns the marshaled data of a
// private key.
func (g keyGenerator) Generate() ([]byte, error) {
	signer := bls.NewSigner()

	data, err := signer.MarshalBinary()
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal signer: %v", err)
	}

	return data, nil
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

// newAccessStore returns a new access store
func newAccessStore() accessstore {
	return accessstore{
		bucket: make(map[string][]byte),
	}
}

// accessstore is an in-memory store access
//
// - implements store.Readable
// - implements store.Writable
type accessstore struct {
	bucket map[string][]byte
}

// Get implements store.Readable
func (a accessstore) Get(key []byte) ([]byte, error) {
	return a.bucket[string(key)], nil
}

// Set implements store.Writable
func (a accessstore) Set(key, value []byte) error {
	a.bucket[string(key)] = value

	return nil
}

// Delete implements store.Writable
func (a accessstore) Delete(key []byte) error {
	delete(a.bucket, string(key))

	return nil
}
