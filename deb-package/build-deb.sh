#! /usr/bin/env bash
set -xe

# cleanup previous installations
rm -rf deb

# create binaries dir
INSTALL_DIR="deb/opt/dedis/dvoting/bin"
mkdir -p $INSTALL_DIR

# get version from git without v prefix
GITVERSION=$(git describe --abbrev=0 --tags || echo '0.0.0')
VERSION=${GITVERSION:1}
versionFile=$(echo $GITVERSION | tr . _)

cp ../memcoin-linux-amd64-${versionFile} $INSTALL_DIR/memcoin

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
    --depends net-tools \
    --before-install pkg/before-install.sh \
    --after-install pkg/after-install.sh \
    --before-remove pkg/before-remove.sh \
    --after-remove pkg/after-remove.sh \
    --url https://dedis.github.com/dedis/dvoting \
    --description 'D-Voting package' \
    --package dist .

# cleanup
rm -rf ./deb
