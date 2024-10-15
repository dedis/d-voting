package types

import (
	"crypto/sha256"

	"go.dedis.ch/dela/serde"
	"go.dedis.ch/dela/serde/registry"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

// suffragiaFormat contains the supported formats for the form. Right now
// only JSON is supported.
var suffragiaFormat = registry.NewSimpleRegistry()

// RegisterSuffragiaFormat registers the engine for the provided format
func RegisterSuffragiaFormat(format serde.Format, engine serde.FormatEngine) {
	suffragiaFormat.Register(format, engine)
}

type Suffragia struct {
	VoterIDs    []string
	Ciphervotes []Ciphervote
}

// Serialize implements the serde.Message
func (s Suffragia) Serialize(ctx serde.Context) ([]byte, error) {
	format := suffragiaFormat.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, s)
	if err != nil {
		return nil, xerrors.Errorf("failed to encode form: %v", err)
	}

	return data, nil
}

// CastVote adds a new vote and its associated user or updates a user's vote.
func (s *Suffragia) CastVote(voterID string, ciphervote Ciphervote) {
	for i, u := range s.VoterIDs {
		if u == voterID {
			s.Ciphervotes[i] = ciphervote
			return
		}
	}

	s.VoterIDs = append(s.VoterIDs, voterID)
	s.Ciphervotes = append(s.Ciphervotes, ciphervote.Copy())
}

// Hash returns the hash of this list of ballots.
func (s *Suffragia) Hash(ctx serde.Context) ([]byte, error) {
	h := sha256.New()
	for i, u := range s.VoterIDs {
		h.Write([]byte(u))
		buf, err := s.Ciphervotes[i].Serialize(ctx)
		if err != nil {
			return nil, xerrors.Errorf("couldn't serialize ciphervote: %v", err)
		}
		h.Write(buf)
	}
	return h.Sum(nil), nil
}

// CiphervotesFromPairs transforms two parallel lists of EGPoints to a list of
// Ciphervotes.
func CiphervotesFromPairs(X, Y [][]kyber.Point) ([]Ciphervote, error) {
	if len(X) != len(Y) {
		return nil, xerrors.Errorf("X and Y must have same length: %d != %d",
			len(X), len(Y))
	}

	if len(X) == 0 {
		return nil, xerrors.Errorf("ElGamal pairs are empty")
	}

	NQ := len(X)   // sequence size
	k := len(X[0]) // number of votes
	res := make([]Ciphervote, k)

	for i := 0; i < k; i++ {
		x := make([]kyber.Point, NQ)
		y := make([]kyber.Point, NQ)

		for j := 0; j < NQ; j++ {
			x[j] = X[j][i]
			y[j] = Y[j][i]
		}

		ciphervote, err := ciphervoteFromPairs(x, y)
		if err != nil {
			return nil, xerrors.Errorf("failed to init from ElGamal pairs: %v", err)
		}

		res[i] = ciphervote
	}

	return res, nil
}

// ciphervoteFromPairs transforms two parallel lists of EGPoints to a list of
// ElGamal pairs.
func ciphervoteFromPairs(ks []kyber.Point, cs []kyber.Point) (Ciphervote, error) {
	if len(ks) != len(cs) {
		return Ciphervote{}, xerrors.Errorf("ks and cs must have same length: %d != %d",
			len(ks), len(cs))
	}

	res := make(Ciphervote, len(ks))

	for i := range ks {
		res[i] = EGPair{
			K: ks[i],
			C: cs[i],
		}
	}

	return res, nil
}
