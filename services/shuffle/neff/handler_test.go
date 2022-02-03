package neff

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"strconv"
	"testing"

	"github.com/dedis/d-voting/services/shuffle/neff/types"
	"go.dedis.ch/kyber/v3"

	evotingController "github.com/dedis/d-voting/contracts/evoting/controller"
	electionTypes "github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
	orderingTypes "go.dedis.ch/dela/core/ordering/cosipbft/types"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/pool"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/core/validation"
	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/xerrors"
)

func TestHandler_Stream(t *testing.T) {
	handler := Handler{}
	receiver := fake.NewBadReceiver()
	err := handler.Stream(fake.Sender{}, receiver)
	require.EqualError(t, err, fake.Err("failed to receive"))

	receiver = fake.NewReceiver(
		fake.NewRecvMsg(fake.NewAddress(0), fake.Message{}),
	)
	err = handler.Stream(fake.Sender{}, receiver)
	require.EqualError(t, err, "expected StartShuffle message, got: fake.Message")

	receiver = fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0),
		types.NewStartShuffle("dummyID", make([]mino.Address, 0))))

	err = handler.Stream(fake.Sender{}, receiver)
	require.EqualError(t, err, "failed to handle StartShuffle message: failed "+
		"to get election: failed to decode electionIDHex: encoding/hex: invalid byte: U+0075 'u'")

	//Test successful Shuffle round from message:
	dummyId := hex.EncodeToString([]byte("dummyId"))
	handler = initValidHandler(dummyId)

	receiver = fake.NewReceiver(fake.NewRecvMsg(fake.NewAddress(0), types.NewStartShuffle(dummyId, make([]mino.Address, 0))))
	err = handler.Stream(fake.Sender{}, receiver)

	require.NoError(t, err)

}

func TestHandler_StartShuffle(t *testing.T) {
	// Some initialisation:
	k := 3

	KsMarshalled, CsMarshalled, pubKey := fakeKCPointsMarshalled(k)

	fakeErr := xerrors.Errorf("fake error")

	handler := Handler{
		me: fake.NewAddress(0),
	}
	dummyId := hex.EncodeToString([]byte("dummyId"))

	// Service not working:
	badService := FakeService{
		err:        fakeErr,
		election:   nil,
		electionId: electionTypes.ID(dummyId),
	}
	handler.service = &badService

	err := handler.handleStartShuffle(dummyId)
	require.EqualError(t, err, "failed to get election: failed to get proof: fake error")

	// Election does not exist
	service := FakeService{
		err:        nil,
		election:   nil,
		electionId: electionTypes.ID(dummyId),
		context:    serdecontext,
	}
	handler.service = &service

	err = handler.handleStartShuffle(dummyId)
	require.EqualError(t, err, "failed to get election: election does not exist")

	// Election still opened:
	election := electionTypes.Election{
		ElectionID:          dummyId,
		AdminID:             "dummyAdminID",
		Status:              0,
		Pubkey:              nil,
		PublicBulletinBoard: electionTypes.PublicBulletinBoard{},
		ShuffleInstances:    []electionTypes.ShuffleInstance{},
		DecryptedBallots:    nil,
		ShuffleThreshold:    1,
		BallotSize:          1,
	}

	service = updateService(election, dummyId)
	handler.service = &service
	handler.context = serdecontext

	err = handler.handleStartShuffle(dummyId)
	require.EqualError(t, err, "the election must be closed: but status is 0")

	// Wrong formatted ballots:
	election.Status = electionTypes.Closed

	election.PublicBulletinBoard.DeleteUser("fakeUser")

	for i := 0; i < k; i++ {
		ballot := electionTypes.EncryptedBallot{electionTypes.Ciphertext{
			K: []byte("fakeVoteK"),
			C: []byte("fakeVoteC"),
		},
		}
		election.PublicBulletinBoard.CastVote("dummyUser"+strconv.Itoa(i), ballot)
	}

	service = updateService(election, dummyId)

	handler.service = &service
	err = handler.handleStartShuffle(dummyId)
	require.EqualError(t, err,
		"failed to make tx: failed to get shuffled ballots: failed to get X, Y:"+
			" failed to get points: failed to unmarshal K: invalid Ed25519 curve point")

	// Wrong formatted Ks
	for i := 0; i < k; i++ {
		ballot := electionTypes.EncryptedBallot{electionTypes.Ciphertext{
			K: KsMarshalled[i],
			C: []byte("fakeVoteC"),
		},
		}
		election.PublicBulletinBoard.CastVote("dummyUser"+strconv.Itoa(i), ballot)
	}

	service = updateService(election, dummyId)

	handler.service = &service

	err = handler.handleStartShuffle(dummyId)
	require.EqualError(t, err, "failed to make tx: failed to get shuffled ballots:"+
		" failed to get X, Y: failed to get points: failed to unmarshal C: invalid Ed25519 curve point")

	// Valid Ballots, bad election.PubKey
	for i := 0; i < k; i++ {
		ballot := electionTypes.EncryptedBallot{electionTypes.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		},
		}
		election.PublicBulletinBoard.CastVote("dummyUser"+strconv.Itoa(i), ballot)
	}

	service = updateService(election, dummyId)

	handler.service = &service

	// Wrong shuffle signer
	election.Pubkey = pubKey

	service = updateService(election, dummyId)
	handler.service = &service

	handler.shuffleSigner = fake.NewBadSigner()

	err = handler.handleStartShuffle(dummyId)
	require.EqualError(t, err, fake.Err("failed to make tx: Could not sign the shuffle "))

	// Bad common signer :
	service = updateService(election, dummyId)

	handler.service = &service
	handler.shuffleSigner = fake.NewSigner()

	// Bad manager

	handler.txmngr = fakeManager{}

	err = handler.handleStartShuffle(dummyId)
	require.EqualError(t, err, fake.Err("failed to make tx: failed to use manager"))

	manager := signed.NewManager(fake.NewSigner(), &evotingController.Client{
		Nonce: 0,
		Blocks: FakeBlockStore{
			getErr:  nil,
			lastErr: nil,
		},
	})

	handler.txmngr = manager

	// Bad pool :

	service = updateService(election, dummyId)
	badPool := FakePool{err: fakeErr,
		service: &service}
	handler.p = &badPool
	handler.service = &service

	err = handler.handleStartShuffle(dummyId)
	require.EqualError(t, err, "failed to add transaction to the pool: fake error")

	// Valid, basic scenario : (all errors fixed)
	fakePool := FakePool{service: &service}

	handler.service = &service
	handler.p = &fakePool

	err = handler.handleStartShuffle(dummyId)
	require.NoError(t, err)

	// Threshold is reached :
	election.ShuffleThreshold = 0
	service = updateService(election, dummyId)
	fakePool = FakePool{service: &service}
	handler.service = &service

	err = handler.handleStartShuffle(dummyId)
	require.NoError(t, err)

	// Service not working :
	election.ShuffleThreshold = 1
	service = FakeService{
		err:        nil,
		election:   &election,
		electionId: electionTypes.ID(dummyId),
		status:     true,
		context:    serdecontext,
	}
	fakePool = FakePool{service: &service}
	service.status = false
	handler.service = &service
	err = handler.handleStartShuffle(dummyId)
	// all transactions got denied
	require.NoError(t, err)

	// Shuffle already started:
	shuffledBallots := append(electionTypes.EncryptedBallots{}, election.PublicBulletinBoard.Ballots...)
	election.ShuffleInstances = append(election.ShuffleInstances, electionTypes.ShuffleInstance{ShuffledBallots: shuffledBallots})

	election.ShuffleThreshold = 2

	service = updateService(election, dummyId)
	fakePool = FakePool{service: &service}
	handler = *NewHandler(handler.me, &service, &fakePool, manager, handler.shuffleSigner, serdecontext)

	err = handler.handleStartShuffle(dummyId)
	require.NoError(t, err)
}

// -----------------------------------------------------------------------------
// Utility functions
func updateService(election electionTypes.Election, dummyId string) FakeService {
	return FakeService{
		err:        nil,
		election:   &election,
		electionId: electionTypes.ID(dummyId),
		context:    serdecontext,
	}
}

func initValidHandler(dummyId string) Handler {
	handler := Handler{}

	election := initFakeElection(dummyId)

	service := FakeService{
		err:        nil,
		election:   &election,
		electionId: electionTypes.ID(dummyId),
		status:     true,
		context:    serdecontext,
	}
	fakePool := FakePool{service: &service}

	handler.service = &service
	handler.p = &fakePool
	handler.me = fake.NewAddress(0)
	handler.shuffleSigner = fake.NewSigner()
	handler.txmngr = signed.NewManager(fake.NewSigner(), &evotingController.Client{
		Nonce: 0,
		Blocks: FakeBlockStore{
			getErr:  nil,
			lastErr: nil,
		},
	})
	handler.context = serdecontext

	return handler
}

func initFakeElection(electionId string) electionTypes.Election {
	k := 3
	KsMarshalled, CsMarshalled, pubKey := fakeKCPointsMarshalled(k)
	election := electionTypes.Election{
		ElectionID:          electionId,
		AdminID:             "dummyAdminID",
		Status:              electionTypes.Closed,
		Pubkey:              pubKey,
		PublicBulletinBoard: electionTypes.PublicBulletinBoard{},
		ShuffleInstances:    []electionTypes.ShuffleInstance{},
		DecryptedBallots:    nil,
		ShuffleThreshold:    1,
		BallotSize:          1,
	}

	for i := 0; i < k; i++ {
		ballot := electionTypes.EncryptedBallot{electionTypes.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		},
		}
		election.PublicBulletinBoard.CastVote("dummyUser"+strconv.Itoa(i), ballot)
	}
	return election
}

func fakeKCPointsMarshalled(k int) ([][]byte, [][]byte, kyber.Point) {
	RandomStream := suite.RandomStream()
	h := suite.Scalar().Pick(RandomStream)
	pubKey := suite.Point().Mul(h, nil)

	KsMarshalled := make([][]byte, 0, k)
	CsMarshalled := make([][]byte, 0, k)

	for i := 0; i < k; i++ {
		// Embed the message into a curve point
		message := "Ballot" + strconv.Itoa(i)
		M := suite.Point().Embed([]byte(message), random.New())

		// ElGamal-encrypt the point to produce ciphertext (K,C).
		k := suite.Scalar().Pick(random.New()) // ephemeral private key
		K := suite.Point().Mul(k, nil)         // ephemeral DH public key
		S := suite.Point().Mul(k, pubKey)      // ephemeral DH shared secret
		C := S.Add(S, M)                       // message blinded with secret

		Kmarshalled, _ := K.MarshalBinary()
		Cmarshalled, _ := C.MarshalBinary()

		KsMarshalled = append(KsMarshalled, Kmarshalled)
		CsMarshalled = append(CsMarshalled, Cmarshalled)
	}
	return KsMarshalled, CsMarshalled, pubKey
}

// FakeProof
// - implements ordering.Proof
type FakeProof struct {
	key   []byte
	value []byte
}

// GetKey implements ordering.Proof. It returns the key associated to the proof.
func (f FakeProof) GetKey() []byte {
	return f.key
}

// GetValue implements ordering.Proof. It returns the value associated to the
// proof if the key exists, otherwise it returns nil.
func (f FakeProof) GetValue() []byte {
	return f.value
}

//
// Fake Service
//

type FakeService struct {
	err        error
	election   *electionTypes.Election
	electionId electionTypes.ID
	status     bool
	channel    chan ordering.Event
	context    serde.Context
}

func (f FakeService) GetProof(key []byte) (ordering.Proof, error) {
	electionIDBuff, _ := hex.DecodeString(string(f.electionId))

	if bytes.Equal(key, electionIDBuff) {
		if f.election == nil {
			return nil, f.err
		}

		electionBuff, err := f.election.Serialize(f.context)
		if err != nil {
			return nil, xerrors.Errorf("failed to serialize election: %v", err)
		}

		proof := FakeProof{
			key:   key,
			value: electionBuff,
		}
		return proof, f.err
	}

	return nil, f.err
}

func (f FakeService) GetStore() store.Readable {
	return nil
}

func (f *FakeService) AddTx(tx FakeTransaction) {
	results := make([]validation.TransactionResult, 3)

	results[0] = FakeTransactionResult{
		status:      true,
		message:     "",
		transaction: FakeTransaction{nonce: 10, id: []byte("dummyId1")},
	}

	results[1] = FakeTransactionResult{
		status:      true,
		message:     "",
		transaction: FakeTransaction{nonce: 11, id: []byte("dummyId2")},
	}

	results[2] = FakeTransactionResult{
		status:      f.status,
		message:     "",
		transaction: tx,
	}

	f.status = true

	f.channel <- ordering.Event{
		Index:        0,
		Transactions: results,
	}
	close(f.channel)

}

func (f *FakeService) Watch(ctx context.Context) <-chan ordering.Event {
	f.channel = make(chan ordering.Event, 100)
	return f.channel
}

func (f FakeService) Close() error {
	return f.err
}

//
// Fake Pool
//

type FakePool struct {
	err         error
	transaction FakeTransaction
	service     *FakeService
}

func (f FakePool) SetPlayers(players mino.Players) error {
	return nil
}

func (f FakePool) AddFilter(filter pool.Filter) {
}

func (f FakePool) Len() int {
	return 0
}

func (f *FakePool) Add(transaction txn.Transaction) error {
	newTx := FakeTransaction{
		nonce: transaction.GetNonce(),
		id:    transaction.GetID(),
	}

	f.transaction = newTx
	f.service.AddTx(newTx)

	return f.err
}

func (f FakePool) Remove(transaction txn.Transaction) error {
	return nil
}

func (f FakePool) Gather(ctx context.Context, config pool.Config) []txn.Transaction {
	return nil
}

func (f FakePool) Close() error {
	return nil
}

//
// Fake Transaction
//

type FakeTransaction struct {
	nonce uint64
	id    []byte
}

func (f FakeTransaction) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

func (f FakeTransaction) Fingerprint(writer io.Writer) error {
	return nil
}

func (f FakeTransaction) GetID() []byte {
	return f.id
}

func (f FakeTransaction) GetNonce() uint64 {
	return f.nonce
}

func (f FakeTransaction) GetIdentity() access.Identity {
	return nil
}

func (f FakeTransaction) GetArg(key string) []byte {
	return nil
}

//
// Fake TransactionResult
//

type FakeTransactionResult struct {
	status      bool
	message     string
	transaction FakeTransaction
}

func (f FakeTransactionResult) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

func (f FakeTransactionResult) GetTransaction() txn.Transaction {
	return f.transaction
}

func (f FakeTransactionResult) GetStatus() (bool, string) {
	return f.status, f.message
}

//
// Fake Result
//

type FakeResult struct {
}

func (f FakeResult) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

func (f FakeResult) Fingerprint(writer io.Writer) error {
	return nil
}

func (f FakeResult) GetTransactionResults() []validation.TransactionResult {
	results := make([]validation.TransactionResult, 1)

	results[0] = FakeTransactionResult{
		status:      true,
		message:     "",
		transaction: FakeTransaction{nonce: 10},
	}

	return results
}

//
// Fake BlockLink
//

type FakeBlockLink struct {
}

func (f FakeBlockLink) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

func (f FakeBlockLink) Fingerprint(writer io.Writer) error {
	return nil
}

func (f FakeBlockLink) GetHash() orderingTypes.Digest {
	return orderingTypes.Digest{}
}

func (f FakeBlockLink) GetFrom() orderingTypes.Digest {
	return orderingTypes.Digest{}
}

func (f FakeBlockLink) GetTo() orderingTypes.Digest {
	return orderingTypes.Digest{}
}

func (f FakeBlockLink) GetPrepareSignature() crypto.Signature {
	return nil
}

func (f FakeBlockLink) GetCommitSignature() crypto.Signature {
	return nil
}

func (f FakeBlockLink) GetChangeSet() authority.ChangeSet {
	return nil
}

func (f FakeBlockLink) GetBlock() orderingTypes.Block {

	result := FakeResult{}

	block, _ := orderingTypes.NewBlock(result)
	return block
}

func (f FakeBlockLink) Reduce() orderingTypes.Link {
	return nil
}

//
// Fake BlockStore
//

type FakeBlockStore struct {
	getErr  error
	lastErr error
}

func (f FakeBlockStore) Len() uint64 {
	return 0
}

func (f FakeBlockStore) Store(link orderingTypes.BlockLink) error {
	return nil
}

func (f FakeBlockStore) Get(id orderingTypes.Digest) (orderingTypes.BlockLink, error) {
	return FakeBlockLink{}, f.getErr
}

func (f FakeBlockStore) GetByIndex(index uint64) (orderingTypes.BlockLink, error) {
	return nil, nil
}

func (f FakeBlockStore) GetChain() (orderingTypes.Chain, error) {
	return nil, nil
}

func (f FakeBlockStore) Last() (orderingTypes.BlockLink, error) {
	return FakeBlockLink{}, f.lastErr
}

func (f FakeBlockStore) Watch(ctx context.Context) <-chan orderingTypes.BlockLink {
	return nil
}

func (f FakeBlockStore) WithTx(transaction store.Transaction) blockstore.BlockStore {
	return nil
}
