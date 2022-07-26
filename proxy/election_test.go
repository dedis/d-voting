package proxy

import (
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http/httptest"
	"testing"

	"github.com/dedis/d-voting/contracts/evoting"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/serde"
	sjson "go.dedis.ch/dela/serde/json"
)

type LightContextEngine struct {
	serde.ContextEngine
}

func (LightContextEngine) GetFormat() serde.Format {
	return types.LightElectionJSONFormat
}

func BenchmarkElectionsGET(b *testing.B) {
	b.StopTimer()
	rand.Seed(0)

	ciphervoteFac := types.CiphervoteFactory{}
	electionFac := types.NewElectionFactory(ciphervoteFac, fakeAuthorityFactory{})
	ctx := serde.NewContext(LightContextEngine{ContextEngine: sjson.NewContext()})

	// ctx = sjson.NewContext()

	md := types.ElectionsMetadata{
		ElectionsIDs: types.ElectionIDs{},
	}

	ctx2 := sjson.NewContext()

	data := map[string][]byte{}

	for i := 0; i < 50; i++ {
		electionIDBuff := make([]byte, 8)

		_, err := rand.Read(electionIDBuff)
		require.NoError(b, err)

		electionID := hex.EncodeToString(electionIDBuff)

		elec := types.Election{
			ElectionID:       electionID,
			Status:           0,
			Pubkey:           nil,
			Suffragia:        types.Suffragia{},
			ShuffleInstances: make([]types.ShuffleInstance, 0),
			DecryptedBallots: nil,
			ShuffleThreshold: 0,
			Roster:           fake.Authority{},
		}

		md.ElectionsIDs.Add(electionID)

		electionBuff, err := elec.Serialize(ctx2)
		require.NoError(b, err)

		data[string(electionIDBuff)] = electionBuff
	}

	mdJSON, err := json.Marshal(md)
	require.NoError(b, err)

	data[evoting.ElectionsMetadataKey] = mdJSON

	e := election{
		orderingSvc: fakeService{data: data},
		context:     ctx,
		electionFac: electionFac,
	}

	rr := httptest.NewRecorder()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		e.Elections(rr, nil)
	}
}

// -----------------------------------------------------------------------------
// Utility functions

type fakeService struct {
	ordering.Service
	data map[string][]byte
}

func (f fakeService) GetProof(key []byte) (ordering.Proof, error) {
	proof := fakeProof{
		key:   key,
		value: f.data[string(key)],
	}

	return proof, nil
}

// fakeProof is a fake Proof
//
// - implements ordering.Proof
type fakeProof struct {
	key   []byte
	value []byte
}

func (f fakeProof) GetKey() []byte {
	return f.key
}

func (f fakeProof) GetValue() []byte {
	return f.value
}

type fakeAuthorityFactory struct {
	serde.Factory
}

func (f fakeAuthorityFactory) AuthorityOf(ctx serde.Context, rosterBuf []byte) (authority.Authority, error) {
	fakeAuthority := fakeAuthority{}
	return fakeAuthority, nil
}

type fakeAuthority struct {
	authority.Authority
}
