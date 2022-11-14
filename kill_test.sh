#! /bin/sh

# This script kills the tmux session started in start_test.sh and
# removes all the data pertaining to the test.
FRONTEND=true
BACKEND=true

while [[ $# -gt 0 ]]; do
    case $1 in
    -h | --help)
        echo "This script is creating n dela voting nodes"
        echo "Options:"
        echo "-h  |  --help     program help (this file)"
        echo "-f  |  --frontend setup the frontend true/false, by default true"
        echo "-b  |  --backend  setup the backend true/false, by default true"
        exit 0
        ;;
    -f | --frontend)
        FRONTEND="$2"
        shift # past argument
        shift # past value
        ;;
    -b | --backend)
        BACKEND="$2"
        shift # past argument
        shift # past value
        ;;
    -* | --*)
        echo "Unknown option $1"
        exit 1
        ;;
    *)
        POSITIONAL_ARGS+=("$1") # save positional arg
        shift                   # past argument
        ;;
    esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

set -o errexit

command -v tmux >/dev/null 2>&1 || {
    echo >&2 "tmux is not on your PATH!"
    exit 1
}

if [ "$FRONTEND" = true ]; then
    tmux send-keys -t $s:{end} C-c
    tmux kill-window -t $s:{end}
fi
if [ "$BACKEND" = true ]; then
    tmux send-keys -t $s:{end} C-c
fi

rm -rf /tmp/node* && tmux kill-session -t d-voting-test
