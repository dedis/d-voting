package json

import (
	"encoding/json"
	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

// transactionFormat defines the JSON format of a transaction
//
// - implements serde.FormatEngine
type transactionFormat struct{}

// Encode implements serde.FormatEngine
func (transactionFormat) Encode(ctx serde.Context, msg serde.Message) ([]byte, error) {
	var m TransactionJSON

	switch t := msg.(type) {
	case types.CreateElection:
		ce := CreateElectionJSON{
			Configuration: t.Configuration,
			AdminID:       t.AdminID,
		}

		m = TransactionJSON{CreateElection: &ce}
	case types.OpenElection:
		oe := OpenElectionJSON{
			ElectionID: t.ElectionID,
		}

		m = TransactionJSON{OpenElection: &oe}
	case types.CastVote:
		ballot, err := t.Ballot.Serialize(ctx)
		if err != nil {
			return nil, xerrors.Errorf("failed to serialize ballot: %v", err)
		}

		cv := CastVoteJSON{
			ElectionID: t.ElectionID,
			UserID:     t.UserID,
			Ciphervote: ballot,
		}

		m = TransactionJSON{CastVote: &cv}
	case types.CloseElection:
		ce := CloseElectionJSON{
			ElectionID: t.ElectionID,
			UserID:     t.UserID,
		}

		m = TransactionJSON{CloseElection: &ce}
	case types.ShuffleBallots:
		ciphervotes := make([]json.RawMessage, len(t.ShuffledBallots))

		for i, ciphervote := range t.ShuffledBallots {
			buf, err := ciphervote.Serialize(ctx)
			if err != nil {
				return nil, xerrors.Errorf("failed to serialize ciphervote: %v", err)
			}

			ciphervotes[i] = buf
		}

		sb := ShuffleBallotsJSON{
			ElectionID:   t.ElectionID,
			Round:        t.Round,
			Ciphervotes:  ciphervotes,
			RandomVector: t.RandomVector,
			Proof:        t.Proof,
			Signature:    t.Signature,
			PublicKey:    t.PublicKey,
		}

		m = TransactionJSON{ShuffleBallots: &sb}
	case types.RegisterPubShares:
		pubShares := make([][][]byte, len(t.Pubshares))

		for i, ballotShares := range t.Pubshares {
			pubShares[i] = make([][]byte, len(ballotShares))
			for i2, share := range ballotShares {
				pubShare, err := share.MarshalBinary()
				if err != nil {
					return nil, xerrors.Errorf("failed to marshal pubShare: %v", err)
				}

				pubShares[i][i2] = pubShare
			}
		}

		rp := RegisterPubSharesJSON{
			ElectionID: t.ElectionID,
			Index:      t.Index,
			PubShares:  pubShares,
			Signature:  t.Signature,
			PublicKey:  t.PublicKey,
		}

		m = TransactionJSON{RegisterPubShares: &rp}
	case types.DecryptBallots:
		db := DecryptBallotsJSON{
			ElectionID:       t.ElectionID,
			UserID:           t.UserID,
			DecryptedBallots: t.DecryptedBallots,
		}

		m = TransactionJSON{DecryptBallots: &db}
	case types.CancelElection:
		ce := CancelElectionJSON{
			ElectionID: t.ElectionID,
			UserID:     t.UserID,
		}

		m = TransactionJSON{CancelElection: &ce}
	default:
		return nil, xerrors.Errorf("unknown type: '%T", msg)
	}

	data, err := ctx.Marshal(m)
	if err != nil {
		return nil, xerrors.Errorf("failed to marshal transactionJSON: %v", err)
	}

	return data, nil
}

// Decode implements serde.FormatEngine
func (transactionFormat) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	m := TransactionJSON{}

	err := ctx.Unmarshal(data, &m)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal transaction json: %v", err)
	}

	switch {
	case m.CreateElection != nil:
		return types.CreateElection{
			Configuration: m.CreateElection.Configuration,
			AdminID:       m.CreateElection.AdminID,
		}, nil
	case m.OpenElection != nil:
		return types.OpenElection{
			ElectionID: m.OpenElection.ElectionID,
		}, nil
	case m.CastVote != nil:
		msg, err := decodeCastVote(ctx, *m.CastVote)
		if err != nil {
			return nil, xerrors.Errorf("failed to decode cast vote: %v", err)
		}

		return msg, nil
	case m.CloseElection != nil:
		return types.CloseElection{
			ElectionID: m.CloseElection.ElectionID,
			UserID:     m.CloseElection.UserID,
		}, nil
	case m.ShuffleBallots != nil:
		msg, err := decodeShuffleBallots(ctx, *m.ShuffleBallots)
		if err != nil {
			return nil, xerrors.Errorf("failed to decode shuffle ballots: %v", err)
		}

		return msg, nil
	case m.RegisterPubShares != nil:
		msg, err := decodeRegisterPubShares(*m.RegisterPubShares)
		if err != nil {
			return nil, xerrors.Errorf("failed to decode register pubShares: %v", err)
		}

		return msg, nil
	case m.DecryptBallots != nil:
		return types.DecryptBallots{
			ElectionID:       m.DecryptBallots.ElectionID,
			UserID:           m.DecryptBallots.UserID,
			DecryptedBallots: m.DecryptBallots.DecryptedBallots,
		}, nil
	case m.CancelElection != nil:
		return types.CancelElection{
			ElectionID: m.CancelElection.ElectionID,
			UserID:     m.CancelElection.UserID,
		}, nil
	}

	return nil, xerrors.Errorf("empty type: %s", data)
}

// TransactionJSON is the JSON message that wraps the different kinds of
// transactions.
type TransactionJSON struct {
	CreateElection    *CreateElectionJSON    `json:",omitempty"`
	OpenElection      *OpenElectionJSON      `json:",omitempty"`
	CastVote          *CastVoteJSON          `json:",omitempty"`
	CloseElection     *CloseElectionJSON     `json:",omitempty"`
	ShuffleBallots    *ShuffleBallotsJSON    `json:",omitempty"`
	RegisterPubShares *RegisterPubSharesJSON `json:",omitempty"`
	DecryptBallots    *DecryptBallotsJSON    `json:",omitempty"`
	CancelElection    *CancelElectionJSON    `json:",omitempty"`
}

// CreateElectionJSON is the JSON representation of a CreateElection transaction
type CreateElectionJSON struct {
	Configuration types.Configuration
	AdminID       string
}

// OpenElectionJSON is the JSON representation of a OpenElection transaction
type OpenElectionJSON struct {
	ElectionID string
}

// CastVoteJSON is the JSON representation of a CastVote transaction
type CastVoteJSON struct {
	ElectionID string
	UserID     string
	Ciphervote json.RawMessage
}

// CloseElectionJSON is the JSON representation of a CloseElection transaction
type CloseElectionJSON struct {
	ElectionID string
	UserID     string
}

// ShuffleBallotsJSON is the JSON representation of a ShuffleBallots transaction
type ShuffleBallotsJSON struct {
	ElectionID   string
	Round        int
	Ciphervotes  []json.RawMessage
	RandomVector types.RandomVector
	Proof        []byte
	Signature    []byte
	PublicKey    []byte
}

type RegisterPubSharesJSON struct {
	ElectionID string
	Index      int
	PubShares  PubsharesUnitJSON
	Signature  []byte
	PublicKey  []byte
}

// DecryptBallotsJSON is the JSON representation of a CombineShares transaction
type DecryptBallotsJSON struct {
	ElectionID       string
	UserID           string
	DecryptedBallots []types.Ballot
}

// CancelElectionJSON is the JSON representation of a CancelElection transaction
type CancelElectionJSON struct {
	ElectionID string
	UserID     string
}

func decodeCastVote(ctx serde.Context, m CastVoteJSON) (serde.Message, error) {
	factory := ctx.GetFactory(types.CiphervoteKey{})
	if factory == nil {
		return nil, xerrors.Errorf("missing ciphervote factory")
	}

	msg, err := factory.Deserialize(ctx, m.Ciphervote)
	if err != nil {
		return nil, xerrors.Errorf("failed to deserialize ciphervote: %v", err)
	}

	ciphervote, ok := msg.(types.Ciphervote)
	if !ok {
		return nil, xerrors.Errorf("invalid ciphervote: '%T'", msg)
	}

	return types.CastVote{
		ElectionID: m.ElectionID,
		UserID:     m.UserID,
		Ballot:     ciphervote,
	}, nil
}

func decodeShuffleBallots(ctx serde.Context, m ShuffleBallotsJSON) (serde.Message, error) {
	factory := ctx.GetFactory(types.CiphervoteKey{})
	if factory == nil {
		return nil, xerrors.Errorf("missing ciphervote factory")
	}

	ciphervotes := make([]types.Ciphervote, len(m.Ciphervotes))

	for i, buff := range m.Ciphervotes {
		msg, err := factory.Deserialize(ctx, buff)
		if err != nil {
			return nil, xerrors.Errorf("failed to deserialize ciphervote: %v", err)
		}

		ciphervote, ok := msg.(types.Ciphervote)
		if !ok {
			return nil, xerrors.Errorf("invalid ciphervote: '%T'", msg)
		}

		ciphervotes[i] = ciphervote
	}

	return types.ShuffleBallots{
		ElectionID:      m.ElectionID,
		Round:           m.Round,
		ShuffledBallots: ciphervotes,
		RandomVector:    m.RandomVector,
		Proof:           m.Proof,
		Signature:       m.Signature,
		PublicKey:       m.PublicKey,
	}, nil
}

func decodeRegisterPubShares(m RegisterPubSharesJSON) (serde.Message, error) {
	pubShares := make([][]types.Pubshare, len(m.PubShares))

	for i, ballotShares := range m.PubShares {
		pubShares[i] = make([]types.Pubshare, len(ballotShares))

		for i2, share := range ballotShares {
			pubShare := suite.Point()
			err := pubShare.UnmarshalBinary(share)
			if err != nil {
				return nil, xerrors.Errorf("could not unmarshal pubShare: %v", err)
			}

			pubShares[i][i2] = pubShare
		}
	}

	return types.RegisterPubShares{
		ElectionID: m.ElectionID,
		Index:      m.Index,
		Pubshares:  pubShares,
		Signature:  m.Signature,
		PublicKey:  m.PublicKey,
	}, nil
}
