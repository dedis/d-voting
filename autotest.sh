#! /bin/sh

# This script uses launch_containers.sh, setup_containers.sh, kill_test.sh to launch multiple 
# times of scenario test with user defined number of nodes. The results of log are kept in directory named logkill$i

# Set how many times we want to run the test
RUN_TIMES=15
# Set the number of nodes in the test, currently this number is fixed for all tests
N_NODE=15
vals=($(seq 1 1 $RUN_TIMES))


for i in "${vals[@]}"
do
    echo "Test $i with $N_NODE nodes"
    # Launch nodes
    ./launch_containers.sh $N_NODE false
    sleep 3
    # Setup block chain
    ./setup_containers.sh $N_NODE
    sleep 3
    # Start scenario test and keep logs 
    NNODES=$N_NODE go test -v -run ^TestScenario$ github.com/dedis/d-voting/integration -count=1 | tee ./log/gotest.log
    sleep 3
    # Stop the test
    ./kill_test.sh
    # move log to a new directory named logkill
    mv log logkill$i
    mkdir log
done