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

cp ../dvoting $INSTALL_DIR/

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
    --url https://dedis.github.com/c4dt/dvoting \
    --description 'D-Voting package' \
    --package dist .

# cleanup
rm -rf ./deb
