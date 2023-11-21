#!/bin/bash

MEMBERS="";

if [ -z "$COMPOSE_FILE" ]; then
  echo "'COMPOSE_FILE' variable not set";
  exit 1;
fi



# share the certificate
echo "[1/4] add nodes to network";
for container in dela-worker-1 dela-worker-2 dela-worker-3; do
  TOKEN_ARGS=$(docker compose exec dela-worker-0 /bin/bash -c 'LLVL=error dvoting --config /data/node minogrpc token');
  echo "generated token for $container";
  docker compose exec "$container" dvoting --config /data/node minogrpc join --address //dela-worker-0:2000 $TOKEN_ARGS;
  echo "$container joined network";
done

# create a new chain with the nodes
echo "[2/4] create a new chain";
for container in dela-worker-0 dela-worker-1 dela-worker-2 dela-worker-3; do
  # add node to the chain
  MEMBERS="$MEMBERS --member $(docker compose exec $container /bin/bash -c 'LLVL=error dvoting --config /data/node ordering export')";
done
docker compose exec dela-worker-0 dvoting --config /data/node ordering setup $MEMBERS;
echo "created new chain";

# authorize the signer to handle the access contract on each node
echo "[3/4] allow nodes to access contracts on each other";
for signer in dela-worker-0 dela-worker-1 dela-worker-2 dela-worker-3; do
  IDENTITY=$(docker compose exec "$signer" crypto bls signer read --path /data/node/private.key --format BASE64_PUBKEY);
  for node in dela-worker-0 dela-worker-1 dela-worker-2 dela-worker-3; do
    docker compose exec "$node" dvoting --config /data/node access add --identity "$IDENTITY";
    echo "$node allowed $signer to access contract on it";
  done
done

# update the access contract
echo "[4/4] grant permissions to update contract";
for container in dela-worker-0 dela-worker-1 dela-worker-2 dela-worker-3; do
  IDENTITY=$(docker compose exec "$container" crypto bls signer read --path /data/node/private.key --format BASE64_PUBKEY);
  docker compose exec dela-worker-0 dvoting --config /data/node pool add\
      --key /data/node/private.key\
      --args go.dedis.ch/dela.ContractArg\
      --args go.dedis.ch/dela.Access\
      --args access:grant_id\
      --args 45564f54\
      --args access:grant_contract\
      --args go.dedis.ch/dela.Evoting \
      --args access:grant_command\
      --args all\
      --args access:identity\
      --args $IDENTITY\
      --args access:command\
      --args GRANT
  echo "$container has been granted permission to update contract";
done
