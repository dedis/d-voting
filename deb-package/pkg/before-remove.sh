#!/bin/sh

# stop service
systemctl stop dvoting.service
systemctl stop exporter.service

rm -f /usr/bin/memcoin
