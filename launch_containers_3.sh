#!/bin/sh

# This script creates a new tmux session and starts nodes according to the
# instructions is README.md. The test session can be kill with kill_test.sh.

set -o errexit

command -v tmux >/dev/null 2>&1 || { echo >&2 "tmux is not on your PATH!"; exit 1; }


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

# Clean containers and tmp dir
if [[ $(docker ps -a -q) ]]; then
    docker rm -f $(docker ps -a -q)
fi

rm -rf ./nodedata    
mkdir nodedata

#Create docker network (only run once)
# docker network create --driver bridge evoting-net


#Start docker images and bind volume
docker run -d -it --name node1 --network evoting-net -v "$(pwd)"/nodedata:/tmp node
docker run -d -it --name node2 --network evoting-net -v "$(pwd)"/nodedata:/tmp node
docker run -d -it --name node3 --network evoting-net -v "$(pwd)"/nodedata:/tmp node

#Get docker images IP addr
addr1=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node1)
addr2=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node2)
addr3=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node3)


pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3

$node1 "eval docker exec node1 memcoin --config /tmp/node1 start --postinstall \
  --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //$addr1:2001 &
" C-m
$node2 "eval docker exec node2 memcoin --config /tmp/node2 start --postinstall \
  --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //$addr2:2001 &
" C-m
$node3 "eval docker exec node3 memcoin --config /tmp/node3 start --postinstall \
  --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //$addr3:2001 &
" C-m

tmux a
