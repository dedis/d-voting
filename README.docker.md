# D-Voting/DELA setup w/ Docker Compose

## Overview

The relevant files are:

* docker-compose.yml
* .env
* the Dockerfiles in ./Dockerfiles

You need to create a local .env file with the following content:

```
DELA_REPLICAS=3                                 # number of DELA nodes to create
DELA_PORT_RANGE=2000-2002                       # DELA ports
DATABASE_USER=dvoting                           # database user to create
DATABASE_PASSWORD=...                           # database password to use
DATABASE_HOST=db                                # database host (must correspond to the name of the service in docker-compose.yml)
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

After running `docker-compose up`, you need to run the script `dela.sh` to create the DELA network.
