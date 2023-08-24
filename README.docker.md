# D-Voting/DELA setup w/ Docker Compose

## Overview

The relevant files are:

* `docker-compose.yml`
* `.env`
* the Dockerfiles in ./Dockerfiles

You need to create a local .env file with the following content:

```
DELA_NODE_URL=http://127.0.0.1:80               # reverse proxy within backend container
DATABASE_USERNAME=dvoting                       # choose any PostgreSQL username
DATABASE_PASSWORD=                              # choose any PostgreSQL password
DATABASE_HOST=db                                # PostgreSQL host *within the Docker network*
DATABASE_PORT=5432                              # PostgreSQL port
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

1. `./init_dela.sh`
2. `docker compose exec backend /bin/bash`
3. execute `npx cli addAdmin --sciper 123455` with your SCIPER to add yourself as admin
5. exit the container and run `docker compose down && docker compose up -d`

## Go debugging environment

To use the Go debugging environment, pass the

```
--file docker-compose.debug.yml
```

flag to all `docker compose` invocations.
