#!/bin/sh


# Clean containers and tmp dir
if [[ $(docker ps -a -q) ]]; then
    docker rm -f $(docker ps -a -q)
fi

rm -rf ./nodedata    
mkdir nodedata

# Create docker network (only run once)
# docker network create --driver bridge evoting-net


N_NODE=$1
vals=($(seq 1 1 $N_NODE))

pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3

#Start docker images and bind volume
for i in "${vals[@]}"
do
    docker run -d -it --name node$i --network evoting-net -v "$(pwd)"/nodedata:/tmp node
    eval docker exec -d node$i memcoin --config /tmp/node$i start --postinstall \
  --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node$i):2001 
done

## TODO
# See the logs



