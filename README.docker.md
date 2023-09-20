# D-Voting/DELA setup w/ Docker Compose

## Overview

The files related to the Docker environment can be found in

* `docker-compose/` (Docker Compose files)
* `Dockerfiles/` (Dockerfiles)
* `scripts/` (helper scripts)

You also need to either create a `.env` file in the project's root
or point to another environment file using the `--env-file` flag
when running `docker compose`.

The environment file needs to contain

```
DELA_NODE_URL=http://172.19.44.254:8080
DATABASE_USERNAME=dvoting
DATABASE_PASSWORD=XXX                       # choose any PostgreSQL password
DATABASE_HOST=db
DATABASE_PORT=5432
DB_PATH=dvoting                             # LMDB database path
FRONT_END_URL=http://127.0.0.1:3000
BACKEND_HOST=backend
BACKEND_PORT=5000
SESSION_SECRET=XXX                          # choose any secret
PUBLIC_KEY=XXX                              # public key of pre-generated key pair
PRIVATE_KEY=XXX                             # private key of pre-generated key pair
PROXYPORT=8080
```

To run the currently released version, run

```
export COMPOSE_FILE=docker-compose/docker-compose.yml
docker compose up
```

this will pull the images from the GitHub container registry.

If you instead run

```
export COMPOSE_FILE=docker-compose/docker-compose.debug.yml
docker compose up
```

the images will be build locally and you can debug your developments.

/!\ Any subsequent `docker compose` commands must be run with `COMPOSE_FILE` being
set to the Docker Compose file that defines the current environment.

Use

```
docker compose down
```

to shut off, and

```
docker compose down -v
```

to delete the volumes and reset your instance.

## Post-install commands

To set up the DELA network, go to `scripts/` and run

```
./init_dela.sh
```

/!\ This script uses `docker compose` as well, so make sure that `COMPOSE_FILE` is
set to the right value.

To set up the permissions, run

```
docker compose exec backend npx cli addAdmin --sciper XXX
docker compose down && docker compose up -d
```

to add yourself as admin and clear the cached permissions.
