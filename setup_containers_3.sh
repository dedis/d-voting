#!/usr/bin/env bash

# This script is creating docker containers, creating a new chain and setting up the services needed to run
# an evoting system. It ends by starting the http server needed by the frontend
# to communicate with the blockchain. This operation is blocking. It is expected
# that the "memcoin" binary is at the root. You can build it with:
#   go build ./cli/memcoin

set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color

#Get docker images IP addr
addr1=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node1)
addr2=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node2)
addr3=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node3)

echo "${GREEN}[1/4]${NC} connect nodes"
conn_token=$(docker exec node1 memcoin --config /tmp/node1 minogrpc token)
eval docker exec node1 memcoin --config /tmp/node2 minogrpc join \
    --address //$addr1:2001 $conn_token
eval docker exec node1 memcoin --config /tmp/node3 minogrpc join \
    --address //$addr1:2001 $conn_token

echo "${GREEN}[2/4]${NC} create a chain"
join1=$(docker exec node1 memcoin --config /tmp/node1 ordering export)
join2=$(docker exec node2 memcoin --config /tmp/node2 ordering export)
join3=$(docker exec node3 memcoin --config /tmp/node3 ordering export)

docker exec node1 memcoin --config /tmp/node1 ordering setup\
    --member $join1\
    --member $join2\
    --member $join3

echo "${GREEN}[3/4]${NC} setup access rights on each node"
access_token=$(docker exec node1 crypto bls signer read --path private.key --format BASE64_PUBKEY)
docker exec node1 memcoin --config /tmp/node1 access add \
    --identity $access_token
docker exec node2 memcoin --config /tmp/node2 access add \
    --identity $access_token
docker exec node3 memcoin --config /tmp/node3 access add \
    --identity $access_token


echo "${GREEN}[4/4]${NC} grant access on the chain"
access_token1=$(docker exec node1 crypto bls signer read --path /tmp/node1/private.key --format BASE64_PUBKEY)
access_token2=$(docker exec node2 crypto bls signer read --path /tmp/node2/private.key --format BASE64_PUBKEY)
access_token3=$(docker exec node3 crypto bls signer read --path /tmp/node3/private.key --format BASE64_PUBKEY)

docker exec node1 memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $access_token\
    --args access:command --args GRANT

docker exec node1 memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $access_token1\
    --args access:command --args GRANT

docker exec node1 memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $access_token2\
    --args access:command --args GRANT    

docker exec node1 memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $access_token3\
    --args access:command --args GRANT