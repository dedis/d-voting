#!/bin/sh


set -o errexit

command -v tmux >/dev/null 2>&1 || { echo >&2 "tmux is not on your PATH!"; exit 1; }


# Launch session
TMUX_SESSION_NAME="d-voting-test"

tmux list-sessions | rg "^$TMUX_SESSION_NAME:" >/dev/null 2>&1 && { echo >&2 "A session with the name $TMUX_SESSION_NAME already exists; kill it and try again"; exit 1; }

tmux new -s $TMUX_SESSION_NAME -d


# Clean containers and tmp dir
if [[ $(docker ps -a -q) ]]; then
    docker rm -f $(docker ps -a -q)
fi

rm -rf ./nodedata    
mkdir nodedata

# Clean logs
if [ -d "./log" ] 
then
    rm -rf ./log
    mkdir log
else
    mkdir log
fi

# Create docker network (only run once)
# docker network create --driver bridge evoting-net


N_NODE=$1
vals=($(seq 1 1 $N_NODE))

pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3

#Start docker images and bind volume
for i in "${vals[@]}"
do
    docker run -d -it --env LLVL=info --name node$i --network evoting-net -v "$(pwd)"/nodedata:/tmp  --publish $(( 9080+$i )):9080 node
    tmux new-window -t $TMUX_SESSION_NAME
    tmux send-keys -t $TMUX_SESSION_NAME:$i.0 "eval docker exec node$i memcoin --config /tmp/node$i start --postinstall \
  --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node$i):2001 | tee ./log/node$i.log" C-m
done

tmux new-window -t $TMUX_SESSION_NAME

tmux a

