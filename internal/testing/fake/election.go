package fake

import (
	"strconv"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
)

var suite = suites.MustFind("Ed25519")

func NewElection(electionID string) types.Election {
	k := 3
	KsMarshalled, CsMarshalled, pubKey := NewKCPointsMarshalled(k)
	pubKeyMarshalled, _ := pubKey.MarshalBinary()

	election := types.Election{
		Title:            "dummyTitle",
		ElectionID:       electionID,
		AdminID:          "dummyAdminID",
		Status:           types.Closed,
		Pubkey:           pubKeyMarshalled,
		EncryptedBallots: types.EncryptedBallots{},
		ShuffleInstances: []types.ShuffleInstance{},
		DecryptedBallots: nil,
		ShuffleThreshold: 1,
	}

	for i := 0; i < k; i++ {
		ballot := types.Ciphertext{
			K: KsMarshalled[i],
			C: CsMarshalled[i],
		}
		election.EncryptedBallots.CastVote("dummyUser"+strconv.Itoa(i), ballot)
	}
	return election
}

func NewKCPointsMarshalled(k int) ([][]byte, [][]byte, kyber.Point) {
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
