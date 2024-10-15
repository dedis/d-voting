package types

import (
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/dela/serde/registry"
	"golang.org/x/xerrors"
)

var msgFormats = registry.NewSimpleRegistry()

// RegisterMessageFormat register the engine for the provided format.
func RegisterMessageFormat(c serde.Format, f serde.FormatEngine) {
	msgFormats.Register(c, f)
}

// StartShuffle is the message the initiator of the SHUFFLE protocol should
// send to all the nodes.
//
// - implements serde.Message
type StartShuffle struct {
	formID    string
	userID    string
	addresses []mino.Address
}

// NewStartShuffle creates a new StartShuffle message.
func NewStartShuffle(formID string, userID string, addresses []mino.Address) StartShuffle {
	return StartShuffle{
		formID:    formID,
		userID:    userID,
		addresses: addresses,
	}
}

// GetFormID returns the formId.
func (s StartShuffle) GetFormID() string {
	return s.formID
}

// GetUserID returns the formId.
func (s StartShuffle) GetUserID() string {
	return s.userID
}

// GetAddresses returns the list of addresses.
func (s StartShuffle) GetAddresses() []mino.Address {
	return append([]mino.Address{}, s.addresses...)
}

// Serialize implements serde.Message. It looks up the format and returns the
// serialized data for the start message.
func (s StartShuffle) Serialize(ctx serde.Context) ([]byte, error) {
	format := msgFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, s)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode StartShuffle message: %v", err)
	}

	return data, nil
}

// EndShuffle is the message the node running during the last round should
// send to all the nodes.
//
// - implements serde.Message
type EndShuffle struct {
}

// NewEndShuffle creates a new EndShuffle message.
func NewEndShuffle() EndShuffle {
	return EndShuffle{}
}

// Serialize implements serde.Message. It looks up the format and returns the
// serialized data for the start message.
func (e EndShuffle) Serialize(ctx serde.Context) ([]byte, error) {
	format := msgFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, e)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode EndShuffle message: %v", err)
	}

	return data, nil
}

// AddrKey is the key for the address factory.
type AddrKey struct{}

// MessageFactory is a message factory for the different SHUFFLE messages.
//
// - implements serde.Factory
type MessageFactory struct {
	addrFactory mino.AddressFactory
}

// NewMessageFactory returns a message factory for the shuffle protocol.
func NewMessageFactory(f mino.AddressFactory) MessageFactory {
	return MessageFactory{
		addrFactory: f,
	}
}

// Deserialize implements serde.Factory.
func (f MessageFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	format := msgFormats.Get(ctx.GetFormat())

	ctx = serde.WithFactory(ctx, AddrKey{}, f.addrFactory)

	msg, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("couldn't decode message: %v", err)
	}

	return msg, nil
}
