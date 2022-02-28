#!/bin/sh

set -e

gcc -o tests tests.c read_ballots.c crypto.c -lsodium
./tests