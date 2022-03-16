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
useradd -M -r -g dedis -d /var/opt/dedis/dvoting \
-s /usr/sbin/nologin -c "D-Voting user" dvoting
fi

# modify user to be in these groups
usermod -aG dedis dvoting
usermod -aG sudo dvoting

# unikernel (qemu_guest script) requires to be started as root
echo "dvoting ALL=(ALL) NOPASSWD:ALL" | tee /etc/sudoers.d/dvoting
