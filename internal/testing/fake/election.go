package fake

import (
	"strconv"

	"github.com/c4dt/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/xerrors"
)

var suite = suites.MustFind("Ed25519")

func NewForm(ctx serde.Context, snapshot store.Snapshot, formID string) (types.Form, error) {
	k := 3
	Ks, Cs, pubKey := NewKCPointsMarshalled(k)

	form := types.Form{
		Configuration: types.Configuration{
			Title: types.Title{
				En:  "dummyTitle",
				Fr:  "",
				De:  "",
				URL: "",
			},
		},
		FormID:           formID,
		Status:           types.Closed,
		Pubkey:           pubKey,
		ShuffleInstances: []types.ShuffleInstance{},
		DecryptedBallots: nil,
		ShuffleThreshold: 1,
	}

	for i := 0; i < k; i++ {
		ballot := types.EGPair{
			K: Ks[i],
			C: Cs[i],
		}
		err := form.CastVote(ctx, snapshot, "dummyUser"+strconv.Itoa(i), types.Ciphervote{ballot})
		if err != nil {
			return form, xerrors.Errorf("couldn't cast vote: %v", err)
		}
	}

	return form, nil
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
	Title: types.Title{En: "formTitle", Fr: "", De: "", URL: ""},
	Scaffold: []types.Subject{
		{
			ID:       "aa",
			Title:    types.Title{En: "subject1", Fr: "", De: "", URL: ""},
			Order:    nil,
			Subjects: nil,
			Selects: []types.Select{
				{
					ID:      "bb",
					Title:   types.Title{En: "Select your favorite snacks", Fr: "", De: "", URL: ""},
					MaxN:    3,
					MinN:    0,
					Choices: []types.Choice{{Choice: "snickers", URL: ""}, {Choice: "mars", URL: ""}, {Choice: "vodka", URL: ""}, {Choice: "babibel", URL: ""}},
				},
			},
			Ranks: []types.Rank{},
			Texts: nil,
		},
		{
			ID:       "dd",
			Title:    types.Title{En: "subject2", Fr: "", De: "", URL: ""},
			Order:    nil,
			Subjects: nil,
			Selects:  nil,
			Ranks:    nil,
			Texts: []types.Text{
				{
					ID:        "ee",
					Title:     types.Title{En: "dissertation", Fr: "", De: "", URL: ""},
					MaxN:      1,
					MinN:      1,
					MaxLength: 3,
					Regex:     "",
					Choices:   []types.Choice{{Choice: "write yes in your language", URL: ""}},
				},
			},
		},
	},
}
