#!/bin/sh

# fix permissions
chown -R root:root /opt/dedis
chmod -R 755 /opt/dedis

chown -R root:root /lib/systemd/system

SERVICE=dvoting.service

# Inspired from Debian packages (e.g. /var/lib/dpkg/info/openssh-server.postinst)
# was-enabled defaults to true, so new installations run enable.
if deb-systemd-helper --quiet was-enabled ${SERVICE}; then
    # Enables the unit on first installation, creates new
    # symlinks on upgrades if the unit file has changed.
    deb-systemd-helper enable ${SERVICE} >/dev/null || true
else
    # Update the statefile to add new symlinks (if any), which need to be
    # cleaned up on purge. Also remove old symlinks.
    deb-systemd-helper update-state ${SERVICE} >/dev/null || true
fi

systemctl start ${SERVICE}
