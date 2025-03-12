package integration

import (
	"fmt"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/d-voting/contracts/evoting/types"
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

// for Scenario
func killNode(proxyArray []string, nodeNub int, t *testing.T) []string {

	proxyArray[nodeNub-1] = proxyArray[len(proxyArray)-1]
	proxyArray[len(proxyArray)-1] = ""
	proxyArray = proxyArray[:len(proxyArray)-1]

	cmd := exec.Command("docker", "kill", fmt.Sprintf("node%v", nodeNub))
	err := cmd.Run()
	require.NoError(t, err)

	return proxyArray
}

// for Scenario
func restartNode(nodeNub int, t *testing.T) {
	cmd := exec.Command("docker", "restart", fmt.Sprintf("node%v", nodeNub))
	err := cmd.Run()
	require.NoError(t, err)

	// Replace the relative path
	cmd = exec.Command("rm", fmt.Sprintf("/Users/jean-baptistezhang/EPFL_cours/semestre_2/d-voting/nodedata/node%v/daemon.sock", nodeNub))
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("bash", "-c", fmt.Sprintf("docker exec -d node%v dvoting --config /tmp/node%v start --postinstall --promaddr :9100 --proxyaddr :9080 --proxykey adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3 --listen tcp://0.0.0.0:2001 --public //172.18.0.%v:2001", nodeNub, nodeNub, nodeNub+1))
	err = cmd.Run()
	require.NoError(t, err)
}

func getThreshold(numNodes int) int {
	return numNodes / 3
}
