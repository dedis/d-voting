package integration

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	var err error
	numNodes := defaultNodes
	t.Log("Start")

	n, ok := os.LookupEnv("NNODES")
	if ok {
		numNodes, err = strconv.Atoi(n)
		require.NoError(t, err)
	}
	t.Run("Basic configuration", getLoadTest(numNodes, 2, 60, 1))
}

func getLoadTest(numNodes, numVotesPerSec, numSec, numForm int) func(*testing.T) {
	return func(t *testing.T) {

		proxyList := make([]string, numNodes)

		for i := 0; i < numNodes; i++ {
			proxyList[i] = fmt.Sprintf("http://localhost:%d", 9080+i)
			t.Log(proxyList[i])
		}

		var wg sync.WaitGroup

		for i := 0; i < numForm; i++ {
			t.Log("Starting worker", i)
			wg.Add(1)

			go startFormProcess(&wg, numNodes, numVotesPerSec*numSec, proxyList, t, numForm, castVotesLoad(numVotesPerSec, numSec))
			time.Sleep(2 * time.Second)

		}

		t.Log("Waiting for workers to finish")
		wg.Wait()

	}
}
