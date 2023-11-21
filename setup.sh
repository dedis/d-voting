#!/usr/bin/env bash

# This script is creating a new chain and setting up the services needed to run
# an evoting system. It ends by starting the http server needed by the frontend
# to communicate with the blockchain. This operation is blocking. It is expected
# that the "dvoting" binary is at the root. You can build it with:
#   go build ./cli/dvoting

set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "${GREEN}[1/7]${NC} connect nodes"
./dvoting --config /tmp/node2 minogrpc join \
    --address //localhost:2001 $(./dvoting --config /tmp/node1 minogrpc token)
./dvoting --config /tmp/node3 minogrpc join \
    --address //localhost:2001 $(./dvoting --config /tmp/node1 minogrpc token)

echo "${GREEN}[2/7]${NC} create a chain"
./dvoting --config /tmp/node1 ordering setup\
    --member $(./dvoting --config /tmp/node1 ordering export)\
    --member $(./dvoting --config /tmp/node2 ordering export)\
    --member $(./dvoting --config /tmp/node3 ordering export)

echo "${GREEN}[3/7]${NC} setup access rights on each node"
./dvoting --config /tmp/node1 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
./dvoting --config /tmp/node2 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)
./dvoting --config /tmp/node3 access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)

echo "${GREEN}[4/7]${NC} grant access on the chain"
./dvoting --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 45564f54\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT

./dvoting --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 45564f54\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node1/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT

./dvoting --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 45564f54\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node2/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT    

./dvoting --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 45564f54\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node3/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT

# The following is not needed anymore thanks to the "postinstall" functionality.
# See #65.

# echo "${GREEN}[5/7]${NC} init shuffle"
# ./dvoting --config /tmp/node1 shuffle init --signer /tmp/node1/private.key
# ./dvoting --config /tmp/node2 shuffle init --signer /tmp/node2/private.key
# ./dvoting --config /tmp/node3 shuffle init --signer /tmp/node3/private.key

# echo "${GREEN}[6/7]${NC} starting http proxy"
# ./dvoting --config /tmp/node1 proxy start --clientaddr 127.0.0.1:8081
# ./dvoting --config /tmp/node1 e-voting registerHandlers --signer private.key
# ./dvoting --config /tmp/node1 dkg registerHandlers

# ./dvoting --config /tmp/node2 proxy start --clientaddr 127.0.0.1:8082
# ./dvoting --config /tmp/node2 e-voting registerHandlers --signer private.key
# ./dvoting --config /tmp/node2 dkg registerHandlers

# ./dvoting --config /tmp/node3 proxy start --clientaddr 127.0.0.1:8083
# ./dvoting --config /tmp/node3 e-voting registerHandlers --signer private.key
# ./dvoting --config /tmp/node3 dkg registerHandlers

# If a form is created with ID "deadbeef" then one must set up DKG
# on each node before the form can proceed:
# ./dvoting --config /tmp/node1 dkg init --formID deadbeef
# ./dvoting --config /tmp/node2 dkg init --formID deadbeef
# ./dvoting --config /tmp/node3 dkg init --formID deadbeef
# ./dvoting --config /tmp/node1 dkg setup --formID deadbeef
