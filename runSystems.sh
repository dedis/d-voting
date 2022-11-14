#!/bin/bash

# This script is creating n dela voting nodes needed to run
# an evoting system. User can pass number of nodes, window attach mode useful for autotest,
# and docker usage.

set -e

# by default run on local
DOCKER=false
ATTACH=true
# by default run and setup everything
RUN=true
SETUP=true
FRONTEND=true
BACKEND=true

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
  -h | --help)
    echo "This script is creating n dela voting nodes"
    echo "Options:"
    echo "-h  |  --help     program help (this file)"
    echo "-n  |  --node     number of d-voting nodes"
    echo "-a  |  --attach   attach tmux window to current shell true/false, by default true"
    echo "-d  |  --docker   launch nodes on docker containers true/false, by default false"
    echo "-r  |  --run      run the nodes true/false, by default true"
    echo "-s  |  --setup    setup the nodes true/false, by default true"
    echo "-f  |  --frontend setup the frontend true/false, by default true"
    echo "-b  |  --backend  setup the backend true/false, by default true"
    exit 0
    ;;
  -r | --run)
    RUN="$2"
    shift # past argument
    shift # past value
    ;;
  -s | --setup)
    SETUP="$2"
    shift # past argument
    shift # past value
    ;;
  -f | --frontend)
    FRONTEND="$2"
    shift # past argument
    shift # past value
    ;;
  -b | --backend)
    BACKEND="$2"
    shift # past argument
    shift # past value
    ;;
  -n | --node)
    N_NODE="$2"
    shift # past argument
    shift # past value
    ;;
  -a | --attach)
    ATTACH="$2"
    shift # past argument
    shift # past value
    ;;
  -d | --docker)
    DOCKER="$2"
    shift # past argument
    shift # past value
    ;;
  -* | --*)
    echo "Unknown option $1"
    exit 1
    ;;
  *)
    POSITIONAL_ARGS+=("$1") # save positional arg
    shift                   # past argument
    ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

set -o errexit

command -v tmux >/dev/null 2>&1 || {
  echo >&2 "tmux is not on your PATH!"
  exit 1
}

pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3

# Launch session
s="d-voting-test"
from=0
if [ "$RUN" == true ]; then

  if [ -z "$N_NODE" ]; then
    echo "Please specify the number of nodes"
    exit 1
  fi

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
  while [ $from -le $to ]; do

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
      docker run -d -it --env LLVL=info --name node$from --network evoting-net -v "$(pwd)"/nodedata:/tmp --publish $((9079 + $from)):9080 node
      tmux send-keys -t $s:$window "docker exec node$from memcoin --config /tmp/node$from start --postinstall \
    --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node$from):2001 | tee ./log/log/node$from.log" C-m
    fi

    ((from++))
  done

fi

if [ "$BACKEND" == true ]; then
  if tmux has-session -t $s 2>/dev/null; then
    # window for the backend
    tmux new-window -t $s -n "backend"
    tmux send-keys -t $s:{end} "cd web/backend && npm start" C-m
  else
    #run it in the current shell
    cd web/backend && npm start
  fi
fi

# window for the frontend
if [ "$FRONTEND" == true ]; then
  if tmux has-session -t $s 2>/dev/null; then
    tmux new-window -t $s -n "frontend"
    tmux send-keys -t $s:{end} "cd web/frontend && REACT_APP_PROXY=http://localhost:9081 REACT_APP_NOMOCK=on npm start" C-m
  else
    #run it in the current shell
    cd web/frontend && REACT_APP_PROXY=http://localhost:9081 REACT_APP_NOMOCK=on npm start
  fi
fi

((from++))

# Setup
if [ "$SETUP" == true ]; then
  if [ -z "$N_NODE" ]; then
    echo "Please specify the number of nodes"
    exit 1
  fi
  if [ "$RUN" == true ]; then
    sleep 8
  fi
  if tmux has-session -t $s 2>/dev/null; then
    # window for the setup
    GREEN='\033[0;32m'
    NC='\033[0m' # No Color

    if [ "$DOCKER" == false ]; then
      echo "${GREEN}[1/4]${NC} connect nodes"

      from=2
      to=$N_NODE
      while [ $from -le $to ]; do
        ./memcoin --config /tmp/node$from minogrpc join \
          --address //localhost:2001 $(./memcoin --config /tmp/node1 minogrpc token)

        ((from++))
      done

      echo "${GREEN}[2/4]${NC} create a chain"

      ARRAY=""
      from=1
      to=$N_NODE
      while [ $from -le $to ]; do
        ARRAY+="--member "
        ARRAY+="$(./memcoin --config /tmp/node$from ordering export) "

        ((from++))
      done

      ./memcoin --config /tmp/node1 ordering setup $ARRAY

      echo "${GREEN}[3/4]${NC} setup access rights on each node"

      from=1

      while [ $from -le $to ]; do
        ./memcoin --config /tmp/node$from access add \
          --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)

        ((from++))
      done

      echo "${GREEN}[4/4]${NC} grant access on the chain"

      ./memcoin --config /tmp/node1 pool add --key private.key --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000 --args access:grant_contract --args go.dedis.ch/dela.Evoting --args access:grant_command --args all --args access:identity --args $(crypto bls signer read --path private.key --format BASE64_PUBKEY) \
        --args access:command --args GRANT

      from=1

      while [ $from -le $to ]; do

        ./memcoin --config /tmp/node1 pool add --key private.key --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000 --args access:grant_contract --args go.dedis.ch/dela.Evoting --args access:grant_command --args all --args access:identity --args $(crypto bls signer read --path /tmp/node$from/private.key --format BASE64_PUBKEY) \
          --args access:command --args GRANT

        ((from++))
      done
    else
      echo "${GREEN}[1/4]${NC} connect nodes"
      conn_token=$(docker exec node1 memcoin --config /tmp/node1 minogrpc token)
      vals=($(seq 2 1 $N_NODE))

      for i in "${vals[@]}"; do
        docker exec node$i memcoin --config /tmp/node$i minogrpc join \
          --address //$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node1):2001 $conn_token

      done

      echo "${GREEN}[2/4]${NC} create a chain"
      vals=($(seq 1 1 $N_NODE))
      ARRAY=""
      for i in "${vals[@]}"; do
        ARRAY+="--member "
        ARRAY+="$(docker exec node$i memcoin --config /tmp/node$i ordering export) "
        echo "Node$i addr is:"
        echo $(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node$i)
      done

      docker exec node1 memcoin --config /tmp/node1 ordering setup $ARRAY

      echo "${GREEN}[3/4]${NC} setup access rights on each node"
      access_token=$(docker exec node1 crypto bls signer read --path private.key --format BASE64_PUBKEY)

      for i in "${vals[@]}"; do
        docker exec node$i memcoin --config /tmp/node$i access add \
          --identity $access_token
        sleep 1
      done

      echo "${GREEN}[4/4]${NC} grant access on the chain"

      docker exec node1 memcoin --config /tmp/node1 pool add --key private.key --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000 --args access:grant_contract --args go.dedis.ch/dela.Evoting --args access:grant_command --args all --args access:identity --args $access_token --args access:command --args GRANT

      sleep 1

      for i in "${vals[@]}"; do
        access_token_tmp=$(docker exec node$i crypto bls signer read --path /tmp/node$i/private.key --format BASE64_PUBKEY)
        docker exec node1 memcoin --config /tmp/node1 pool add --key private.key --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000 --args access:grant_contract --args go.dedis.ch/dela.Evoting --args access:grant_command --args all --args access:identity --args $access_token_tmp --args access:command --args GRANT
        sleep 1
      done
    fi
  fi
fi

if [ "$ATTACH" == true ]; then
  tmux a
fi
