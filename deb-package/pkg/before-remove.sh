#!/bin/sh

# stop service
systemctl stop dvoting.service
systemctl stop dvoting-uk.service

rm -f /usr/bin/memcoin
