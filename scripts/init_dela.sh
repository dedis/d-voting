#!/bin/bash

MEMBERS="";


# share the certificate
for container in dela-worker-1 dela-worker-2 dela-worker-3 dela-worker-4; do
  TOKEN_ARGS=$(docker compose exec dela-worker-0 /bin/bash -c 'LLVL=error dvoting --config /data/node minogrpc token');
  docker compose exec "$container" dvoting --config /data/node minogrpc join --address //dela-worker-0:2000 $TOKEN_ARGS;
done

# create a new chain with the nodes
for container in dela-worker-0 dela-worker-1 dela-worker-2 dela-worker-3 dela-worker-4; do
  # add node to the chain
  MEMBERS="$MEMBERS --member $(docker compose exec $container /bin/bash -c 'LLVL=error dvoting --config /data/node ordering export')";
done
docker compose exec dela-worker-0 dvoting --config /data/node ordering setup $MEMBERS;

# authorize the signer to handle the access contract on each node
for signer in dela-worker-0 dela-worker-1 dela-worker-2 dela-worker-3 dela-worker-4; do
  IDENTITY=$(docker compose exec "$signer" crypto bls signer read --path /data/node/private.key --format BASE64_PUBKEY);
  for node in dela-worker-0 dela-worker-1 dela-worker-2 dela-worker-3 dela-worker-4; do
    docker compose exec "$node" dvoting --config /data/node access add --identity "$IDENTITY";
  done
done

# update the access contract
for container in dela-worker-0 dela-worker-1 dela-worker-2 dela-worker-3 dela-worker-4; do
  IDENTITY=$(docker compose exec "$container" crypto bls signer read --path /data/node/private.key --format BASE64_PUBKEY);
  docker compose exec dela-worker-0 dvoting --config /data/node pool add\
      --key /data/node/private.key\
      --args go.dedis.ch/dela.ContractArg\
      --args go.dedis.ch/dela.Access\
      --args access:grant_id\
      --args 0300000000000000000000000000000000000000000000000000000000000000\
      --args access:grant_contract\
      --args go.dedis.ch/dela.Evoting \
      --args access:grant_command\
      --args all\
      --args access:identity\
      --args $IDENTITY\
      --args access:command\
      --args GRANT
done
