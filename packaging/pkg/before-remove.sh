#!/bin/sh

SERVICE=dvoting.service
INSTALLDIR=/opt/dedis/dvoting

# stop service
systemctl stop ${SERVICE}
