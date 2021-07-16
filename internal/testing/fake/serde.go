package fake

import (
	"crypto/rand"
	"encoding/json"
	"io"

	"go.dedis.ch/dela/serde"
)

func init() {
	// A random value is injected to prevent hardcoded value in the tests.
	fakeFmtValue = make([]byte, 8)
	rand.Read(fakeFmtValue)
}

const (
	// GoodFormat should register working format engines.
	GoodFormat = serde.Format("FakeGood")

	// BadFormat should register non-working format engines.
	BadFormat = serde.Format("FakeBad")

	// MsgFormat should register an engine for fake.Message.
	MsgFormat = serde.Format("FakeMsg")
)

var fakeFmtValue []byte

// GetFakeFormatValue returns the value of the fake format serialization.
func GetFakeFormatValue() []byte {
	return append([]byte{}, fakeFmtValue...)
}

// Message is a fake implementation if a serde message.
//
// - implements serde.Message
type Message struct {
	Digest []byte
}

// Fingerprint implements serde.Fingerprinter.
func (m Message) Fingerprint(w io.Writer) error {
	w.Write(m.Digest)
	return nil
}

// Serialize implements serde.Message.
func (m Message) Serialize(ctx serde.Context) ([]byte, error) {
	return ctx.Marshal(struct{}{})
}

// MessageFactory is a fake implementation of a serde factory.
//
// - implements serde.Factory
type MessageFactory struct {
	err error
}

// NewBadMessageFactory returns a new message factory that returns an error.
func NewBadMessageFactory() MessageFactory {
	return MessageFactory{err: fakeErr}
}

// Deserialize implements serde.Factory.
func (f MessageFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	return Message{}, f.err
}

// Format is a fake format engine implementation.
//
// - implements serde.FormatEngine
type Format struct {
	err  error
	Msg  serde.Message
	Call *Call
}

// NewBadFormat returns a new format engine that always return an error.
func NewBadFormat() Format {
	return Format{err: fakeErr}
}

// Encode implements serde.FormatEngine.
func (f Format) Encode(ctx serde.Context, m serde.Message) ([]byte, error) {
	f.Call.Add(ctx, m)

	return GetFakeFormatValue(), f.err
}

// Decode implements serde.FormatEngine.
func (f Format) Decode(ctx serde.Context, data []byte) (serde.Message, error) {
	f.Call.Add(ctx, data)
	return f.Msg, f.err
}

// MessageFormat is a format engine to encode and decode fake messages.
//
// - implements serde.FormatEngine
type MessageFormat struct{}

// NewMsgFormat creates a new format.
func NewMsgFormat() MessageFormat {
	return MessageFormat{}
}

// Encode implements serde.FormatEngine.
func (f MessageFormat) Encode(ctx serde.Context, m serde.Message) ([]byte, error) {
	return Message{}.Serialize(ctx)
}

// Decode implements serde.FormatEngine.
func (f MessageFormat) Decode(serde.Context, []byte) (serde.Message, error) {
	return Message{}, nil
}

// ContextEngine is a fake implementation of a serde context engine that is
// using JSON as the underlying marshaler.
//
// - implements serde.ContextEngine
type ContextEngine struct {
	Count  *Counter
	format serde.Format
	err    error
}

// NewContext returns a new serde context.
func NewContext() serde.Context {
	return serde.NewContext(ContextEngine{
		format: GoodFormat,
	})
}

// NewContextWithFormat returns a new serde context that is using the provided
// format.
func NewContextWithFormat(f serde.Format) serde.Context {
	return serde.NewContext(ContextEngine{
		format: f,
	})
}

// NewBadContext returns a new serde context that produces errors.
func NewBadContext() serde.Context {
	return serde.NewContext(ContextEngine{
		format: BadFormat,
		err:    fakeErr,
	})
}

// NewBadContextWithDelay returns a new serde context that produces errors after
// some calls.
func NewBadContextWithDelay(delay int) serde.Context {
	return serde.NewContext(ContextEngine{
		Count:  &Counter{Value: delay},
		format: BadFormat,
		err:    fakeErr,
	})
}

// NewMsgContext returns a new serde context that always produces a fake
// message.
func NewMsgContext() serde.Context {
	return serde.NewContext(ContextEngine{
		format: MsgFormat,
	})
}

// GetFormat implements serde.ContextEngine.
func (ctx ContextEngine) GetFormat() serde.Format {
	return ctx.format
}

// Marshal implements serde.ContextEngine.
func (ctx ContextEngine) Marshal(m interface{}) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	if !ctx.Count.Done() {
		ctx.Count.Decrease()
		return data, nil
	}

	return data, ctx.err
}

// Unmarshal implements serde.ContextEngine.
func (ctx ContextEngine) Unmarshal(data []byte, m interface{}) error {
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}

	if !ctx.Count.Done() {
		ctx.Count.Decrease()
		return nil
	}

	return ctx.err
}
