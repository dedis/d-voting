#!/usr/bin/env bash

# This script is creating a new chain and setting up the services needed to run
# an evoting system. It ends by starting the http server needed by the frontend
# to communicate with the blockchain. This operation is blocking.

set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color

N_NODE=$1


echo "${GREEN}[1/4]${NC} connect nodes"
conn_token=$(docker exec node1 memcoin --config /tmp/node1 minogrpc token)
vals=($(seq 2 1 $N_NODE))

for i in "${vals[@]}"
do
    eval docker exec node$i memcoin --config /tmp/node$i minogrpc join \
    --address //$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node1):2001 $conn_token
done


echo "${GREEN}[2/4]${NC} create a chain"
vals=($(seq 1 1 $N_NODE))
ARRAY=""
foo+=" World"
for i in "${vals[@]}"
do
    ARRAY+="--member "
    ARRAY+="$(docker exec node$i memcoin --config /tmp/node$i ordering export) "
    echo "Node$i addr is:"
    echo $(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node$i)
done


docker exec node1 memcoin --config /tmp/node1 ordering setup $ARRAY


echo "${GREEN}[3/4]${NC} setup access rights on each node"
access_token=$(docker exec node1 crypto bls signer read --path private.key --format BASE64_PUBKEY)

for i in "${vals[@]}"
do
    docker exec node$i memcoin --config /tmp/node$i access add \
    --identity $access_token
done



echo "${GREEN}[4/4]${NC} grant access on the chain"


docker exec node1 memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $access_token\
    --args access:command --args GRANT

for i in "${vals[@]}"
do
    access_token_tmp=$(docker exec node$i crypto bls signer read --path /tmp/node$i/private.key --format BASE64_PUBKEY)
    docker exec node1 memcoin --config /tmp/node1 pool add\
        --key private.key\
        --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
        --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
        --args access:grant_contract --args go.dedis.ch/dela.Evoting\
        --args access:grant_command --args all\
        --args access:identity --args $access_token_tmp\
        --args access:command --args GRANT
done


