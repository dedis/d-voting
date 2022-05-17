#! /bin/sh

RUN_TIMES=5
N_NODE=15
vals=($(seq 1 1 $RUN_TIMES))


for i in "${vals[@]}"
do
    echo "Test $i with $N_NODE nodes"
    ./launch_containers.sh $N_NODE false
    sleep 3
    ./setup_containers.sh $N_NODE
    sleep 3
    DVOTING_NB_NODE=$N_NODE go test -v -run ^TestScenario$ github.com/dedis/d-voting/integration -count=1 | tee ./log/gotest.log
    sleep 3
    ./kill_test.sh
    mv log logkill$i
    mkdir log
done