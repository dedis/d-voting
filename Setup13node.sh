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
memcoin --config /tmp/node4 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node5 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node6 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node7 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node8 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node9 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node10 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node11 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node12 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)
memcoin --config /tmp/node13 minogrpc join \
    --address 127.0.0.1:2001 $(memcoin --config /tmp/node1 minogrpc token)

echo "${GREEN}[2/7]${NC} create a chain"
memcoin --config /tmp/node1 ordering setup\
    --member $(memcoin --config /tmp/node1 ordering export)\
    --member $(memcoin --config /tmp/node2 ordering export)\
    --member $(memcoin --config /tmp/node3 ordering export)\
    --member $(memcoin --config /tmp/node4 ordering export)\
    --member $(memcoin --config /tmp/node5 ordering export)\
    --member $(memcoin --config /tmp/node6 ordering export)\
    --member $(memcoin --config /tmp/node7 ordering export)\
    --member $(memcoin --config /tmp/node8 ordering export)\
    --member $(memcoin --config /tmp/node9 ordering export)\
    --member $(memcoin --config /tmp/node10 ordering export)\
    --member $(memcoin --config /tmp/node11 ordering export)\
    --member $(memcoin --config /tmp/node12 ordering export)\
    --member $(memcoin --config /tmp/node13 ordering export)\

echo "${GREEN}[3/7]${NC} setup access rights on each node"
memcoin --config /tmp/node1 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node2 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node3 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node4 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node5 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node6 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node7 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node8 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node9 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node10 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node11 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node12 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
memcoin --config /tmp/node13 access add \
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

memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node4/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT

memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node5/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT

memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node6/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT
memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node7/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT
memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node8/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT
memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node9/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT
memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node10/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT
memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node11/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT
memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node12/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT
memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node13/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT

echo "${GREEN}[5/7]${NC} init shuffle"
memcoin --config /tmp/node1 shuffle init --signer /tmp/node1/private.key
memcoin --config /tmp/node2 shuffle init --signer /tmp/node2/private.key
memcoin --config /tmp/node3 shuffle init --signer /tmp/node3/private.key
memcoin --config /tmp/node4 shuffle init --signer /tmp/node4/private.key
memcoin --config /tmp/node5 shuffle init --signer /tmp/node5/private.key
memcoin --config /tmp/node6 shuffle init --signer /tmp/node6/private.key
memcoin --config /tmp/node7 shuffle init --signer /tmp/node7/private.key
memcoin --config /tmp/node8 shuffle init --signer /tmp/node8/private.key
memcoin --config /tmp/node9 shuffle init --signer /tmp/node9/private.key
memcoin --config /tmp/node10 shuffle init --signer /tmp/node10/private.key
memcoin --config /tmp/node11 shuffle init --signer /tmp/node11/private.key
memcoin --config /tmp/node12 shuffle init --signer /tmp/node12/private.key
memcoin --config /tmp/node13 shuffle init --signer /tmp/node13/private.key

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

memcoin --config /tmp/node4 proxy start --clientaddr 127.0.0.1:8084
memcoin --config /tmp/node4 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node4 dkg registerHandlers

memcoin --config /tmp/node5 proxy start --clientaddr 127.0.0.1:8085
memcoin --config /tmp/node5 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node5 dkg registerHandlers

memcoin --config /tmp/node6 proxy start --clientaddr 127.0.0.1:8086
memcoin --config /tmp/node6 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node6 dkg registerHandlers

memcoin --config /tmp/node7 proxy start --clientaddr 127.0.0.1:8087
memcoin --config /tmp/node7 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node7 dkg registerHandlers

memcoin --config /tmp/node8 proxy start --clientaddr 127.0.0.1:8088
memcoin --config /tmp/node8 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node8 dkg registerHandlers

memcoin --config /tmp/node9 proxy start --clientaddr 127.0.0.1:8089
memcoin --config /tmp/node9 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node9 dkg registerHandlers

memcoin --config /tmp/node10 proxy start --clientaddr 127.0.0.1:8090
memcoin --config /tmp/node10 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node10 dkg registerHandlers

memcoin --config /tmp/node11 proxy start --clientaddr 127.0.0.1:8091
memcoin --config /tmp/node11 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node11 dkg registerHandlers

memcoin --config /tmp/node12 proxy start --clientaddr 127.0.0.1:8092
memcoin --config /tmp/node12 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node12 dkg registerHandlers

memcoin --config /tmp/node13 proxy start --clientaddr 127.0.0.1:8093
memcoin --config /tmp/node13 e-voting registerHandlers --signer private.key
memcoin --config /tmp/node13 dkg registerHandlers