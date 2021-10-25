package controller

import (
	"io/ioutil"
	"testing"

	"github.com/dedis/d-voting/internal/testing/fake"
	"github.com/dedis/d-voting/services/shuffle/neff"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/dela/cli/node"
	"go.dedis.ch/dela/core/ordering/cosipbft/blockstore"
)

func TestInitAction_Execute(t *testing.T) {

	ctx := node.Context{
		Injector: node.NewInjector(),
		Flags:    make(node.FlagSet),
		Out:      ioutil.Discard,
	}

	action := initAction{}

	err := action.Execute(ctx)
	require.EqualError(t, err, "failed to resolve shuffle: couldn't find dependency for 'shuffle.Shuffle'")

	ctx.Injector.Inject(neff.NewNeffShuffle(fake.Mino{}, fakeService{}, fakePool{}, &blockstore.InDisk{}))
	err = action.Execute(ctx)
	require.NoError(t, err)

}
