source .env;
LEADER_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' d-voting-dela-1);
MEMBERS="";

# share the certificate
for i in $(seq 2 "$DELA_REPLICAS"); do
  TOKEN_ARGS=$(docker exec d-voting-dela-1 /bin/bash -c 'LLVL=error memcoin --config /tmp/node minogrpc token');
  docker exec d-voting-dela-"$i" memcoin --config /tmp/node minogrpc join --address //"$LEADER_IP":2000 $TOKEN_ARGS;
done

# create a new chain with the nodes
for i in $(seq 1 "$DELA_REPLICAS"); do
  # add node to the chain
  MEMBERS="$MEMBERS --member $(docker exec d-voting-dela-$i /bin/bash -c 'LLVL=error memcoin --config /tmp/node ordering export')";
done
docker exec d-voting-dela-1 memcoin --config /tmp/node ordering setup $MEMBERS;

# authorize the signer to handle the access contract on each node
for i in $(seq 1 "$DELA_REPLICAS"); do
  docker exec d-voting-dela-"$i" /bin/bash -c 'memcoin --config /tmp/node access add --identity $(crypto bls signer read --path /data/private.key --format BASE64_PUBKEY)';
done

IDENTITY=$(docker exec d-voting-dela-1 crypto bls signer read --path /data/private.key --format BASE64_PUBKEY);
# update the access contract
docker exec d-voting-dela-1 memcoin --config /tmp/node pool add\
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
