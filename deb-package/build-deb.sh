#! /usr/bin/env bash
set -xe

# cleanup previous installations
rm -rf deb

# create binaries dir
INSTALL_DIR="deb/opt/dedis/dvoting/bin"
mkdir -p $INSTALL_DIR

# copy binaries to deb/opt/dedis/dvoting/bin
UKAPP_DIR="../contracts/evoting/unikernel/apps/combine_shares"
cp $UKAPP_DIR/qemu-guest $INSTALL_DIR
cp $UKAPP_DIR/build/combine_shares_kvm-x86_64 $INSTALL_DIR

DVOTING_CLI_DIR=".."
cp $DVOTING_CLI_DIR/memcoin $INSTALL_DIR

# Prometheus Node Exporter
NE_DIR="deb/opt/exporter"
NE_VERSION="1.3.1"
mkdir -p ${NE_DIR}
wget https://github.com/prometheus/node_exporter/releases/download/v${NE_VERSION}/node_exporter-${NE_VERSION}.linux-amd64.tar.gz
tar xfz node_exporter-${NE_VERSION}.linux-amd64.tar.gz
mv node_exporter-${NE_VERSION}.linux-amd64/* ${NE_DIR}/
rm -rf node_exporter-${NE_VERSION}.linux-amd64*

# add config files
cp -a pkg/etc deb
cp -a pkg/lib deb
cp -a pkg/opt deb
cp -a pkg/var deb

# add folders
mkdir -p deb/var/log/dedis/dvoting

# adjust permissions
find deb ! -perm -a+r -exec chmod a+r {} \;

# get version from git without v prefix
GITVERSION=$(git describe --abbrev=0 --tags)
VERSION=${GITVERSION:1}
if [[ -z "${ITERATION}" ]]
then
  ITERATION="0"
fi

# fpm needs an existing output directory
OUTPUT_DIR="dist"
mkdir -p $OUTPUT_DIR

fpm \
    --force -t deb -a all -s dir -C deb -n dedis-dvoting -v ${VERSION} \
    --iteration ${ITERATION} \
    --deb-user dvoting \
    --deb-group dvoting \
    --depends bridge-utils \
    --depends fuse \
    --depends qemu-kvm \
    --depends libvirt-daemon-system \
    --depends net-tools \
    --depends sgabios \
    --depends socat \
    --depends uuid-runtime \
    --depends virtinst \
    --depends virt-manager \
    --before-install pkg/before-install.sh \
    --after-install pkg/after-install.sh \
    --before-remove pkg/before-remove.sh \
    --after-remove pkg/after-remove.sh \
    --url https://dedis.github.com/dedis/dvoting \
    --description 'D-Voting package for Unicore' \
    --package dist .

# cleanup
rm -rf ./deb
