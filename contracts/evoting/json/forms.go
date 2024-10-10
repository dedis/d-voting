package json

import (
	"encoding/hex"
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

// formFormat defines how the form messages are encoded/decoded using
// the JSON format.
//
// - implements serde.FormatEngine
type formFormat struct{}

// Encode implements serde.FormatEngine
func (formFormat) Encode(ctx serde.Context, message serde.Message) ([]byte, error) {
	switch m := message.(type) {
	case types.Form:

		var pubkey []byte
		var err error

		if m.Pubkey != nil {
			pubkey, err = m.Pubkey.MarshalBinary()
			if err != nil {
				return nil, xerrors.Errorf("failed to marshall public key: %v", err)
			}
		}

		suffragias := make([]string, len(m.SuffragiaIDs))
		for i, suf := range m.SuffragiaIDs {
			suffragias[i] = hex.EncodeToString(suf)
		}

		suffragiaHashes := make([]string, len(m.SuffragiaHashes))
		for i, sufH := range m.SuffragiaHashes {
			suffragiaHashes[i] = hex.EncodeToString(sufH)
		}

		shuffleInstances, err := encodeShuffleInstances(ctx, m.ShuffleInstances)
		if err != nil {
			return nil, xerrors.Errorf("failed to encode shuffle instances: %v", err)
		}

		rosterBuf, err := m.Roster.Serialize(ctx)
		if err != nil {
			return nil, xerrors.Errorf("failed to serialize roster: %v", err)
		}

		pubsharesUnits, err := encodePubsharesUnits(m.PubsharesUnits)
		if err != nil {
			return nil, xerrors.Errorf("failed to encode submissions of pubShares: %v",
				err)
		}

		formJSON := FormJSON{
			Configuration:    m.Configuration,
			FormID:           m.FormID,
			Status:           uint16(m.Status),
			Pubkey:           pubkey,
			BallotSize:       m.BallotSize,
			Suffragias:       suffragias,
			SuffragiaHashes:  suffragiaHashes,
			BallotCount:      m.BallotCount,
			ShuffleInstances: shuffleInstances,
			ShuffleThreshold: m.ShuffleThreshold,
			PubsharesUnits:   pubsharesUnits,
			DecryptedBallots: m.DecryptedBallots,
			RosterBuf:        rosterBuf,
		}

		buff, err := ctx.Marshal(&formJSON)
		if err != nil {
			return nil, xerrors.Errorf("failed to marshal form: %v", err)
		}

		return buff, nil
	default:
		return nil, xerrors.Errorf("unknown format: %T", message)
	}
}

// Decode implements serde.FormatEngine
func (formFormat) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	var formJSON FormJSON

	err := ctx.Unmarshal(data, &formJSON)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal form: %v", err)
	}

	var pubKey kyber.Point

	if formJSON.Pubkey != nil {
		pubKey = suite.Point()
		err = pubKey.UnmarshalBinary(formJSON.Pubkey)
		if err != nil {
			return nil, xerrors.Errorf("failed to unmarshal pubkey: %v", err)
		}
	}

	suffragias := make([][]byte, len(formJSON.Suffragias))
	for i, suff := range formJSON.Suffragias {
		suffragias[i], err = hex.DecodeString(suff)
		if err != nil {
			return nil, xerrors.Errorf("failed to decode suffragia-address: %v", err)
		}
	}

	suffragiaHashes := make([][]byte, len(formJSON.SuffragiaHashes))
	for i, suffH := range formJSON.SuffragiaHashes {
		suffragiaHashes[i], err = hex.DecodeString(suffH)
		if err != nil {
			return nil, xerrors.Errorf("failed to decode suffragia-hash: %v", err)
		}
	}

	shuffleInstances, err := decodeShuffleInstances(ctx, formJSON.ShuffleInstances)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode shuffle instances: %v", err)
	}

	fac := ctx.GetFactory(ctypes.RosterKey{})
	rosterFac, ok := fac.(authority.Factory)
	if !ok {
		return nil, xerrors.Errorf("failed to get roster factory: %T", fac)
	}

	roster, err := rosterFac.AuthorityOf(ctx, formJSON.RosterBuf)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode roster: %v", err)
	}

	pubSharesSubmissions, err := decodePubSharesUnits(formJSON.PubsharesUnits)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode pubShares submissions: %v", err)
	}

	return types.Form{
		Configuration:    formJSON.Configuration,
		FormID:           formJSON.FormID,
		Status:           types.Status(formJSON.Status),
		Pubkey:           pubKey,
		BallotSize:       formJSON.BallotSize,
		SuffragiaIDs:     suffragias,
		SuffragiaHashes:  suffragiaHashes,
		BallotCount:      formJSON.BallotCount,
		ShuffleInstances: shuffleInstances,
		ShuffleThreshold: formJSON.ShuffleThreshold,
		PubsharesUnits:   pubSharesSubmissions,
		DecryptedBallots: formJSON.DecryptedBallots,
		Roster:           roster,
	}, nil
}

// FormJSON defines the Form in the JSON format
type FormJSON struct {
	Configuration types.Configuration

	// FormID is the hex-encoded SHA256 of the transaction ID that creates
	// the form
	FormID string

	AdminID string
	Status  uint16
	Pubkey  []byte `json:"Pubkey,omitempty"`

	// BallotSize represents the total size in bytes of one ballot. It is used
	// to pad smaller ballots such that all  ballots cast have the same size
	BallotSize int

	// Suffragias are the hex-encoded addresses of the Suffragia storages.
	Suffragias []string

	// BallotCount represents the total number of ballots cast.
	BallotCount uint32

	// SuffragiaHashes are the hex-encoded sha256-hashes of the ballots
	// in every Suffragia.
	SuffragiaHashes []string

	// ShuffleInstances is all the shuffles, along with their proof and identity
	// of shuffler.
	ShuffleInstances []ShuffleInstanceJSON

	// ShuffleThreshold is set based on the roster. We save it so we do not have
	// to compute it based on the roster each time we need it.
	ShuffleThreshold int

	PubsharesUnits PubsharesUnitsJSON

	DecryptedBallots []types.Ballot

	// roster is set when the form is created based on the current
	// roster of the node stored in the global state. The roster will not change
	// during a form and will be used for DKG and Neff. Its type is
	// authority.Authority.

	RosterBuf []byte
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

func encodeShuffleInstances(ctx serde.Context,
	shuffleInstances []types.ShuffleInstance) ([]ShuffleInstanceJSON, error) {

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

func encodeShuffleInstance(ctx serde.Context,
	shuffleInstance types.ShuffleInstance) (ShuffleInstanceJSON, error) {

	var res ShuffleInstanceJSON
	shuffledBallots := make([]json.RawMessage, len(shuffleInstance.ShuffledBallots))

	for i, shuffledBallot := range shuffleInstance.ShuffledBallots {
		buff, err := shuffledBallot.Serialize(ctx)
		if err != nil {
			return res, xerrors.Errorf("failed to serialize ciphervote: %v", err)
		}

		shuffledBallots[i] = buff
	}

	res = ShuffleInstanceJSON{
		ShuffledBallots:   shuffledBallots,
		ShuffleProofs:     shuffleInstance.ShuffleProofs,
		ShufflerPublicKey: shuffleInstance.ShufflerPublicKey,
	}

	return res, nil
}

func decodeShuffleInstances(ctx serde.Context,
	shuffleInstancesJSON []ShuffleInstanceJSON) ([]types.ShuffleInstance, error) {

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

func decodeShuffleInstance(ctx serde.Context,
	shuffleInstanceJSON ShuffleInstanceJSON) (types.ShuffleInstance, error) {

	var res types.ShuffleInstance
	fac := ctx.GetFactory(types.CiphervoteKey{})

	factory, ok := fac.(types.CiphervoteFactory)
	if !ok {
		return res, xerrors.Errorf("invalid ciphervote factory: '%T'", fac)
	}

	shuffledBallots := make([]types.Ciphervote, len(shuffleInstanceJSON.ShuffledBallots))

	for i, ciphervoteJSON := range shuffleInstanceJSON.ShuffledBallots {
		msg, err := factory.Deserialize(ctx, ciphervoteJSON)
		if err != nil {
			return res, xerrors.Errorf("failed to deserialize shuffle instance json: %v", err)
		}

		ciphervote, ok := msg.(types.Ciphervote)
		if !ok {
			return res, xerrors.Errorf("wrong type: '%T'", msg)
		}

		shuffledBallots[i] = ciphervote
	}

	res = types.ShuffleInstance{
		ShuffledBallots:   shuffledBallots,
		ShuffleProofs:     shuffleInstanceJSON.ShuffleProofs,
		ShufflerPublicKey: shuffleInstanceJSON.ShufflerPublicKey,
	}

	return res, nil
}

// PubsharesUnitJSON is the JSON representation of a submission of pubShares by
// one node.The first dimension is the pubshares marshalled into bytes.
type PubsharesUnitJSON [][][]byte

// PubsharesUnitsJSON defines the JSON representation of the
// types.PubsharesUnits as used in the form.
type PubsharesUnitsJSON struct {
	// PubsharesJSON contains all the pubShares submitted.
	PubsharesJSON []PubsharesUnitJSON
	PubKeys       [][]byte
	Indexes       []int
}

func encodePubsharesUnits(units types.PubsharesUnits) (
	PubsharesUnitsJSON, error) {
	var unitsJSON PubsharesUnitsJSON

	submissionsJSON := make([]PubsharesUnitJSON, len(units.Pubshares))

	for i, submission := range units.Pubshares {
		submissionsJSON[i] = make([][][]byte, len(submission))

		for i2, ballotShares := range submission {
			submissionsJSON[i][i2] = make([][]byte, len(ballotShares))

			for i3, pubShare := range ballotShares {
				pubShareMarshaled, err := pubShare.MarshalBinary()
				if err != nil {
					return unitsJSON, xerrors.Errorf("could not marshal public share: %v", err)
				}

				submissionsJSON[i][i2][i3] = pubShareMarshaled
			}
		}
	}

	unitsJSON.Indexes = units.Indexes
	unitsJSON.PubKeys = units.PubKeys
	unitsJSON.PubsharesJSON = submissionsJSON

	return unitsJSON, nil
}

func decodePubSharesUnits(unitsJSON PubsharesUnitsJSON) (types.PubsharesUnits, error) {
	var units types.PubsharesUnits

	submissions := make([]types.PubsharesUnit, len(unitsJSON.PubsharesJSON))

	for i, submissionJSON := range unitsJSON.PubsharesJSON {
		submissions[i] = make([][]types.Pubshare, len(submissionJSON))

		for i2, ballotSharesJSON := range submissionJSON {
			submissions[i][i2] = make([]types.Pubshare, len(ballotSharesJSON))

			for i3, pubShareJSON := range ballotSharesJSON {
				pubShare := suite.Point()
				err := pubShare.UnmarshalBinary(pubShareJSON)
				if err != nil {
					return units, xerrors.Errorf("could not unmarshal public share: %v", err)
				}

				submissions[i][i2][i3] = pubShare
			}
		}
	}

	units.Indexes = unitsJSON.Indexes
	units.PubKeys = unitsJSON.PubKeys
	units.Pubshares = submissions

	return units, nil
}
