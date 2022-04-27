#!/bin/sh

# This script creates a new tmux session and starts nodes according to the
# instructions is README.md. The test session can be kill with kill_test.sh.

set -o errexit

command -v tmux >/dev/null 2>&1 || { echo >&2 "tmux is not on your PATH!"; exit 1; }

make build

# Launch session
s="d-voting-test"

tmux list-sessions | rg "^$s:" >/dev/null 2>&1 && { echo >&2 "A session with the name $s already exists; kill it and try again"; exit 1; }

tmux new -s $s -d

tmux split-window -t $s -h
tmux split-window -t $s:0.%0
tmux split-window -t $s:0.%1

# session s, window 0, panes 0 to 2
master="tmux send-keys -t $s:0.%0"
node1="tmux send-keys -t $s:0.%1"
node2="tmux send-keys -t $s:0.%2"
node3="tmux send-keys -t $s:0.%3"

pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3

$node1 "LLVL=info ./memcoin --config /tmp/node1 start --postinstall --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //localhost:2001" C-m
$node2 "LLVL=info ./memcoin --config /tmp/node2 start --postinstall --promaddr :9101 --proxyaddr :9081 --proxykey $pk --listen tcp://0.0.0.0:2002 --public //localhost:2002" C-m
$node3 "LLVL=info ./memcoin --config /tmp/node3 start --postinstall --promaddr :9102 --proxyaddr :9082 --proxykey $pk --listen tcp://0.0.0.0:2003 --public //localhost:2003" C-m

tmux a
