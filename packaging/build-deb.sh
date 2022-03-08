#! /usr/bin/env bash
set -xe

# cleanup previous installations
rm -rf deb

mkdir -p deb/opt/dedis/dvoting     # install dir

# get version from the virtual environment
VERSION=$(python setup.py --version)
if [[ -z "${ITERATION}" ]]
then
  ITERATION="0"
fi

# ... ideally, you'd just copy your binary to deb/opt/dedis/dvoting/bin ....

# add config files
cp -a pkg/etc deb
cp -a pkg/lib deb

# add folders
mkdir -p deb/var/log/dedis/dvoting
mkdir -p deb/var/opt/dedis/dvoting

# adjust permissions
find deb ! -perm -a+r -exec chmod a+r {} \;

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
