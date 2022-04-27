package evoting

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"github.com/stretchr/testify/require"
)

// This test case simulates a blockchain node that uses the Unikernel to perform
// the smart contract operation. The operation on the smart contract will
// combine the ballots shares and write the result, i.e. the ballot in
// plain-text.
//
// You must setup the SHARED_FOLDER and UK_ADDR env variables with the folder
// shared between the Unikernel and the blockchain node, and the Unikernel
// address respectively. The folder path must be absolute.
func Test_Unikernel(t *testing.T) {
	t.Skip()

	sharedFolder := os.Getenv("SHARED_FOLDER")
	ukAddr := os.Getenv("UK_ADDR")

	require.NotEmpty(t, sharedFolder)
	require.NotEmpty(t, ukAddr)

	n := 10

	nodes, publicKey := computeDKG(t, n)

	subdirBuf := make([]byte, 12)

	rand.Seed(time.Now().UnixNano())

	_, err := rand.Read(subdirBuf)
	require.NoError(t, err)

	subdir := hex.EncodeToString(subdirBuf) + "/"

	folder := filepath.Join(sharedFolder, subdir)

	err = os.MkdirAll(folder, os.ModePerm)
	require.NoError(t, err)

	defer os.RemoveAll(folder)

	// Encrypt secret and compute public shares. We manually encrypt two ballots
	// that are two chunks long each.

	ballot0 := []byte("This message is 55 bytes long, which requires 2 chunks.")

	b0K0, b0C0, remainder := elGamalEncrypt(suite, publicKey, ballot0[:29])
	require.Equal(t, 0, len(remainder))

	b0K1, b0C1, remainder := elGamalEncrypt(suite, publicKey, ballot0[29:])
	require.Equal(t, 0, len(remainder))

	ballot1 := []byte("This is another ballot, which also requires 2 chunks.")

	b1K0, b1C0, remainder := elGamalEncrypt(suite, publicKey, ballot1[:29])
	require.Equal(t, 0, len(remainder))

	b1K1, b1C1, remainder := elGamalEncrypt(suite, publicKey, ballot1[29:])
	require.Equal(t, 0, len(remainder))

	// initialize the pubshares slice according to the number of nodes, ballots,
	// and chunks per ballot.
	pubshares := make([]types.PubsharesUnit, n) // number of nodes
	for i := range pubshares {
		pubshares[i] = make(types.PubsharesUnit, 2) // number of ballots
		for j := range pubshares[i] {
			pubshares[i][j] = make([]types.Pubshare, 2) // chunks per ballot
		}
	}

	// initialize the indexes slice. All shares are in oder from the nodes'
	// indexes.
	indexes := make([]int, n)
	for i := range indexes {
		indexes[i] = i
	}

	// compute and store the pubshares for each node/ballot/chunk
	for i, node := range nodes {
		// > node i, ballot 0, chunk 0
		S := suite.Point().Mul(node.secretShare.V, b0K0)
		v := suite.Point().Sub(b0C0, S)
		pubshares[i][0][0] = v

		// > node i, ballot 0, chunk 1
		S = suite.Point().Mul(node.secretShare.V, b0K1)
		v = suite.Point().Sub(b0C1, S)
		pubshares[i][0][1] = v

		// > node i, ballot 1, chunk 0
		S = suite.Point().Mul(node.secretShare.V, b1K0)
		v = suite.Point().Sub(b1C0, S)
		pubshares[i][1][0] = v

		// > node i, ballot 1, chunk 1
		S = suite.Point().Mul(node.secretShare.V, b1K1)
		v = suite.Point().Sub(b1C1, S)
		pubshares[i][1][1] = v
	}

	units := types.PubsharesUnits{
		Pubshares: pubshares,
		Indexes:   indexes,
	}

	err = exportPubshares(folder, units, 2, 2)
	require.NoError(t, err)

	// We are expecting two files, one for each ballot. A file contains the
	// pubshares, in order, of chunks.

	file0, err := os.ReadFile(filepath.Join(folder, "ballot_0"))
	require.NoError(t, err)

	file1, err := os.ReadFile(filepath.Join(folder, "ballot_1"))
	require.NoError(t, err)

	// size is <kyber point size> * <number of nodes> * <number of chunks>
	require.Len(t, file0, 32*n*2)
	require.Len(t, file1, 32*n*2)

	point := suite.Point()

	// > check file of ballot 0
	for i := 0; i < n; i++ {
		// > ballot 0, node i, chunk 0
		err = point.UnmarshalBinary(file0[i*32 : i*32+32])
		require.NoError(t, err)
		require.True(t, point.Equal(pubshares[i][0][0]))

		// > ballot 0, node i, chunk 1
		base := 32 * n
		err = point.UnmarshalBinary(file0[base+32*i : base+32*i+32])
		require.NoError(t, err)
		require.True(t, point.Equal(pubshares[i][0][1]))
	}

	// > check file of ballot 1
	for i := 0; i < n; i++ {
		// > ballot 1, node i, chunk 0
		err = point.UnmarshalBinary(file1[i*32 : i*32+32])
		require.NoError(t, err)
		require.True(t, point.Equal(pubshares[i][1][0]))

		// > ballot 1, node i, chunk 1
		base := 32 * n
		err = point.UnmarshalBinary(file1[base+32*i : base+32*i+32])
		require.NoError(t, err)
		require.True(t, point.Equal(pubshares[i][1][1]))
	}

	//
	// Unikernel read the result and perform its computation
	//

	const dialTimeout = time.Second * 60
	const writeTimeout = time.Second * 60
	const readTimeout = time.Second * 60

	conn, err := net.DialTimeout("tcp", ukAddr, dialTimeout)
	if err != nil {
		require.NoError(t, err)
	}

	numChunks := 2
	nbrSubmissions := 10

	buf := bytes.Buffer{}

	nc := make([]byte, 4)
	binary.LittleEndian.PutUint32(nc, uint32(numChunks))

	nn := make([]byte, 4)
	binary.LittleEndian.PutUint32(nn, uint32(nbrSubmissions))

	buf.Write(nc)
	buf.Write(nn)

	fmt.Println("sending subdir:", subdir)

	buf.Write([]byte(subdir))

	conn.SetWriteDeadline(time.Now().Add(writeTimeout))

	_, err = buf.WriteTo(conn)
	require.NoError(t, err)

	readRes := make([]byte, 256)

	conn.SetReadDeadline(time.Now().Add(readTimeout))

	//
	// Reading back the Unikernel result
	//

	n, err = conn.Read(readRes)
	require.NoError(t, err)

	readRes = readRes[:n]

	fmt.Println("readRes:", string(readRes))

	rawBallots, err := importBallots(folder, numChunks)
	require.NoError(t, err)

	for _, r := range rawBallots {
		fmt.Printf("r: %v\n", string(r))
	}
}
