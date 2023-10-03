#!/bin/bash -e
# This puts all the different steps of initializing a Dela d-voting network into one shell script.
# This can be used for development by calling the script and then testing the result locally.
# The script must be called from the root of the github tree, else it returns an error.
# If the script is called with `./scripts/run_local.sh clean`, it stops all services.
# For development, the calls to the different parts can be adjusted, e.g., comment all but
# `start_backend` to only restart the backend.

if [[ $(git rev-parse --show-toplevel) != $(pwd) ]]; then
  echo "ERROR: This script must be started from the root of the git repo"
  exit 1
fi

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
. "$SCRIPT_DIR/local_vars.sh"

asdf_shell() {
  if ! asdf list "$1" | grep -wq "$2"; then
    asdf install "$1" "$2"
  fi
  asdf local "$1" "$2"
}
asdf_shell nodejs 16.20.2
asdf_shell golang 1.21.0
mkdir -p nodes

function build_dela() {
  echo "Building dela-node"
  export GOBIN=$(pwd)/bin
  PATH="$PATH":"$GOBIN"
  if ! [[ -f $GOBIN/crypto ]]; then
    go install github.com/c4dt/dela/cli/crypto
  fi
  if ! [[ -f $GOBIN/dvoting ]]; then
    go install ./cli/dvoting
  fi

  echo "Installing node directories"
  for d in backend frontend; do
    DIR=web/$d
    if ! [[ -d $DIR/node_modules ]]; then
      (cd $DIR && npm ci)
    fi
  done
}

function keypair() {
  if ! [[ "$PUBLIC_KEY" ]]; then
    if ! [[ -f nodes/keypair ]]; then
      echo "Getting keypair"
      (cd web/backend && npm run keygen) | tail -n 2 >nodes/keypair
    fi
    . nodes/keypair
    export PUBLIC_KEY PRIVATE_KEY
  fi
}

function kill_nodes() {
  pkill dvoting || true

  echo "Waiting for nodes to stop"
  for n in $(seq 4); do
    NODEPORT=$((2000 + n * 2 - 2))
    while lsof -ln | grep -q :$NODEPORT; do sleep .1; done
  done

  rm -rf nodes/node*
}

function init_nodes() {
  kill_nodes
  keypair

  echo "Starting nodes"
  for n in $(seq 4); do
    NODEPORT=$((2000 + n * 2 - 2))
    PROXYPORT=$((2001 + n * 2 - 2))
    NODEDIR=./nodes/node-$n
    mkdir -p $NODEDIR
    rm -f $NODEDIR/node.log
    dvoting --config $NODEDIR start --postinstall --proxyaddr :$PROXYPORT --proxykey $PUBLIC_KEY \
      --listen tcp://0.0.0.0:$NODEPORT --public grpc://localhost:$NODEPORT --routing tree --noTLS |
      ts "Node-$n: " | tee $NODEDIR/node.log &
  done

  echo "Waiting for nodes to start up"
  for n in $(seq 4); do
    NODEDIR=./nodes/node-$n
    while ! [[ -S $NODEDIR/daemon.sock && -f $NODEDIR/node.log && $(cat $NODEDIR/node.log | wc -l) -ge 2 ]]; do
      sleep .2
    done
  done
}

function init_dela() {
  echo "Initializing dela"
  echo "  Share the certificate"
  for n in $(seq 2 4); do
    TOKEN_ARGS=$(dvoting --config ./nodes/node-1 minogrpc token)
    NODEDIR=./nodes/node-$n
    dvoting --config $NODEDIR minogrpc join --address grpc://localhost:2000 $TOKEN_ARGS
  done

  echo "  Create a new chain with the nodes"
  for n in $(seq 4); do
    NODEDIR=./nodes/node-$n
    # add node to the chain
    MEMBERS="$MEMBERS --member $(dvoting --config $NODEDIR ordering export)"
  done
  dvoting --config ./nodes/node-1 ordering setup $MEMBERS

  echo "  Authorize the signer to handle the access contract on each node"
  for s in $(seq 4); do
    NODEDIR=./nodes/node-$s
    IDENTITY=$(crypto bls signer read --path $NODEDIR/private.key --format BASE64_PUBKEY)
    for n in $(seq 4); do
      NODEDIR=./nodes/node-$n
      dvoting --config $NODEDIR access add --identity "$IDENTITY"
    done
  done

  echo "  Update the access contract"
  for n in $(seq 4); do
    NODEDIR=./nodes/node-$n
    IDENTITY=$(crypto bls signer read --path $NODEDIR/private.key --format BASE64_PUBKEY)
    dvoting --config ./nodes/node-1 pool add --key ./nodes/node-1/private.key --args github.com/c4dt/dela.ContractArg \
      --args github.com/c4dt/dela.Access --args access:grant_id \
      --args 0300000000000000000000000000000000000000000000000000000000000000 --args access:grant_contract \
      --args github.com/c4dt/dela.Evoting --args access:grant_command --args all --args access:identity --args $IDENTITY \
      --args access:command --args GRANT
  done

  rm -f cookies.txt
}

function kill_db() {
  docker rm -f postgres_dvoting || true
  rm -rf nodes/llmdb*
}

function init_db() {
  kill_db

  echo "Starting postgres database"
  docker run -d -v "$(pwd)/web/backend/src/migration.sql:/docker-entrypoint-initdb.d/init.sql" \
    -e POSTGRES_PASSWORD=$DATABASE_PASSWORD -e POSTGRES_USER=$DATABASE_USERNAME \
    --name postgres_dvoting -p 5432:5432 postgres:15 >/dev/null

  echo "Adding $SCIPER_ADMIN to admin"
  (cd web/backend && npx ts-node src/cli.ts addAdmin --sciper $SCIPER_ADMIN | grep -v Executing)
}

function kill_backend() {
  pkill -f "web/backend" || true
  rm -f cookies.txt
}

function start_backend() {
  kill_backend
  keypair

  echo "Running backend"
  (cd web/backend && npm run start-dev | ts "Backend: " &)

  while ! lsof -ln | grep -q :6000; do sleep .1; done
}

function kill_frontend() {
  pkill -f "web/frontend" || true
}

function start_frontend() {
  kill_frontend
  keypair

  echo "Running frontend"
  (cd web/frontend && npm run start-dev | ts "Frontend: " &)
}

case "$1" in
clean)
  kill_frontend
  kill_nodes
  kill_backend
  kill_db
  rm -f bin/*
  exit
  ;;

build)
  build_dela
  ;;

init_nodes)
  init_nodes
  ;;

init_dela)
  init_dela
  ;;

init_db)
  init_db
  ;;

backend)
  start_backend
  ;;

frontend)
  start_frontend
  ;;

*)
  build_dela
  init_nodes
  init_dela
  init_db
  start_backend
  start_frontend
  ;;
esac
