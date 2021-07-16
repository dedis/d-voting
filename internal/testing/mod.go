package testing

import (
	"reflect"
	"strings"
	"testing"

	"github.com/golang/protobuf/descriptor"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

// CoverProtoMessage triggers a test on the message definition to force the
// coverage to be taken in account.
func CoverProtoMessage(t *testing.T, message proto.Message) error {
	buffer, err := proto.Marshal(message)
	require.NoError(t, err)
	err = proto.Unmarshal(buffer, message)
	require.NoError(t, err)
	proto.Merge(message, message)
	proto.DiscardUnknown(message)
	require.NotNil(t, message.String())
	message.(descriptor.Message).Descriptor()

	// Run the getters
	tt := reflect.TypeOf(message)
	for i := 0; i < tt.NumMethod(); i++ {
		m := tt.Method(i).Name
		if strings.HasPrefix(m, "Get") {
			reflect.ValueOf(message).MethodByName(m).Call(nil)
			reflect.Zero(tt).MethodByName(m).Call(nil)
		}
	}

	return nil
}
