package integration

import (
	"github.com/dedis/d-voting/services/dkg"
	"go.dedis.ch/dela/core/txn"
	"golang.org/x/xerrors"
)

func initDkg(nodes []dVotingCosiDela, formID []byte, m txn.Manager) (dkg.Actor, error) {
	var actor dkg.Actor
	var err error

	for _, node := range nodes {
		d := node.(dVotingNode).GetDkg()

		// put Listen in a goroutine to optimize for speed
		actor, err = d.Listen(formID, m)
		if err != nil {
			return nil, xerrors.Errorf("failed to GetDkg: %v", err)
		}
	}

	_, err = actor.Setup()
	if err != nil {
		return nil, xerrors.Errorf("failed to Setup: %v", err)
	}

	return actor, nil
}
