#!/bin/sh

# stop service
systemctl stop dvoting.service
systemctl stop unikernel.service

rm -f /usr/bin/memcoin
