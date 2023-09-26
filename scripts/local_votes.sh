#!/bin/bash

if ! [[ "$1" && "$2" ]]; then
  echo "Syntax is: $0 #votes FORMID"
  exit 1
fi

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
. "$SCRIPT_DIR/local_login.sh"

FORMID=$2
echo "Casting $1 votes to form $FORMID"
for i in $(seq $1); do
  echo "Casting vote #$i"
  curl -sk "$FRONTEND_URL/api/evoting/forms/$FORMID/vote" -X POST -H 'Content-Type: Application/json' \
    -H "Origin: $FRONTEND_URL" -b cookies.txt \
    --data-raw '{"Ballot":[{"K":[54,152,33,11,201,233,212,157,204,176,136,138,54,213,239,198,79,55,71,26,91,244,98,215,208,239,48,253,195,53,192,94],"C":[105,121,87,164,68,242,166,194,222,179,253,231,213,63,34,66,212,41,214,175,178,83,229,156,255,38,55,234,168,222,81,185]}],"UserID":null}' \
    >/dev/null
  sleep 1
done
