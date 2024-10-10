# D-Voting/DELA setup w/ Docker Compose

## Overview

The files related to the Docker environment can be found in

* `docker-compose/` (Docker Compose files)
* `Dockerfiles/` (Dockerfiles)
* `scripts/` (helper scripts)

### Setup

It is recommended to use the `run_docker.sh` helper script for setting up and
tearing down the environment as it handles all the necessary intermediary steps
to have a working D-Voting application.

This script needs to be executed at the project's root.

To set up the environment:

```
./scripts/run_docker.sh
```

This will run the subcommands:

- `setup` which will build the images and start the containers
- `init_dela` which will initialize the DELA network
- `local_admin` which will add local admin accounts for testing and debugging
- `local_login` which will set a local cookie that allows for interacting w/ the API via command-line
- `add_proxies` which will set up the DELA node proxies

Each of these subcommands can also be run by invoking the script w/ the subcommand:

```
./scripts/run_docker.sh <subcommand>
```

/!\ The `init_dela` subcommand must only be run exactly **once**.

To tear down the environment:

```
./scripts/run_docker.sh teardown
```

This will:

- remove the local cookie
- stop and remove the containers and their attached volumes
- remove the images

/!\ This command is meant to reset your environment. If you want to stop one or more
containers, use the appropriate `docker compose` commands (see below for using the correct `docker-compose.yml`).

### Docker environment

There are two Docker Compose file you may use:

* `docker-compose/docker-compose.yml` (recommended, default in `.env.example` and `run_docker.sh`), or
* `docker-compose/docker-compose.debug.yml`, which contains some additional debugging tools

To run `docker compose` commands w/ the right `docker-compose.yml`, you need to either run

```
export COMPOSE_FILE=<path to Docker Compose file>
```

or

```
source .env
```
