package integration

import (
	"github.com/dedis/d-voting/services/shuffle"
	"go.dedis.ch/dela/core/txn/signed"
	"golang.org/x/xerrors"
)

func initShuffle(nodes []dVotingCosiDela) (shuffle.Actor, error) {
	var sActor shuffle.Actor

	for _, node := range nodes {
		client := client{
			srvc: node.GetOrdering(),
			mgr:  node.GetValidationSrv(),
		}

		var err error
		shuffler := node.GetShuffle()

		sActor, err = shuffler.Listen(signed.NewManager(node.GetShuffleSigner(), client))
		if err != nil {
			return nil, xerrors.Errorf("failed to init Shuffle: %v", err)
		}
	}

	return sActor, nil
}
