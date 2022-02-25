#!/bin/sh

set -e

gcc -o tests tests.c crypto.c -lsodium
./tests