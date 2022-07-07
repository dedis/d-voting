#!/bin/sh
#
# This script uploads .deb packages, creates a snapshot and public a new
# version. It expects as input the folder containing the *.deb files.
#
# Example:
#
#   ./upload-artifacts ./dist/
#
# Inspired from
# https://github.com/aptly-dev/aptly/blob/c9f5763a70ba2227e8bdf31c2f0ab7f481b6f2e0/upload-artifacts.sh
#
# Note that at least one package must have been uploaded from the server, and
# the repo created. For example with the following commands.
#
#   aptly repo create -distribution=squeeze -component=main dvoting-release
#
#   aptly publish snapshot --distribution="squeeze" --gpg-key="XXX" \
#     --passphrase "XXX" -gpg-provider="gpg2" dvoting-0.0.1 s3:apt.dedis.ch:
#

set -e

if [ -z ${1+x} ]; then echo "please give the folder"; exit 1; fi

builds="$1/"
packages=${builds}*.deb
folder=`mktemp -u tmp.XXXXXXXXXXXXXXX`
aptly_user="$APTLY_USER"
aptly_password="$APTLY_PASSWORD"
gpg_passphrase="$GPG_PASSPHRASE" # will be passed over the network, use TLS !
aptly_api="http://128.179.33.123"

gitversion=$(git describe --abbrev=0 --tags)
version=${gitversion:1}

echo "Publishing version '$version' from $builds"

for file in $packages; do
    url=${aptly_api}/api/files/$folder
    echo "Uploading $file -> $url"
    curl -fsS -X POST -u $aptly_user:$aptly_password -F "file=@$file" ${url}
    echo
done

aptly_repository=dvoting-release
aptly_snapshot=dvoting-$version
aptly_published=s3:apt.dedis.ch:/squeeze

echo "Adding packages to $aptly_repository..."
curl -fsS -X POST \
    -u $aptly_user:$aptly_password \
    ${aptly_api}/api/repos/$aptly_repository/file/$folder
echo

echo "Check if snapshot $aptly_snapshot already exists"
res=$(curl -s -o /dev/null -w "%{http_code}" http://128.179.33.123/api/snapshots/$aptly_snapshot)

if [ $res = "404" ]; then
    echo "Creating snapshot $aptly_snapshot from repo $aptly_repository..."
    curl -fsS -X POST \
        -u $aptly_user:$aptly_password \
        -H 'Content-Type: application/json' \
        --data  '{"Name":"'"$aptly_snapshot"'"}' \
        ${aptly_api}/api/repos/$aptly_repository/snapshots
    echo
fi

data='{
    "AcquireByHash": true,
    "Snapshots": [
        {
            "Component": "main",
            "Name": "'"$aptly_snapshot"'"
        }
    ],
    "Signing": {
        "Batch": true,
        "GpgKey": "9AD6DDAC613708D9216294DA5B455ACFE943ED69",
        "Passphrase": "'"$gpg_passphrase"'"
    }
}'

echo "Switching published repo to use snapshot $aptly_snapshot..."

curl -fsS -X PUT -H 'Content-Type: application/json' \
    --data "$data" \
    -u $aptly_user:$aptly_password \
    ${aptly_api}/api/publish/$aptly_published
echo

curl -fsS -X DELETE \
    -u $aptly_user:$aptly_password \
    ${aptly_api}/api/files/$folder
echo
