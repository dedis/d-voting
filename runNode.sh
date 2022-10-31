#!/bin/bash

# This script is creating n dela voting nodes needed to run
# an evoting system. User can pass number of nodes, window attach mode useful for autotest,
# and docker usage.

set -e

# by default run on local
DOCKER=false
ATTACH=true

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      echo      "This script is creating n dela voting nodes"
      echo      "Options:"
      echo      "-h  |  --help     program help (this file)"
      echo      "-n  |  --node     number of d-voting nodes"
      echo      "-a  |  --attach   attach tmux window to current shell true/false, by default true"
      echo      "-d  |  --docker   launch nodes on docker containers true/false, by default false"
      exit 0
      ;;
    -n|--node)
      N_NODE="$2"
      shift # past argument
      shift # past value
      ;;
    -a|--attach)
      ATTACH="$2"
      shift # past argument
      shift # past value
      ;;
    -d|--docker)
      DOCKER="$2"
      shift # past argument
      shift # past value
      ;;
    -*|--*)
      echo "Unknown option $1"
      exit 1
      ;;
    *)
      POSITIONAL_ARGS+=("$1") # save positional arg
      shift # past argument
      ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

set -o errexit

command -v tmux >/dev/null 2>&1 || { echo >&2 "tmux is not on your PATH!"; exit 1; }


pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3


# Launch session
s="d-voting-test"

# check if session already exists, if so run the kill_test.sh script
if tmux has-session -t $s 2>/dev/null; then
  echo "Session $s already exists, killing it"
  ./kill_test.sh
fi



tmux new-session -d -s $s

# Checks that we can afford to have at least one Byzantine node and keep the
# system working, which is not possible with less than 4 nodes.
if [ $N_NODE -le 3 ]; then
  echo "Warning: the number of nodes is less or equal than 3, it will not be resiliant if one node is down"
fi


# Clean logs
rm -rf ./log/log
mkdir -p ./log/log

crypto bls signer new --save private.key --force

if [ "$DOCKER" == false ]; then
    go build -o memcoin ./cli/memcoin
else
    # Clean created containers and tmp dir
    if [[ $(docker ps -a -q --filter ancestor=node) ]]; then
        docker rm -f $(docker ps -a -q --filter ancestor=node)
    fi

    rm -rf ./nodedata    
    mkdir nodedata

    # Create docker network (only run once)
    docker network create --driver bridge evoting-net || true

fi

from=1
to=$N_NODE
while [ $from -le $to ]
do

echo $from
tmux new-window -t $s
window=$from

if [ "$DOCKER" == false ]; then
    tmux send-keys -t $s:$window "PROXY_LOG=info LLVL=info ./memcoin \
      --config /tmp/node$from \
      start \
      --postinstall \
      --promaddr :$((9099 + $from)) \
      --proxyaddr :$((9079 + $from)) \
      --proxykey $pk \
      --listen tcp://0.0.0.0:$((2000 + $from)) \
      --routing tree \
      --public //localhost:$((2000 + $from))| tee ./log/log/node$from.log" C-m
else
    docker run -d -it --env LLVL=info --name node$from --network evoting-net -v "$(pwd)"/nodedata:/tmp  --publish $(( 9079+$from )):9080 node
    tmux send-keys -t $s:$window "docker exec node$from memcoin --config /tmp/node$from start --postinstall \
  --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node$from):2001 | tee ./log/log/node$from.log" C-m
fi

((from++))
done

tmux new-window -t $s


if [ "$ATTACH" == true ]; then
    tmux a
fi
