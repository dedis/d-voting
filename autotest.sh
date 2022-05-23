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
    ./runNode.sh -n $N_NODE -a false -d true
    sleep 3
    # Setup block chain
    ./setupnNode.sh -n $N_NODE -d true
    sleep 3
    # Start scenario test and keep logs 
    NNODES=$N_NODE go test -v -run ^TestScenario$ github.com/dedis/d-voting/integration -count=1 | tee ./log/log/gotest.log
    sleep 3
    # Stop the test
    ./kill_test.sh
    # move log to a new directory named logkill
    mv log/log log/logkill$i
    mkdir -p log/log
done

echo "Test $RUN_TIMES times test and succeeded $(grep -c ok  ./log*/gotest.log| awk 'BEGIN{FS=":"}{x+=$2}END{print x}') times"
