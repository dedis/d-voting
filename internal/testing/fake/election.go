package fake

import (
	"strconv"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
)

var suite = suites.MustFind("Ed25519")

func NewForm(formID string) types.Form {
	k := 3
	Ks, Cs, pubKey := NewKCPointsMarshalled(k)

	form := types.Form{
		Configuration: types.Configuration{
			MainTitle: "dummyTitle",
		},
		FormID: formID,
		Status: types.Closed,
		Pubkey: pubKey,
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
		form.Suffragia.CastVote("dummyUser"+strconv.Itoa(i), types.Ciphervote{ballot})
	}

	return form
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

// BasicConfiguration returns a basic form configuration
var BasicConfiguration = types.Configuration{
	MainTitle: "formTitle",
	Scaffold: []types.Subject{
		{
			ID:       "aa",
			Title:    "subject1",
			Order:    nil,
			Subjects: nil,
			Selects: []types.Select{
				{
					ID:      "bb",
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
			ID:       "dd",
			Title:    "subject2",
			Order:    nil,
			Subjects: nil,
			Selects:  nil,
			Ranks:    nil,
			Texts: []types.Text{
				{
					ID:        "ee",
					Title:     "dissertation",
					MaxN:      1,
					MinN:      1,
					MaxLength: 3,
					Regex:     "",
					Choices:   []string{"write yes in your language"},
				},
			},
		},
		{
			ID:       "ff",
			Title:    "subject3",
			Order:    nil,
			Subjects: nil,
			Selects:  nil,
			Ranks: []types.Rank{
				{
					ID:      "gg",
					Title:   "Rank your favorite snacks",
					MaxN:    4,
					MinN:    2,
					Choices: []string{"snickers", "mars", "vodka", "babibel"},
				},
			},
			Texts: nil,
		},
	},
}
