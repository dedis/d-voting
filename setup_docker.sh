#!/usr/bin/env bash

# This script is creating a new chain and setting up the services needed to run
# an evoting system. It ends by starting the http server needed by the frontend
# to communicate with the blockchain. This operation is blocking.

set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color

N_NODE=$1
vals=($(seq 2 1 $N_NODE))

# for i in "${vals[@]}"
# do
#     A=$((i+200))
#     echo $A
# done

echo "${GREEN}[1/7]${NC} connect nodes"
conn_token=$(docker exec node1 memcoin --config /tmp/node1 minogrpc token)

vals=($(seq 2 1 $N_NODE))

for i in "${vals[@]}"
do
    eval docker exec node$i memcoin --config /tmp/node$i minogrpc join \
    --address 127.0.0.1:2001 $conn_token
done



echo "${GREEN}[2/7]${NC} create a chain"
vals=($(seq 1 1 $N_NODE))
ARRAY=()
for i in "${vals[@]}"
do
    ARRAY+=($(docker exec node$i memcoin --config /tmp/node$i ordering export))
done


docker exec node1 memcoin --config /tmp/node1 ordering setup\
    --member $ARRAY


echo "${GREEN}[3/7]${NC} setup access rights on each node"
access_token=$(docker exec node1 crypto bls signer read --path private.key --format BASE64_PUBKEY)

for i in "${vals[@]}"
do
    docker exec node$i memcoin --config /tmp/node$i access add \
    --identity $access_token
done



echo "${GREEN}[4/7]${NC} grant access on the chain"


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



echo "${GREEN}[5/7]${NC} init shuffle"
for i in "${vals[@]}"
do
    docker exec node$i memcoin --config /tmp/node$i shuffle init --signer /tmp/node$i/private.key
done

echo "${GREEN}[6/7]${NC} starting http proxy"
for i in "${vals[@]}"
do
    docker exec node$i memcoin --config /tmp/node$i proxy start --clientaddr 127.0.0.1:$((i+8080))
    docker exec node$i memcoin --config /tmp/node$i e-voting registerHandlers --signer private.key
    docker exec node$i memcoin --config /tmp/node$i dkg registerHandlers
done



# If an election is created with ID "deadbeef" then one must set up DKG
# on each node before the election can proceed:
# memcoin --config /tmp/node1 dkg init --electionID deadbeef
# memcoin --config /tmp/node2 dkg init --electionID deadbeef
# memcoin --config /tmp/node3 dkg init --electionID deadbeef
# memcoin --config /tmp/node1 dkg setup --electionID deadbeef

# docker exec nodetest go test -v -run ^TestScenario$ github.com/dedis/d-voting/integration