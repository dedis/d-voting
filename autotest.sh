#! /bin/sh

# This script uses runNode.sh, setupnNode.sh, kill_test.sh to launch multiple 
# times of scenario test with user defined number of nodes. The results of log are kept in directory named logkill$i


POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      echo      "This script uses runNode.sh, setupnNode.sh, kill_test.sh to launch multiple times of scenario test with user defined number of nodes. The results of log are kept in directory named logkill "
      echo      ""
      echo      "Options:"
      echo      "-h  |  --help       program help (this file)"
      echo      "-n  |  --node       number of d-voting nodes"
      echo      "-r  |  --run_time   set how many times we want to run the test"
      exit 0
      ;;
    -n|--node)
      N_NODE="$2"
      shift # past argument
      shift # past value
      ;;
    -r|--run_time)
      RUN_TIMES="$2"
      shift # past argument
      shift # past value
      ;;
    -*|--*)
      echo "Unknown option $1"
      exit 1
      ;;
    *)
      POSITIONAL_ARGS+=("$1") # save positional arg
      shift # past argument
      ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters


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
done

echo "Test $RUN_TIMES times test and succeeded $(grep -c ok  ./log/log*/gotest.log| awk 'BEGIN{FS=":"}{x+=$2}END{print x}') times"
