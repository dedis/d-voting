package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemcoin_Main(t *testing.T) {
	main()
}

// This test creates a chain with initially 3 nodes. It then adds node 4 and 5
// in two blocks. Node 4 does not share its certificate which means others won't
// be able to communicate, but the chain should proceed because of the
// threshold.
func TestMemcoin_Scenario_SetupAndTransactions(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "memcoin1")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	sigs := make(chan os.Signal)
	wg := sync.WaitGroup{}
	wg.Add(5)

	node1 := filepath.Join(dir, "node1")
	node2 := filepath.Join(dir, "node2")
	node3 := filepath.Join(dir, "node3")
	node4 := filepath.Join(dir, "node4")
	node5 := filepath.Join(dir, "node5")

	cfg := config{Channel: sigs, Writer: io.Discard}

	runNode(t, node1, cfg, 2111, &wg)
	runNode(t, node2, cfg, 2112, &wg)
	runNode(t, node3, cfg, 2113, &wg)
	runNode(t, node4, cfg, 2114, &wg)
	runNode(t, node5, cfg, 2115, &wg)

	defer func() {
		// Simulate a Ctrl+C
		close(sigs)
		wg.Wait()
	}()

	require.True(t, waitDaemon(t, []string{node1, node2, node3}), "daemon failed to start")

	// Share the certificates.
	shareCert(t, node2, node1, "//127.0.0.1:2111")
	shareCert(t, node3, node1, "//127.0.0.1:2111")
	shareCert(t, node5, node1, "//127.0.0.1:2111")

	// Setup the chain with nodes 1 and 2.
	args := append(append(
		append(
			[]string{os.Args[0], "--config", node1, "ordering", "setup"},
			getExport(t, node1)...,
		),
		getExport(t, node2)...),
		getExport(t, node3)...)

	err = run(args)
	require.NoError(t, err)

	// Add node 4 to the current chain. This node is not reachable from the
	// others but transactions should work as the threshold is correct.
	args = append([]string{
		os.Args[0],
		"--config", node1, "ordering", "roster", "add",
		"--wait", "60s"},
		getExport(t, node4)...,
	)

	err = run(args)
	require.NoError(t, err)

	// Add node 5 which should be participating.
	args = append([]string{
		os.Args[0],
		"--config", node1, "ordering", "roster", "add",
		"--wait", "60s"},
		getExport(t, node5)...,
	)

	err = run(args)
	require.NoError(t, err)

	// Run a few transactions.
	for i := 0; i < 5; i++ {
		err = runWithCfg(args, config{})
		require.EqualError(t, err, "command error: transaction refused: duplicate in roster: 127.0.0.1:2115")
	}

	// Test a timeout waiting for a transaction.
	args[7] = "1ns"
	err = runWithCfg(args, config{})
	require.EqualError(t, err, "command error: transaction not found after timeout")

	// Test a bad command.
	err = runWithCfg([]string{os.Args[0], "ordering", "setup"}, cfg)
	require.EqualError(t, err, `Required flag "member" not set`)
}

// This test creates a chain with two nodes, then gracefully close them. It
// finally restarts both of them to make sure the chain can proceed after the
// restart. It basically tests if the components are correctly loaded from the
// persisten storage.
func TestMemcoin_Scenario_RestartNode(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "memcoin2")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	node1 := filepath.Join(dir, "node1")
	node2 := filepath.Join(dir, "node2")

	// Setup the chain and closes the node.
	setupChain(t, []string{node1, node2}, []uint16{2210, 2211})

	sigs := make(chan os.Signal)
	wg := sync.WaitGroup{}
	wg.Add(2)

	cfg := config{Channel: sigs, Writer: io.Discard}

	// Now the node are restarted. It should correctly follow the existing chain
	// and then participate to new blocks.
	runNode(t, node1, cfg, 2210, &wg)
	runNode(t, node2, cfg, 2211, &wg)

	defer func() {
		// Simulate a Ctrl+C
		close(sigs)
		wg.Wait()
	}()

	require.True(t, waitDaemon(t, []string{node1, node2}), "daemon failed to start")

	args := append([]string{
		os.Args[0],
		"--config", node1, "ordering", "roster", "add",
		"--wait", "60s"},
		getExport(t, node1)...,
	)

	err = run(args)
	require.EqualError(t, err, "command error: transaction refused: duplicate in roster: 127.0.0.1:2210")
}

// -----------------------------------------------------------------------------
// Utility functions

const testDialTimeout = 500 * time.Millisecond

func runNode(t *testing.T, node string, cfg config, port uint16, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		err := runWithCfg(makeNodeArg(node, port), cfg)
		require.NoError(t, err)
	}()
}

func setupChain(t *testing.T, nodes []string, ports []uint16) {
	sigs := make(chan os.Signal)
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))

	cfg := config{Channel: sigs, Writer: io.Discard}

	for i, node := range nodes {
		runNode(t, node, cfg, ports[i], &wg)
	}

	defer func() {
		// Simulate a Ctrl+C
		close(sigs)
		wg.Wait()
	}()

	waitDaemon(t, nodes)

	shareCert(t, nodes[1], nodes[0], fmt.Sprintf("//127.0.0.1:%d", ports[0]))

	args := append(append(
		[]string{os.Args[0], "--config", nodes[0], "ordering", "setup"},
		getExport(t, nodes[0])...),
		getExport(t, nodes[1])...,
	)

	err := run(args)
	require.NoError(t, err)
}

func waitDaemon(t *testing.T, daemons []string) bool {
	num := 50

	for _, daemon := range daemons {
		path := filepath.Join(daemon, "daemon.sock")

		for i := 0; i < num; i++ {
			// Windows: we have to check the file as Dial on Windows creates the
			// file and prevent to listen.
			_, err := os.Stat(path)
			if !os.IsNotExist(err) {
				conn, err := net.DialTimeout("unix", path, testDialTimeout)
				if err == nil {
					conn.Close()
					break
				}
			}

			time.Sleep(100 * time.Millisecond)

			if i+1 >= num {
				return false
			}
		}
	}

	return true
}

func makeNodeArg(path string, port uint16) []string {
	return []string{
		os.Args[0], "--config", path, "start", "--listen", "tcp://127.0.0.1:" + strconv.Itoa(int(port)),
	}
}

func shareCert(t *testing.T, path string, src string, addr string) {
	args := append(
		[]string{os.Args[0], "--config", path, "minogrpc", "join", "--address", addr},
		getToken(t, src)...,
	)

	err := run(args)
	require.NoError(t, err)
}

func getToken(t *testing.T, path string) []string {
	buffer := new(bytes.Buffer)
	cfg := config{
		Writer: buffer,
	}

	args := []string{os.Args[0], "--config", path, "minogrpc", "token"}
	err := runWithCfg(args, cfg)
	require.NoError(t, err)

	return strings.Split(buffer.String(), " ")
}

func getExport(t *testing.T, path string) []string {
	buffer := bytes.NewBufferString("--member ")
	cfg := config{
		Writer: buffer,
	}

	args := []string{os.Args[0], "--config", path, "ordering", "export"}

	err := runWithCfg(args, cfg)
	require.NoError(t, err)

	return strings.Split(buffer.String(), " ")
}
