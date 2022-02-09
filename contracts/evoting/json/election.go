package json

import (
	"encoding/json"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	ctypes "go.dedis.ch/dela/core/ordering/cosipbft/types"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

var suite = suites.MustFind("Ed25519")

// electionFormat defines how the election messages are encoded/decoded using
// the JSON format.
//
// - implements serde.FormatEngine
type electionFormat struct{}

// Encode implements serde.FormatEngine
func (electionFormat) Encode(ctx serde.Context, message serde.Message) ([]byte, error) {
	switch m := message.(type) {
	case types.Election:

		var pubkey []byte
		var err error

		if m.Pubkey != nil {
			pubkey, err = m.Pubkey.MarshalBinary()
			if err != nil {
				return nil, xerrors.Errorf("failed to marshall public key: %v", err)
			}
		}

		suffragia, err := encodeSuffragia(ctx, m.Suffragia)
		if err != nil {
			return nil, xerrors.Errorf("failed to encode suffragia: %v", err)
		}

		shuffleInstances, err := encodeShuffleInstances(ctx, m.ShuffleInstances)
		if err != nil {
			return nil, xerrors.Errorf("failed to encode shuffle instances: %v", err)
		}

		rosterBuf, err := m.Roster.Serialize(ctx)
		if err != nil {
			return nil, xerrors.Errorf("failed to serialize roster: %v", err)
		}

		electionJSON := ElectionJSON{
			Configuration:    m.Configuration,
			ElectionID:       m.ElectionID,
			AdminID:          m.AdminID,
			Status:           uint16(m.Status),
			Pubkey:           pubkey,
			BallotSize:       m.BallotSize,
			Suffragia:        suffragia,
			ShuffleInstances: shuffleInstances,
			ShuffleThreshold: m.ShuffleThreshold,
			DecryptedBallots: m.DecryptedBallots,
			RosterBuf:        rosterBuf,
		}

		buff, err := ctx.Marshal(&electionJSON)
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal election: %v", err)
		}

		return buff, nil
	default:
		return nil, xerrors.Errorf("unknown format: %T", message)
	}
}

// Decode implements serde.FormatEngine
func (electionFormat) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	var electionJSON ElectionJSON

	err := ctx.Unmarshal(data, &electionJSON)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal election: %v", err)
	}

	var pubKey kyber.Point

	if electionJSON.Pubkey != nil {
		pubKey = suite.Point()
		err = pubKey.UnmarshalBinary(electionJSON.Pubkey)
		if err != nil {
			return nil, xerrors.Errorf("failed to unmarshal pubkey: %v", err)
		}
	}

	suffragia, err := decodeSuffragia(ctx, electionJSON.Suffragia)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode suffragia: %v", err)
	}

	shuffleInstances, err := decodeShuffleInstances(ctx, electionJSON.ShuffleInstances)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode shuffle instances: %v", err)
	}

	fac := ctx.GetFactory(ctypes.RosterKey{})
	rosterFac, ok := fac.(authority.Factory)
	if !ok {
		return nil, xerrors.Errorf("failed to get roster factory: %T", fac)
	}

	roster, err := rosterFac.AuthorityOf(ctx, electionJSON.RosterBuf)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode roster: %v", err)
	}

	return types.Election{
		Configuration:    electionJSON.Configuration,
		ElectionID:       electionJSON.ElectionID,
		AdminID:          electionJSON.AdminID,
		Status:           types.Status(electionJSON.Status),
		Pubkey:           pubKey,
		BallotSize:       electionJSON.BallotSize,
		Suffragia:        suffragia,
		ShuffleInstances: shuffleInstances,
		ShuffleThreshold: electionJSON.ShuffleThreshold,
		DecryptedBallots: electionJSON.DecryptedBallots,
		Roster:           roster,
	}, nil
}

// ElectionJSON defines the Election in the JSON format
type ElectionJSON struct {
	Configuration types.Configuration

	// ElectionID is the hex-encoded SHA256 of the transaction ID that creates
	// the election
	ElectionID string

	AdminID string
	Status  uint16
	Pubkey  []byte `json:"Pubkey,omitempty"`

	// BallotSize represents the total size in bytes of one ballot. It is used
	// to pad smaller ballots such that all  ballots cast have the same size
	BallotSize int

	Suffragia SuffragiaJSON

	// ShuffleInstances is all the shuffles, along with their proof and identity
	// of shuffler.
	ShuffleInstances []ShuffleInstanceJSON

	// ShuffleThreshold is set based on the roster. We save it so we do not have
	// to compute it based on the roster each time we need it.
	ShuffleThreshold int

	DecryptedBallots []types.Ballot

	// roster is set when the election is created based on the current
	// roster of the node stored in the global state. The roster will not change
	// during an election and will be used for DKG and Neff. Its type is
	// authority.Authority.

	RosterBuf []byte
}

// SuffragiaJSON defines the JSON representation of a suffragia.
type SuffragiaJSON struct {
	UserIDs     []string
	Ciphervotes []json.RawMessage
}

func encodeSuffragia(ctx serde.Context, suffragia types.Suffragia) (SuffragiaJSON, error) {
	ciphervotes := make([]json.RawMessage, len(suffragia.Ciphervotes))

	for i, ciphervote := range suffragia.Ciphervotes {
		buff, err := ciphervote.Serialize(ctx)
		if err != nil {
			return SuffragiaJSON{}, xerrors.Errorf("failed to serialize ciphervote: %v", err)
		}

		ciphervotes[i] = buff
	}
	return SuffragiaJSON{
		UserIDs:     suffragia.UserIDs,
		Ciphervotes: ciphervotes,
	}, nil
}

func decodeSuffragia(ctx serde.Context, suffragiaJSON SuffragiaJSON) (types.Suffragia, error) {
	fac := ctx.GetFactory(types.CiphervoteKey{})

	factory, ok := fac.(types.CiphervoteFactory)
	if !ok {
		return types.Suffragia{}, xerrors.Errorf("invalid ciphervote factory: '%T'", fac)
	}

	ciphervotes := make([]types.Ciphervote, len(suffragiaJSON.Ciphervotes))

	for i, ciphervoteJSON := range suffragiaJSON.Ciphervotes {
		msg, err := factory.Deserialize(ctx, ciphervoteJSON)
		if err != nil {
			return types.Suffragia{}, xerrors.Errorf("failed to deserialize ciphervote json: %v", err)
		}

		ciphervote, ok := msg.(types.Ciphervote)
		if !ok {
			return types.Suffragia{}, xerrors.Errorf("wrong type: '%T'", msg)
		}

		ciphervotes[i] = ciphervote
	}

	return types.Suffragia{
		UserIDs:     suffragiaJSON.UserIDs,
		Ciphervotes: ciphervotes,
	}, nil
}

// ShuffleInstanceJSON defines the JSON representation of a shuffle instance
type ShuffleInstanceJSON struct {
	// ShuffledBallots contains the list of shuffled ciphertext for this round
	ShuffledBallots []json.RawMessage

	// ShuffleProofs is the proof of the shuffle for this round
	ShuffleProofs []byte

	// ShufflerPublicKey is the key of the node who made the given shuffle.
	ShufflerPublicKey []byte
}

func encodeShuffleInstances(ctx serde.Context, shuffleInstances []types.ShuffleInstance) ([]ShuffleInstanceJSON, error) {
	res := make([]ShuffleInstanceJSON, len(shuffleInstances))

	for i, shuffleInstance := range shuffleInstances {
		shuffleInstanceJSON, err := encodeShuffleInstance(ctx, shuffleInstance)
		if err != nil {
			return nil, xerrors.Errorf("failed to encode shuffle instance: %v", err)
		}

		res[i] = shuffleInstanceJSON
	}

	return res, nil
}

func encodeShuffleInstance(ctx serde.Context, shuffleInstance types.ShuffleInstance) (ShuffleInstanceJSON, error) {
	shuffledBallots := make([]json.RawMessage, len(shuffleInstance.ShuffledBallots))

	for i, shuffledBallot := range shuffleInstance.ShuffledBallots {
		buff, err := shuffledBallot.Serialize(ctx)
		if err != nil {
			return ShuffleInstanceJSON{}, xerrors.Errorf("failed to serialize ciphervote: %v", err)
		}

		shuffledBallots[i] = buff
	}

	return ShuffleInstanceJSON{
		ShuffledBallots:   shuffledBallots,
		ShuffleProofs:     shuffleInstance.ShuffleProofs,
		ShufflerPublicKey: shuffleInstance.ShufflerPublicKey,
	}, nil
}

func decodeShuffleInstances(ctx serde.Context, shuffleInstancesJSON []ShuffleInstanceJSON) ([]types.ShuffleInstance, error) {
	res := make([]types.ShuffleInstance, len(shuffleInstancesJSON))

	for i, shuffleInstanceJSON := range shuffleInstancesJSON {
		shuffleInstance, err := decodeShuffleInstance(ctx, shuffleInstanceJSON)
		if err != nil {
			return nil, xerrors.Errorf("failed to decode shuffle instance: %v", err)
		}

		res[i] = shuffleInstance
	}

	return res, nil
}

func decodeShuffleInstance(ctx serde.Context, shuffleInstanceJSON ShuffleInstanceJSON) (types.ShuffleInstance, error) {
	fac := ctx.GetFactory(types.CiphervoteKey{})

	factory, ok := fac.(types.CiphervoteFactory)
	if !ok {
		return types.ShuffleInstance{}, xerrors.Errorf("invalid ciphervote factory: '%T'", fac)
	}

	shuffledBallots := make([]types.Ciphervote, len(shuffleInstanceJSON.ShuffledBallots))

	for i, ciphervoteJSON := range shuffleInstanceJSON.ShuffledBallots {
		msg, err := factory.Deserialize(ctx, ciphervoteJSON)
		if err != nil {
			return types.ShuffleInstance{}, xerrors.Errorf("failed to deserialize shuffle instance json: %v", err)
		}

		ciphervote, ok := msg.(types.Ciphervote)
		if !ok {
			return types.ShuffleInstance{}, xerrors.Errorf("wrong type: '%T'", msg)
		}

		shuffledBallots[i] = ciphervote
	}

	return types.ShuffleInstance{
		ShuffledBallots:   shuffledBallots,
		ShuffleProofs:     shuffleInstanceJSON.ShuffleProofs,
		ShufflerPublicKey: shuffleInstanceJSON.ShufflerPublicKey,
	}, nil
}
