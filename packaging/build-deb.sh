#! /usr/bin/env bash
set -xe

# cleanup previous installations
rm -rf deb

# create install dir
INSTALL_DIR="deb/opt/dedis/dvoting/bin"
mkdir -p $(INSTALL_DIR)

# copy binaries to deb/opt/dedis/dvoting/bin
UKAPP_DIR="../contracts/evoting/unikernel/apps/combine_shares"
cp $(UKAPP_DIR)/run $(INSTALL_DIR)
cp $(UKAPP_DIR)/build/combine_shares_kvm-x86_64 $(INSTALL_DIR)

DVOTING_CLI_DIR="../cli/memcoin"
cp $(DVOTING_CLI_DIR)/memcoin $(INSTALL_DIR)

# add config files
cp -a pkg/etc deb
cp -a pkg/lib deb

# add folders
mkdir -p deb/var/log/dedis/dvoting
mkdir -p deb/var/opt/dedis/dvoting

# adjust permissions
find deb ! -perm -a+r -exec chmod a+r {} \;

# get version from the git
VERSION=$(git describe)
if [[ -z "${ITERATION}" ]]
then
  ITERATION="0"
fi

fpm \
    --force -t deb -a all -s dir -C deb -n dedis-dvoting -v ${VERSION} \
    --iteration ${ITERATION} \
    --deb-user dvoting \
    --deb-group dvoting \
    --depends qemu-kvm \
    --depends libvirt-daemon-system \
    --depends libvirt-clients \
    --depends bridge-utils \
    --depends virtinst \
    --depends virt-manager \
    --before-install pkg/before-install.sh \
    --after-install pkg/after-install.sh \
    --before-remove pkg/before-remove.sh \
    --after-remove pkg/after-remove.sh \
    --url https://dedis.github.com/dedis/dvoting \
    --description 'D-Voting package for Unicore' \
    --package dist \
    .

# cleanup
rm -rf ./deb
