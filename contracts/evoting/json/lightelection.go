package json

import (
	"github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	"golang.org/x/xerrors"
)

// lightElectionFormat defines how the election messages are encoded/decoded
// using the JSON format.
//
// - implements serde.FormatEngine
type lightElectionFormat struct{}

// Encode implements serde.FormatEngine
func (lightElectionFormat) Encode(ctx serde.Context, message serde.Message) ([]byte, error) {
	return nil, xerrors.Errorf("encoding of a light election not supported")
}

// Decode implements serde.FormatEngine
func (lightElectionFormat) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	var electionJSON LightElectionJSON

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
		Configuration: types.Configuration{
			MainTitle: electionJSON.Configuration.MainTitle,
		},
		ElectionID: electionJSON.ElectionID,
		Status:     types.Status(electionJSON.Status),
		Pubkey:     pubKey,
		BallotSize: electionJSON.BallotSize,
	}, nil
}

// LightConfiguration represents what is need from the configuration in the
// light election.
type LightConfiguration struct {
	MainTitle string
}

// LightElectionJSON defines the Election in the JSON format
type LightElectionJSON struct {
	Configuration LightConfiguration

	// ElectionID is the hex-encoded SHA256 of the transaction ID that creates
	// the election
	ElectionID string

	Status uint16
	Pubkey []byte `json:"Pubkey,omitempty"`

	// BallotSize represents the total size in bytes of one ballot. It is used
	// to pad smaller ballots such that all  ballots cast have the same size
	BallotSize int
}
