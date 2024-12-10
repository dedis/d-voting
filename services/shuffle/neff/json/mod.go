package json

import (
	"go.dedis.ch/d-voting/services/shuffle/neff/types"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

func init() {
	types.RegisterMessageFormat(serde.FormatJSON, NewMsgFormat())
}

type Address []byte

type StartShuffle struct {
	FormId    string
	Addresses []Address
}

type EndShuffle struct {
}

type Message struct {
	StartShuffle *StartShuffle `json:",omitempty"`
	EndShuffle   *EndShuffle   `json:",omitempty"`
}

// MsgFormat is the engine to encode and decode SHUFFLE messages in JSON format.
//
// - implements serde.FormatEngine
type MsgFormat struct {
	suite suites.Suite
}

func NewMsgFormat() MsgFormat {
	return MsgFormat{
		suite: suites.MustFind("Ed25519"),
	}
}

// Encode implements serde.FormatEngine. It returns the serialized data for the
// message in JSON format.
func (f MsgFormat) Encode(ctx serde.Context, msg serde.Message) ([]byte, error) {
	var m Message

	switch in := msg.(type) {

	case types.StartShuffle:

		addrs := make([]Address, len(in.GetAddresses()))
		for i, addr := range in.GetAddresses() {
			data, err := addr.MarshalText()
			if err != nil {
				return nil, xerrors.Errorf("couldn't marshal address: %v", err)
			}

			addrs[i] = data
		}

		startShuffle := StartShuffle{
			FormId:    in.GetFormId(),
			Addresses: addrs,
		}
		m = Message{StartShuffle: &startShuffle}

	case types.EndShuffle:
		endShuffle := EndShuffle{}
		m = Message{EndShuffle: &endShuffle}

	default:
		return nil, xerrors.Errorf("unsupported message of type '%T'", msg)
	}

	data, err := ctx.Marshal(m)
	if err != nil {
		return nil, xerrors.Errorf("couldn't marshal: %v", err)
	}

	return data, nil
}

// Decode implements serde.FormatEngine. It populates the message from the JSON
// data if appropriate, otherwise it returns an error.
func (f MsgFormat) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	m := Message{}
	err := ctx.Unmarshal(data, &m)
	if err != nil {
		return nil, xerrors.Errorf("couldn't deserialize message: %v", err)
	}

	if m.StartShuffle != nil {
		return f.decodeStartShuffle(ctx, m.StartShuffle)
	}

	if m.EndShuffle != nil {
		return f.decodeEndShuffle(ctx, m.EndShuffle)
	}

	return nil, xerrors.New("message is empty")
}

func (f MsgFormat) decodeStartShuffle(ctx serde.Context, startShuffle *StartShuffle) (serde.Message, error) {
	factory := ctx.GetFactory(types.AddrKey{})

	fac, ok := factory.(mino.AddressFactory)
	if !ok {
		return nil, xerrors.Errorf("invalid factory of type '%T'", factory)
	}

	addrs := make([]mino.Address, len(startShuffle.Addresses))
	for i, addr := range startShuffle.Addresses {
		addrs[i] = fac.FromText(addr)
	}

	s := types.NewStartShuffle(startShuffle.FormId, addrs)

	return s, nil
}

func (f MsgFormat) decodeEndShuffle(ctx serde.Context, endShuffle *EndShuffle) (serde.Message, error) {
	e := types.NewEndShuffle()

	return e, nil
}
