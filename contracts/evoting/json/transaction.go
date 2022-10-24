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
	case types.CreateForm:
		ce := CreateFormJSON{
			Configuration: t.Configuration,
			AdminID:       t.AdminID,
		}

		m = TransactionJSON{CreateForm: &ce}
	case types.OpenForm:
		oe := OpenFormJSON{
			FormID: t.FormID,
		}

		m = TransactionJSON{OpenForm: &oe}
	case types.CastVote:
		ballot, err := t.Ballot.Serialize(ctx)
		if err != nil {
			return nil, xerrors.Errorf("failed to serialize ballot: %v", err)
		}

		cv := CastVoteJSON{
			FormID: t.FormID,
			UserID:     t.UserID,
			Ciphervote: ballot,
		}

		m = TransactionJSON{CastVote: &cv}
	case types.CloseForm:
		ce := CloseFormJSON{
			FormID: t.FormID,
			UserID:     t.UserID,
		}

		m = TransactionJSON{CloseForm: &ce}
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
			FormID:   t.FormID,
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
			FormID: t.FormID,
			Index:      t.Index,
			PubShares:  pubShares,
			Signature:  t.Signature,
			PublicKey:  t.PublicKey,
		}

		m = TransactionJSON{RegisterPubShares: &rp}
	case types.CombineShares:
		db := CombineSharesJSON{
			FormID: t.FormID,
			UserID:     t.UserID,
		}

		m = TransactionJSON{CombineShares: &db}
	case types.CancelForm:
		ce := CancelFormJSON{
			FormID: t.FormID,
			UserID:     t.UserID,
		}

		m = TransactionJSON{CancelForm: &ce}
	case types.DeleteForm:
		de := DeleteFormJSON{
			FormID: t.FormID,
		}

		m = TransactionJSON{DeleteForm: &de}
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
	case m.CreateForm != nil:
		return types.CreateForm{
			Configuration: m.CreateForm.Configuration,
			AdminID:       m.CreateForm.AdminID,
		}, nil
	case m.OpenForm != nil:
		return types.OpenForm{
			FormID: m.OpenForm.FormID,
		}, nil
	case m.CastVote != nil:
		msg, err := decodeCastVote(ctx, *m.CastVote)
		if err != nil {
			return nil, xerrors.Errorf("failed to decode cast vote: %v", err)
		}

		return msg, nil
	case m.CloseForm != nil:
		return types.CloseForm{
			FormID: m.CloseForm.FormID,
			UserID:     m.CloseForm.UserID,
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
	case m.CombineShares != nil:
		return types.CombineShares{
			FormID: m.CombineShares.FormID,
			UserID:     m.CombineShares.UserID,
		}, nil
	case m.CancelForm != nil:
		return types.CancelForm{
			FormID: m.CancelForm.FormID,
			UserID:     m.CancelForm.UserID,
		}, nil
	case m.DeleteForm != nil:
		return types.DeleteForm{
			FormID: m.DeleteForm.FormID,
		}, nil
	}

	return nil, xerrors.Errorf("empty type: %s", data)
}

// TransactionJSON is the JSON message that wraps the different kinds of
// transactions.
type TransactionJSON struct {
	CreateForm    *CreateFormJSON    `json:",omitempty"`
	OpenForm      *OpenFormJSON      `json:",omitempty"`
	CastVote          *CastVoteJSON          `json:",omitempty"`
	CloseForm     *CloseFormJSON     `json:",omitempty"`
	ShuffleBallots    *ShuffleBallotsJSON    `json:",omitempty"`
	RegisterPubShares *RegisterPubSharesJSON `json:",omitempty"`
	CombineShares     *CombineSharesJSON     `json:",omitempty"`
	CancelForm    *CancelFormJSON    `json:",omitempty"`
	DeleteForm    *DeleteFormJSON    `json:",omitempty"`
}

// CreateFormJSON is the JSON representation of a CreateForm transaction
type CreateFormJSON struct {
	Configuration types.Configuration
	AdminID       string
}

// OpenFormJSON is the JSON representation of a OpenForm transaction
type OpenFormJSON struct {
	FormID string
}

// CastVoteJSON is the JSON representation of a CastVote transaction
type CastVoteJSON struct {
	FormID string
	UserID     string
	Ciphervote json.RawMessage
}

// CloseFormJSON is the JSON representation of a CloseForm transaction
type CloseFormJSON struct {
	FormID string
	UserID     string
}

// ShuffleBallotsJSON is the JSON representation of a ShuffleBallots transaction
type ShuffleBallotsJSON struct {
	FormID   string
	Round        int
	Ciphervotes  []json.RawMessage
	RandomVector types.RandomVector
	Proof        []byte
	Signature    []byte
	PublicKey    []byte
}

type RegisterPubSharesJSON struct {
	FormID string
	Index      int
	PubShares  PubsharesUnitJSON
	Signature  []byte
	PublicKey  []byte
}

// CombineSharesJSON is the JSON representation of a CombineShares transaction
type CombineSharesJSON struct {
	FormID string
	UserID     string
}

// CancelFormJSON is the JSON representation of a CancelForm transaction
type CancelFormJSON struct {
	FormID string
	UserID     string
}

// DeleteFormJSON is the JSON representation of a DeleteForm transaction
type DeleteFormJSON struct {
	FormID string
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
		FormID: m.FormID,
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
		FormID:      m.FormID,
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
		FormID: m.FormID,
		Index:      m.Index,
		Pubshares:  pubShares,
		Signature:  m.Signature,
		PublicKey:  m.PublicKey,
	}, nil
}
