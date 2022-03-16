#!/bin/sh

# fix permissions
# dvoting:dedis will be applied automatically on sub dirs
chown root:root /opt/dedis

# allow ls in sub dirs
chmod 755 /opt/dedis

# not recursive for nows
chown root:root /lib/systemd/system

SERVICE=unikernel.service

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
