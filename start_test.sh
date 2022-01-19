#!/bin/sh

# This script creates a new tmux session and starts nodes according to the
# instructions is README.md. The test session can be kill with kill_test.sh.

set -o errexit

command -v tmux >/dev/null 2>&1 || { echo >&2 "tmux is not on your PATH!"; exit 1; }

go install ./cli/memcoin

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

$node1 "LLVL=info memcoin --config /tmp/node1 start --port 2001" C-m
$node2 "LLVL=info memcoin --config /tmp/node2 start --port 2002" C-m
$node3 "LLVL=info memcoin --config /tmp/node3 start --port 2003" C-m

tmux a
