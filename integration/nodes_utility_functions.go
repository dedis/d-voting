package integration

import (
	"sync"
	"time"

	"github.com/dedis/d-voting/contracts/evoting/types"
	"golang.org/x/xerrors"
)

func closeNodes(nodes []dVotingCosiDela) error {
	wait := sync.WaitGroup{}
	wait.Add(len(nodes))

	for _, n := range nodes {
		go func(node dVotingNode) {
			defer wait.Done()
			node.GetOrdering().Close()
		}(n.(dVotingNode))
	}

	done := make(chan struct{})

	go func() {
		wait.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(time.Second * 30):
		return xerrors.New("failed to close: timeout")
	}
}


// waitForStatus polls the nodes until they all updated to the expected status
// for the given form. An error is raised if the timeout expires.
func waitForStatus(status types.Status, formFac types.FormFactory,
	formID []byte, nodes []dVotingCosiDela, numNodes int, timeOut time.Duration) error {

	expiration := time.Now().Add(timeOut)

	isOK := func() (bool, error) {
		for _, node := range nodes {
			form, err := getForm(formFac, formID, node.GetOrdering())
			if err != nil {
				return false, xerrors.Errorf("failed to get form: %v", err)
			}

			if form.Status != status {
				return false, nil
			}
		}

		return true, nil
	}

	for {
		if time.Now().After(expiration) {
			return xerrors.New("status check expired")
		}

		ok, err := isOK()
		if err != nil {
			return xerrors.Errorf("failed to check status: %v", err)
		}

		if ok {
			return nil
		}

		time.Sleep(time.Millisecond * 100)
	}
}
