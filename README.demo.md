# DELA nodes

1. `git clone https://github.com/dedis/d-voting.git`
2. `git checkout demo-20230912`
3. `cd docker-compose`
4. `export COMPOSE_FILE=docker-compose.dela.yml`
5.

```
cat << EOF > .env
PUBLIC_KEY=7a97bfc968c74fbf0553fdec1a97d8265bd3a32e31edd61e296ea701a009147e
PROXYPORT=8080
HOSTNAME=XXX
EOF
```
6. `docker compose up -d`
8. get `TOKEN_ARGS_WORKER_{0,1}` from DELA leader and

```
docker compose exec dela-worker-0 memcoin --config /data/node minogrpc join --address //dela-worker-0:2000 $TOKEN_ARGS_WORKER_0;
docker compose exec dela-worker-1 memcoin --config /data/node minogrpc join --address //dela-worker-0:2000 $TOKEN_ARGS_WORKER_1;
```

9. give the output of

```
docker compose exec dela-worker-0 /bin/bash -c 'LLVL=error memcoin --config /data/node ordering export';
docker compose exec dela-worker-1 /bin/bash -c 'LLVL=error memcoin --config /data/node ordering export';
```

to DELA leader
