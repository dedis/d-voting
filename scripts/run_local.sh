#!/bin/bash -e

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
  if ! [[ -d dela/ ]]; then
    git clone -b fix-bbolt https://github.com/dedis/dela.git
  fi
  export GOBIN=$(pwd)/bin
  PATH="$PATH":"$GOBIN"
  if ! [[ -f $GOBIN/crypto ]]; then
    (cd dela/cli/crypto && go install)
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

function init_nodes() {
  pkill dvoting || true
  rm -rf nodes/node*
  keypair

  echo "Starting nodes"
  for n in $(seq 4); do
    NODEPORT=$((2000 + n * 2))
    PROXYPORT=$((2001 + n * 2))
    NODEDIR=./nodes/node-$n
    mkdir -p $NODEDIR
    rm -f $NODEDIR/node.log
    dvoting --config $NODEDIR start --postinstall --proxyaddr :$PROXYPORT --proxykey $PUBLIC_KEY \
      --listen tcp://0.0.0.0:$NODEPORT --public http://localhost:$NODEPORT --routing tree |
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
    dvoting --config $NODEDIR minogrpc join --address //localhost:2002 $TOKEN_ARGS
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
    dvoting --config ./nodes/node-1 pool add --key ./nodes/node-1/private.key --args go.dedis.ch/dela.ContractArg \
      --args go.dedis.ch/dela.Access --args access:grant_id \
      --args 0300000000000000000000000000000000000000000000000000000000000000 --args access:grant_contract \
      --args go.dedis.ch/dela.Evoting --args access:grant_command --args all --args access:identity --args $IDENTITY \
      --args access:command --args GRANT
  done
}

function init_db() {
  docker rm -f postgres_dvoting || true
  echo "Starting postgres database"
  docker run -d -v "$(pwd)/web/backend/src/migration.sql:/docker-entrypoint-initdb.d/init.sql" \
    -e POSTGRES_PASSWORD=$DATABASE_PASSWORD -e POSTGRES_USER=$DATABASE_USERNAME \
    --name postgres_dvoting -p 5432:5432 postgres:15 >/dev/null

  echo "Adding SCIPER to admin"
  (cd web/backend && npx ts-node src/cli.ts addAdmin --sciper $SCIPER_ADMIN | grep -v Executing)
}

function start_backend() {
  pkill -f "npm run start" || true
  keypair

  echo "Running backend"
  (cd web/backend && npm run start | ts "Backend: " &)
}

function start_frontend() {
  pkill -f "web/frontend" || true
  keypair

  echo "Running frontend"
  (cd web/frontend && npm run start | ts "Frontend: " &)
}

SCIPER_ADMIN=111443
export DATABASE_USERNAME=dvoting
export DATABASE_PASSWORD=postgres
export FRONT_END_URL="https://localhost:3000"
export DELA_NODE_URL="http://localhost:2003"
export BACKEND_HOST="localhost"
export BACKEND_PORT="6000"
export SESSION_SECRET="session secret"
export REACT_APP_NOMOCK=on

build_dela
init_nodes
init_dela
init_db
start_backend
start_frontend
