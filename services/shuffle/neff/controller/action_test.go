package controller

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli/node"
)

func TestInitAction_Execute(t *testing.T) {
	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    make(node.FlagSet),
		Out:      ioutil.Discard,
	}

	action := InitAction{}

	err := action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve shuffle: couldn't find "+
		"dependency for 'shuffle.Shuffle'")
}
