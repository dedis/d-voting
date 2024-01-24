#!/bin/bash -e

# The script must be called from the root of the github tree, else it returns an error.
# This script currently only works on Linux due to differences in network management on Windows/macOS.

if [[ $(git rev-parse --show-toplevel) != $(pwd) ]]; then
  echo "ERROR: This script must be started from the root of the git repo";
  exit 1;
fi


source ./.env;
export COMPOSE_FILE=${COMPOSE_FILE:-./docker-compose/docker-compose.yml};


function setup() {
  docker compose build;
  docker compose up -d;
}

function teardown() {
  rm -f cookies.txt;
  docker compose down -v;
  docker image rm ghcr.io/c4dt/d-voting-frontend:latest ghcr.io/c4dt/d-voting-backend:latest ghcr.io/c4dt/d-voting-dela:latest;
}

function init_dela() {
  LEADER=dela-worker-0;
  echo "$LEADER is the initial leader node";

  echo "add nodes to the chain";
  MEMBERS=""
  for node in $(seq 0 3); do
    MEMBERS="$MEMBERS --member $(docker compose exec dela-worker-$node /bin/bash -c 'LLVL=error dvoting --config /data/node ordering export')";
  done
  docker compose exec "$LEADER" dvoting --config /data/node ordering setup $MEMBERS;

  echo "authorize signers to handle access contract on each node";
  for signer in $(seq 0 3); do
    IDENTITY=$(docker compose exec "dela-worker-$signer" crypto bls signer read --path /data/node/private.key --format BASE64_PUBKEY);
    for node in $(seq 0 3); do
      docker compose exec "dela-worker-$node" dvoting --config /data/node access add --identity "$IDENTITY";
    done
  done

  echo "update the access contract";
  for node in $(seq 0 3); do
    IDENTITY=$(docker compose exec dela-worker-"$node" crypto bls signer read --path /data/node/private.key --format BASE64_PUBKEY);
    docker compose exec "$LEADER" dvoting --config /data/node pool add\
        --key /data/node/private.key\
        --args go.dedis.ch/dela.ContractArg\
        --args go.dedis.ch/dela.Access\
        --args access:grant_id\
        --args 45564f54\
        --args access:grant_contract\
        --args go.dedis.ch/dela.Evoting \
        --args access:grant_command\
        --args all\
        --args access:identity\
        --args $IDENTITY\
        --args access:command\
        --args GRANT
  done
}


function local_admin() {
  echo "adding local user $REACT_APP_SCIPER_ADMIN to admins";
  docker compose exec backend npx cli addAdmin --sciper "$REACT_APP_SCIPER_ADMIN";
  docker compose restart backend;
}


function local_login() {
  if ! [ -f cookies.txt ]; then
   echo "getting dummy login cookie";
   curl -k "$FRONT_END_URL/api/get_dev_login/$REACT_APP_SCIPER_ADMIN" -c cookies.txt -o /dev/null -s;
  fi
}

function add_proxies() {

  echo "adding proxies";

  for node in $(seq 0 3); do
    echo "adding proxy for node dela-worker-$node";
    curl -sk "$FRONT_END_URL/api/proxies/" -X POST -H 'Content-Type: application/json' -b cookies.txt --data "{\"NodeAddr\":\"grpc://dela-worker-$node:$NODEPORT\",\"Proxy\":\"http://172.19.44.$((254 - node)):$PROXYPORT\"}";
  done
}

case "$1" in

setup)
  setup;
  ;;

init_dela)
  init_dela;
  ;;

teardown)
  teardown;
  exit
  ;;

local_admin)
  local_admin;
  ;;

add_proxies)
  local_login;
  add_proxies;
  ;;

*)
  setup;
  sleep 16;     # give DELA nodes time to start up
  init_dela;
  local_admin;
  sleep 8;      # give backend time to restart
  local_login;
  add_proxies;
  ;;
esac
