package fake

import (
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/xerrors"
)

var suite = suites.MustFind("Ed25519")

func NewElection(electionID string) types.Election {
	k := 3
	Ks, Cs, pubKey := NewKCPointsMarshalled(k)

	election := types.Election{
		Configuration: types.Configuration{
			MainTitle: "dummyTitle",
		},
		ElectionID: electionID,
		AdminID:    "dummyAdminID",
		Status:     types.Closed,
		Pubkey:     pubKey,
		Suffragia: types.Suffragia{
			Ciphervotes: []types.Ciphervote{},
		},
		ShuffleInstances: []types.ShuffleInstance{},
		DecryptedBallots: nil,
		ShuffleThreshold: 1,
	}

	for i := 0; i < k; i++ {
		ballot := types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}
		election.Suffragia.CastVote("dummyUser"+strconv.Itoa(i), types.Ciphervote{ballot})
	}

	return election
}

func NewKCPointsMarshalled(k int) ([]kyber.Point, []kyber.Point, kyber.Point) {
	RandomStream := suite.RandomStream()
	h := suite.Scalar().Pick(RandomStream)
	pubKey := suite.Point().Mul(h, nil)

	Ks := make([]kyber.Point, 0, k)
	Cs := make([]kyber.Point, 0, k)

	for i := 0; i < k; i++ {
		// Embed the message into a curve point
		message := "Ballot" + strconv.Itoa(i)
		M := suite.Point().Embed([]byte(message), random.New())

		// ElGamal-encrypt the point to produce ciphertext (K,C).
		k := suite.Scalar().Pick(random.New()) // ephemeral private key
		K := suite.Point().Mul(k, nil)         // ephemeral DH public key
		S := suite.Point().Mul(k, pubKey)      // ephemeral DH shared secret
		C := S.Add(S, M)                       // message blinded with secret

		Ks = append(Ks, K)
		Cs = append(Cs, C)
	}
	return Ks, Cs, pubKey
}

// BasicConfiguration returns a basic election configuration
var BasicConfiguration = types.Configuration{
	MainTitle: "electionTitle",
	Scaffold: []types.Subject{
		{
			ID:       encodeID("aa"),
			Title:    "subject1",
			Order:    nil,
			Subjects: nil,
			Selects: []types.Select{
				{
					ID:      encodeID("bb"),
					Title:   "Select your favorite snacks",
					MaxN:    3,
					MinN:    0,
					Choices: []string{"snickers", "mars", "vodka", "babibel"},
				},
			},
			Ranks: []types.Rank{},
			Texts: nil,
		},
		{
			ID:       encodeID("dd"),
			Title:    "subject2",
			Order:    nil,
			Subjects: nil,
			Selects:  nil,
			Ranks:    nil,
			Texts: []types.Text{
				{
					ID:        encodeID("ee"),
					Title:     "dissertation",
					MaxN:      1,
					MinN:      1,
					MaxLength: 3,
					Regex:     "",
					Choices:   []string{"write yes in your language"},
				},
			},
		},
	},
}

func encodeID(ID string) types.ID {
	return types.ID(base64.StdEncoding.EncodeToString([]byte(ID)))
}

// NewConn returns a new fake connection
func NewConn(network, address string, timeout time.Duration) (net.Conn, error) {
	return &Conn{}, nil
}

// Conn simulates a connection that would be done with the Unikernel
//
// - implements net.Conn
type Conn struct {
}

func (f *Conn) Read(b []byte) (n int, err error) {
	n = copy(b, []byte("OK"))

	return n, nil
}

// Write implements net.Conn. It does what the Unikernel is expected to do
func (f *Conn) Write(b []byte) (n int, err error) {

	if len(b) < 4+4+1 {
		return 0, xerrors.Errorf("not enough bytes")
	}

	numChunks := binary.LittleEndian.Uint32(b[:4])
	numNodes := binary.LittleEndian.Uint32(b[4:8])

	folder := string(b[8:])

	err = combineShares(folder, int(numChunks), int(numNodes))
	if err != nil {
		return 0, xerrors.Errorf("failed to combine shares: %v", err)
	}

	return len(b), nil
}

// combineShares implements what the Unikernel is expected to do
func combineShares(folder string, numChunks, numNodes int) error {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return xerrors.Errorf("failed to read dir: %v", err)
	}

	expectedSize := 32 * numChunks * numNodes

	for _, file := range files {
		decrypted, err := os.Create(filepath.Join(folder, file.Name()+".decrypted"))
		if err != nil {
			return xerrors.Errorf("failed to create decrypted: %v", err)
		}

		if file.IsDir() {
			continue
		}

		if !strings.HasPrefix(file.Name(), "ballot_") {
			continue
		}

		if file.Size() != int64(expectedSize) {
			return xerrors.Errorf("unexpected file size: %d != %d: %s",
				file.Size(), expectedSize, filepath.Join(folder, file.Name()))
		}

		buf, err := os.ReadFile(filepath.Join(folder, file.Name()))
		if err != nil {
			return xerrors.Errorf("failed to read file: %v", err)
		}

		for c := 0; c < int(numChunks); c++ {
			pubshares := make([]*share.PubShare, numNodes)

			for n := 0; n < int(numNodes); n++ {
				base := (32 * c * int(numNodes)) + 32*n

				v := suite.Point()

				err = v.UnmarshalBinary(buf[base : base+32])
				if err != nil {
					return xerrors.Errorf("failed to unmarshal point: %v", err)
				}

				pubshares[n] = &share.PubShare{
					I: n,
					V: v,
				}
			}

			res, err := share.RecoverCommit(suite, pubshares, int(numNodes), int(numNodes))
			if err != nil {
				return xerrors.Errorf("failed to recover commit: %v", err)
			}

			_, err = res.MarshalTo(decrypted)
			if err != nil {
				return xerrors.Errorf("failed to marshal to file: %v", err)
			}
		}

		decrypted.Close()
	}

	return nil
}

func (f *Conn) Close() error {
	panic("not implemented")
}

func (f *Conn) LocalAddr() net.Addr {
	panic("not implemented")
}

func (f *Conn) RemoteAddr() net.Addr {
	panic("not implemented")
}

func (f *Conn) SetDeadline(t time.Time) error {
	panic("not implemented")
}

func (f *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (f *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}
