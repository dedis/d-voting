#!/bin/bash

export COMPOSE_FILE=${2:-../docker-compose/docker-compose.dev.yml}

docker compose exec backend npx cli addAdmin --sciper $1
docker compose stop backend
docker compose rm backend
docker compose up backend
