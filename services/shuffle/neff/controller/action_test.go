package controller

import (
	"io"
	"testing"

	"github.com/c4dt/dela/cli/node"
	"github.com/stretchr/testify/require"
)

func TestInitAction_Execute(t *testing.T) {
	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    make(node.FlagSet),
		Out:      io.Discard,
	}

	action := InitAction{}

	err := action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve shuffle: couldn't find "+
		"dependency for 'shuffle.Shuffle'")
}
