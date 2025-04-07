package json

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/d-voting/internal/testing/fake"
	"go.dedis.ch/d-voting/services/shuffle/neff/types"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
)

func TestMessageFormat_StartShuffle_Encode(t *testing.T) {
	format := NewMsgFormat()
	ctx := serde.NewContext(fake.ContextEngine{})

	_, err := format.Encode(fake.NewBadContext(), types.StartShuffle{})
	require.EqualError(t, err, fake.Err("couldn't marshal"))

	_, err = format.Encode(ctx, fake.Message{})
	require.EqualError(t, err, "unsupported message of type 'fake.Message'")

	startShuffle := types.NewStartShuffle("", []mino.Address{fake.NewBadAddress()})

	_, err = format.Encode(ctx, startShuffle)
	require.EqualError(t, err, fake.Err("couldn't marshal address"))

	startShuffle = types.NewStartShuffle("dummyId", []mino.Address{fake.NewAddress(0)})

	data, err := format.Encode(ctx, startShuffle)
	require.NoError(t, err)

	regexp := `{"StartShuffle":{"FormId":"dummyId","Addresses":\["AAAAAA=="\]`
	require.Regexp(t, regexp, string(data))
}

func TestMessageFormat_EndShuffle_Encode(t *testing.T) {
	endShuffle := types.NewEndShuffle()

	format := NewMsgFormat()
	ctx := serde.NewContext(fake.ContextEngine{})

	data, err := format.Encode(ctx, endShuffle)
	require.NoError(t, err)

	regexp := `{"EndShuffle":{}}`
	require.Regexp(t, regexp, string(data))
}

func TestMessageFormat_StartShuffle_Decode(t *testing.T) {
	format := NewMsgFormat()
	ctx := serde.NewContext(fake.ContextEngine{})
	ctx = serde.WithFactory(ctx, types.AddrKey{}, fake.AddressFactory{})

	_, err := format.Decode(fake.NewBadContext(), []byte(`{}`))
	require.EqualError(t, err, fake.Err("couldn't deserialize message"))

	_, err = format.Decode(ctx, []byte(`{}`))
	require.EqualError(t, err, "message is empty")

	badCtx := serde.WithFactory(ctx, types.AddrKey{}, nil)
	_, err = format.Decode(badCtx, []byte(`{"StartShuffle":{}}`))
	require.EqualError(t, err, "invalid factory of type '<nil>'")

	expected := types.NewStartShuffle(
		"dummyId",
		[]mino.Address{fake.NewAddress(0)},
	)

	data, err := format.Encode(ctx, expected)
	require.NoError(t, err)

	startShuffle, err := format.Decode(ctx, data)
	require.NoError(t, err)
	require.Equal(t, expected.GetFormId(), startShuffle.(types.StartShuffle).GetFormId())
	require.Len(t, startShuffle.(types.StartShuffle).GetAddresses(), len(expected.GetAddresses()))

}

func TestMessageFormat_EndShuffle_Decode(t *testing.T) {
	format := NewMsgFormat()
	ctx := serde.NewContext(fake.ContextEngine{})
	ctx = serde.WithFactory(ctx, types.AddrKey{}, fake.AddressFactory{})

	expected := types.NewEndShuffle()

	data, err := format.Encode(ctx, expected)
	require.NoError(t, err)

	_, err = format.Decode(ctx, data)
	require.NoError(t, err)
}
