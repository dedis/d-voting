package fake

import (
	"encoding/base64"
	"strconv"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
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
