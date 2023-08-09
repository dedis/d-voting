# D-Voting/DELA setup w/ Docker Compose

## Overview

The relevant files are:

* `docker-compose.yml`
* `.env`
* the Dockerfiles in ./Dockerfiles

You need to create a local .env file with the following content:

```
DELA_REPLICAS=3                                 # number of Dela nodes to deploy
DELA_NODE_URL=http://localhost:8080             # Dela node URL (port must be in DELA_PROXY_PORT_RANGE)
DELA_PORT_RANGE=2000-2002                       # Dela ports (at least DELA_REPLICAS ports)
DELA_PROXY_PORT_RANGE=8080-8082                 # Dela proxy ports (at least DELA_REPLICAS ports)
DATABASE_USERNAME=dvoting                       # choose any PostgreSQL username
DATABASE_PASSWORD=                              # choose any PostgreSQL password
DATABASE_HOST=db                                # PostgreSQL host
FRONT_END_URL=http://localhost:3000             # frontend URL
BACKEND_HOST=backend                            # backend host
BACKEND_PORT=5000                               # backend port
SESSION_SECRET=                                 # choose any secret
PUBLIC_KEY=                                     # pre-generated key pair
PRIVATE_KEY=                                    # pre-generated key pair
PROXYPORT=8080                                  # port of Dela proxy (must be one in DELA_PROXY_PORT_RANGE)
```

You can then run

```
docker compose up
```

to build the images and build and run the containers.

Use

```
docker compose down
```

to shut off, and

```
docker compose down -v
```

to delete the volumes (this will reset your instance).

## Post-install commands

1. run the script `DELA_REPLICAS=... init_dela.sh` to initialize the DELA network with `DELA_REPLICAS set to the same value as in .env`
2. run `docker exec -it d-voting-backend-1 /bin/bash` to connect to the backend
3. execute `node -e 'require("./dbUtils").addAdmin("./dvoting-users", <sciper>)'` with your Sciper to add yourself as admin
