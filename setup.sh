#!/usr/bin/env bash

# This script is creating a new chain and setting up the services needed to run
# an evoting system. It ends by starting the http server needed by the frontend
# to communicate with the blockchain. This operation is blocking.

set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "${GREEN}[1/7]${NC} connect nodes"
memcoin --config /tmp/node2 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node3 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)

echo "${GREEN}[2/7]${NC} create a chain"
memcoin --config /tmp/node1 ordering setup\
    --member $(memcoin --config /tmp/node1 ordering export)\
    --member $(memcoin --config /tmp/node2 ordering export)\
    --member $(memcoin --config /tmp/node3 ordering export)

echo "${GREEN}[3/7]${NC} setup access rights on each node"
memcoin --config /tmp/node1 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node2 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node3 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)

echo "${GREEN}[4/7]${NC} grant access on the chain"
memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT

memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node1/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT

memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node2/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT    

memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node3/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT

echo "${GREEN}[5/7]${NC} init shuffle"
memcoin --config /tmp/node1 shuffle init --signer /tmp/node1/private.key
memcoin --config /tmp/node2 shuffle init --signer /tmp/node2/private.key
memcoin --config /tmp/node3 shuffle init --signer /tmp/node3/private.key

echo "${GREEN}[6/7]${NC} starting http proxy"
memcoin --config /tmp/node1 proxy start --clientaddr 127.0.0.1:8081
memcoin --config /tmp/node1 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node1 dkg registerHandlers

memcoin --config /tmp/node2 proxy start --clientaddr 127.0.0.1:8082
memcoin --config /tmp/node2 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node2 dkg registerHandlers

memcoin --config /tmp/node3 proxy start --clientaddr 127.0.0.1:8083
memcoin --config /tmp/node3 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node3 dkg registerHandlers

# If an election is created with ID "deadbeef" then one must set up DKG
# on each node before the election can proceed:
# memcoin --config /tmp/node1 dkg init --electionID deadbeef
# memcoin --config /tmp/node2 dkg init --electionID deadbeef
# memcoin --config /tmp/node3 dkg init --electionID deadbeef
# memcoin --config /tmp/node1 dkg setup --electionID deadbeef
