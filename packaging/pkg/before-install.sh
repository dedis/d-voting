#!/bin/sh

# create dvoting group
if ! getent group dvoting >/dev/null; then
    groupadd -r dvoting
fi

# create dedis group
if ! getent group dedis >/dev/null; then
    groupadd -r dedis
fi

# create dvoting user
if ! getent passwd dvoting >/dev/null; then
useradd -M -r -g dvoting -d /var/opt/dedis/dvoting \
-s /usr/sbin/nologin -c "D-Voting" dvoting
fi

# modify user to be in these groups
usermod -aG dedis dvoting
