#!/usr/bin/env bash

# This script is creating a new chain and setting up the services needed to run
# an evoting system. It ends by starting the http server needed by the frontend
# to communicate with the blockchain. This operation is blocking. It is expected
# that the "memcoin" binary is at the root. You can build it with:
#   go build ./cli/memcoin

# by default run on local
DOCKER=false

POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      echo      "This script is setting n dela voting nodes and granting access on block chain"
      echo      "Options:"
      echo      "-h  |  --help     program help (this file)"
      echo      "-n  |  --node     number of d-voting nodes"
      echo      "-d  |  --docker   launch nodes on docker containers true/false"
      exit 0
      ;;
    -n|--node)
      N_NODE="$2"
      shift # past argument
      shift # past value
      ;;
    -d|--docker)
      DOCKER="$2"
      shift # past argument
      shift # past value
      ;;
    -*|--*)
      echo "Unknown option $1"
      exit 1
      ;;
    *)
      POSITIONAL_ARGS+=("$1") # save positional arg
      shift # past argument
      ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color

if [ "$DOCKER" == false ]; then
    echo "${GREEN}[1/4]${NC} connect nodes"

    from=2
    to=$N_NODE
    while [ $from -le $to ]
    do
    ./memcoin --config /tmp/node$from minogrpc join \
        --address //localhost:2001 $(./memcoin --config /tmp/node1 minogrpc token)

    ((from++))
    done

    echo "${GREEN}[2/4]${NC} create a chain"

    ARRAY=""
    from=1
    to=$N_NODE
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
else
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
        sleep 1  
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

    sleep 1  

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
        sleep 1  
    done   
fi