package json

import (
	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

var suite = suites.MustFind("Ed25519")

func init() {
	types.RegisterFormat(serde.FormatJSON, jsonEngine{})
}

// jsonEngine defines how the election messages are encoded/decoded using the
// JSON format.
//
// - implements serde.FormatEngine
type jsonEngine struct{}

// Encode implements serde.FormatEngine
func (jsonEngine) Encode(ctx serde.Context, message serde.Message) ([]byte, error) {
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

		electionJSON := ElectionJSON{
			Configuration:       m.Configuration,
			ElectionID:          m.ElectionID,
			AdminID:             m.AdminID,
			Status:              uint16(m.Status),
			Pubkey:              pubkey,
			BallotSize:          m.BallotSize,
			PublicBulletinBoard: m.PublicBulletinBoard,
			ShuffleInstances:    m.ShuffleInstances,
			ShuffleThreshold:    m.ShuffleThreshold,
			DecryptedBallots:    m.DecryptedBallots,
			RosterBuf:           m.RosterBuf,
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
func (jsonEngine) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
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

	return types.Election{
		Configuration:       electionJSON.Configuration,
		ElectionID:          electionJSON.ElectionID,
		AdminID:             electionJSON.AdminID,
		Status:              types.Status(electionJSON.Status),
		Pubkey:              pubKey,
		BallotSize:          electionJSON.BallotSize,
		PublicBulletinBoard: electionJSON.PublicBulletinBoard,
		ShuffleInstances:    electionJSON.ShuffleInstances,
		ShuffleThreshold:    electionJSON.ShuffleThreshold,
		DecryptedBallots:    electionJSON.DecryptedBallots,
		RosterBuf:           electionJSON.RosterBuf,
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

	// PublicBulletinBoard is a map from User ID to their ballot EncryptedBallot
	PublicBulletinBoard types.PublicBulletinBoard

	// ShuffleInstances is all the shuffles, along with their proof and identity
	// of shuffler.
	ShuffleInstances []types.ShuffleInstance

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
