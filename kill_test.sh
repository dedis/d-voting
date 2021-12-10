#! /bin/sh

# This script kills the tmux session started in start_test.sh and
# removes all the data pertaining to the test.

rm -rf /tmp/node* && tmux kill-session -t d-voting-test
