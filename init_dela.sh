#!/bin/bash

MEMBERS="";

# create signing keys
for container in dela-leader dela-worker-1 dela-worker-2; do
  docker compose exec "$container" crypto bls signer new --save /data/private.key;
done

# share the certificate
for container in dela-worker-1 dela-worker-2; do
  TOKEN_ARGS=$(docker compose exec dela-leader /bin/bash -c 'LLVL=error memcoin --config /data/node minogrpc token');
  docker compose exec "$container" memcoin --config /data/node minogrpc join --address //dela-leader:2000 $TOKEN_ARGS;
done

# create a new chain with the nodes
for container in dela-leader dela-worker-1 dela-worker-2; do
  # add node to the chain
  MEMBERS="$MEMBERS --member $(docker compose exec $container /bin/bash -c 'LLVL=error memcoin --config /data/node ordering export')";
done
docker compose exec dela-leader memcoin --config /data/node ordering setup $MEMBERS;

# authorize the signer to handle the access contract on each node
IDENTITY=$(docker compose exec dela-leader crypto bls signer read --path /data/private.key --format BASE64_PUBKEY);
for container in dela-leader dela-worker-1 dela-worker-2; do
  docker compose exec "$container" memcoin --config /data/node access add --identity "$IDENTITY";
done

# update the access contract
docker compose exec dela-leader memcoin --config /data/node pool add\
    --key /data/private.key\
    --args go.dedis.ch/dela.ContractArg\
    --args go.dedis.ch/dela.Access\
    --args access:grant_id\
    --args 0200000000000000000000000000000000000000000000000000000000000000\
    --args access:grant_contract\
    --args go.dedis.ch/dela.Value\
    --args access:grant_command\
    --args all\
    --args access:identity\
    --args $IDENTITY\
    --args access:command\
    --args GRANT
