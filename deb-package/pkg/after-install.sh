#!/bin/sh

# fix permissions
# dvoting:dedis will be applied automatically on sub dirs
chown root:root /opt/dedis

# allow ls in sub dirs
chmod 755 /opt/dedis
chmod 755 /opt/exporter

chown root:root /lib/systemd/system

enable_service() {
SERVICE=$1
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
}

enable_service dvoting-uk.service

DVOTING_SERVICE=dvoting.service
enable_service ${DVOTING_SERVICE}
systemctl start ${DVOTING_SERVICE}

ln -s /opt/dedis/dvoting/bin/memcoin /usr/bin/memcoin

enable_service exporter.service
systemctl start exporter.service
