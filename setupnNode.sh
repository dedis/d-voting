#!/usr/bin/env bash

# This script is creating a new chain and setting up the services needed to run
# an evoting system. It ends by starting the http server needed by the frontend
# to communicate with the blockchain. This operation is blocking. It is expected
# that the "memcoin" binary is at the root. You can build it with:
#   go build ./cli/memcoin

set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color


echo "${GREEN}[1/4]${NC} connect nodes"

from=2
to=$1
while [ $from -le $to ]
do
./memcoin --config /tmp/node$from minogrpc join \
    --address //localhost:2001 $(./memcoin --config /tmp/node1 minogrpc token)

((from++))
done

echo "${GREEN}[2/4]${NC} create a chain"

ARRAY=""
from=1
to=$1
while [ $from -le $to ]
do
 ARRAY+="--member "
 ARRAY+="$(./memcoin --config /tmp/node$from ordering export) "

((from++))
done

./memcoin --config /tmp/node1 ordering setup $ARRAY


echo "${GREEN}[3/4]${NC} setup access rights on each node"

from=1

while [ $from -le $to ]
do
./memcoin --config /tmp/node$from access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)

((from++))
done


echo "${GREEN}[4/4]${NC} grant access on the chain"

./memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT


from=1

while [ $from -le $to ]
do

./memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node$from/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT


((from++))
done
