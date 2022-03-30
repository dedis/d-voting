#!/bin/sh

# stop service
systemctl stop dvoting.service
systemctl stop dvoting-uk.service

rm -f /usr/bin/memcoin
rm -f /usr/bin/node_exporter
