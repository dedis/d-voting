#!/usr/bin/env bash

rm -rf ./nodedata    
mkdir nodedata

docker-compose up

./setup_docker.sh

docker exec nodetest go test -v -run ^TestScenario$ github.com/dedis/d-voting/integration

