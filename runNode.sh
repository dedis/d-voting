#!/bin/bash


set -o errexit

command -v tmux >/dev/null 2>&1 || { echo >&2 "tmux is not on your PATH!"; exit 1; }


pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3


# Launch session
s="d-voting-test"

tmux list-sessions | rg "^$s:" >/dev/null 2>&1 && { echo >&2 "A session with the name $s already exists; kill it and try again"; exit 1; }

tmux new-session -d -s $s


from=1
to=$1
while [ $from -le $to ]
do

echo $from
tmux new-window -t $s
window=$from
tmux send-keys -t $s:$window "LLVL=info ./memcoin --config /tmp/node$from start --postinstall --promaddr :$((9099 + $from)) --proxyaddr :$((9079 + $from)) --proxykey $pk --listen tcp://0.0.0.0:$((2000 + $from)) --public //localhost:$((2000 + $from))" C-m
((from++))

done

tmux new-window -t $s

tmux a
