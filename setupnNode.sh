#!/usr/bin/env bash

# This script is creating a new chain and setting up the services needed to run
# an evoting system. It ends by starting the http server needed by the frontend
# to communicate with the blockchain. This operation is blocking. It is expected
# that the "memcoin" binary is at the root. You can build it with:
#   go build ./cli/memcoin

set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color


echo "${GREEN}[1/7]${NC} connect nodes"

fromb=2
to=$1
while [ $fromb -le $to ]
do
./memcoin --config /tmp/node$fromb minogrpc join \
    --address //localhost:2001 $(./memcoin --config /tmp/node1 minogrpc token)

((fromb++))
done

echo "${GREEN}[2/7]${NC} create a chain"
if [ $1 -eq 3 ]
then 

./memcoin --config /tmp/node1 ordering setup\
 --member $(./memcoin --config /tmp/node1 ordering export)\
 --member $(./memcoin --config /tmp/node2 ordering export)\
  --member $(./memcoin --config /tmp/node3 ordering export)

elif [ $1 -eq 4 ]
then

./memcoin --config /tmp/node1 ordering setup\
    --member $(./memcoin --config /tmp/node1 ordering export)\
    --member $(./memcoin --config /tmp/node2 ordering export)\
    --member $(./memcoin --config /tmp/node3 ordering export)\
     --member $(./memcoin --config /tmp/node4 ordering export)

elif [ $1 -eq 5 ]
then

./memcoin --config /tmp/node1 ordering setup\
    --member $(./memcoin --config /tmp/node1 ordering export)\
    --member $(./memcoin --config /tmp/node2 ordering export)\
    --member $(./memcoin --config /tmp/node3 ordering export)\
     --member $(./memcoin --config /tmp/node4 ordering export)\
     --member $(./memcoin --config /tmp/node5 ordering export)

elif [ $1 -eq 6 ]
then

./memcoin --config /tmp/node1 ordering setup\
    --member $(./memcoin --config /tmp/node1 ordering export)\
    --member $(./memcoin --config /tmp/node2 ordering export)\
    --member $(./memcoin --config /tmp/node3 ordering export)\
     --member $(./memcoin --config /tmp/node4 ordering export)\
     --member $(./memcoin --config /tmp/node5 ordering export)\
     --member $(./memcoin --config /tmp/node6 ordering export)

elif [ $1 -eq 7 ]
then

./memcoin --config /tmp/node1 ordering setup\
    --member $(./memcoin --config /tmp/node1 ordering export)\
    --member $(./memcoin --config /tmp/node2 ordering export)\
    --member $(./memcoin --config /tmp/node3 ordering export)\
     --member $(./memcoin --config /tmp/node4 ordering export)\
     --member $(./memcoin --config /tmp/node5 ordering export)\
     --member $(./memcoin --config /tmp/node6 ordering export)\
     --member $(./memcoin --config /tmp/node7 ordering export)

elif [ $1 -eq 10 ]
then
./memcoin --config /tmp/node1 ordering setup\
    --member $(./memcoin --config /tmp/node1 ordering export)\
    --member $(./memcoin --config /tmp/node2 ordering export)\
    --member $(./memcoin --config /tmp/node3 ordering export)\
     --member $(./memcoin --config /tmp/node4 ordering export)\
     --member $(./memcoin --config /tmp/node5 ordering export)\
     --member $(./memcoin --config /tmp/node6 ordering export)\
     --member $(./memcoin --config /tmp/node7 ordering export)\
     --member $(./memcoin --config /tmp/node8 ordering export)\
     --member $(./memcoin --config /tmp/node9 ordering export)\
     --member $(./memcoin --config /tmp/node10 ordering export)
elif [ $1 -eq 13 ]
then

./memcoin --config /tmp/node1 ordering setup\
    --member $(./memcoin --config /tmp/node1 ordering export)\
    --member $(./memcoin --config /tmp/node2 ordering export)\
    --member $(./memcoin --config /tmp/node3 ordering export)\
     --member $(./memcoin --config /tmp/node4 ordering export)\
     --member $(./memcoin --config /tmp/node5 ordering export)\
     --member $(./memcoin --config /tmp/node6 ordering export)\
     --member $(./memcoin --config /tmp/node7 ordering export)\
     --member $(./memcoin --config /tmp/node8 ordering export)\
     --member $(./memcoin --config /tmp/node9 ordering export)\
     --member $(./memcoin --config /tmp/node10 ordering export)\
     --member $(./memcoin --config /tmp/node11 ordering export)\
     --member $(./memcoin --config /tmp/node12 ordering export)\
     --member $(./memcoin --config /tmp/node13 ordering export)
     

else
  echo "give 3,4,5,6,10  or 13 "
fi

echo "${GREEN}[3/7]${NC} setup access rights on each node"

fromb=1

while [ $fromb -le $to ]
do
./memcoin --config /tmp/node$fromb access add \
    --identity $(crypto bls signer read --path private.key --format BASE64_PUBKEY)

((fromb++))
done


echo "${GREEN}[4/7]${NC} grant access on the chain"

./memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT


fromb=1

while [ $fromb -le $to ]
do

./memcoin --config /tmp/node1 pool add\
    --key private.key\
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access\
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract --args go.dedis.ch/dela.Evoting\
    --args access:grant_command --args all\
    --args access:identity --args $(crypto bls signer read --path /tmp/node$fromb/private.key --format BASE64_PUBKEY)\
    --args access:command --args GRANT


((fromb++))
done

# The following is not needed anymore thanks to the "postinstall" functionality.
# See #65.

# echo "${GREEN}[5/7]${NC} init shuffle"
# ./memcoin --config /tmp/node1 shuffle init --signer /tmp/node1/private.key
# ./memcoin --config /tmp/node2 shuffle init --signer /tmp/node2/private.key
# ./memcoin --config /tmp/node3 shuffle init --signer /tmp/node3/private.key

# echo "${GREEN}[6/7]${NC} starting http proxy"
# ./memcoin --config /tmp/node1 proxy start --clientaddr 127.0.0.1:8081
# ./memcoin --config /tmp/node1 e-voting registerHandlers --signer private.key
# ./memcoin --config /tmp/node1 dkg registerHandlers

# ./memcoin --config /tmp/node2 proxy start --clientaddr 127.0.0.1:8082
# ./memcoin --config /tmp/node2 e-voting registerHandlers --signer private.key
# ./memcoin --config /tmp/node2 dkg registerHandlers

# ./memcoin --config /tmp/node3 proxy start --clientaddr 127.0.0.1:8083
# ./memcoin --config /tmp/node3 e-voting registerHandlers --signer private.key
# ./memcoin --config /tmp/node3 dkg registerHandlers

# If an election is created with ID "deadbeef" then one must set up DKG
# on each node before the election can proceed:
# ./memcoin --config /tmp/node1 dkg init --electionID deadbeef
# ./memcoin --config /tmp/node2 dkg init --electionID deadbeef
# ./memcoin --config /tmp/node3 dkg init --electionID deadbeef
# ./memcoin --config /tmp/node1 dkg setup --electionID deadbeef
