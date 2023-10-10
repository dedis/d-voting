#!/bin/bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
. "$SCRIPT_DIR/local_login.sh"

echo "add form"
RESP=$(curl -sk "$FRONTEND_URL/api/evoting/forms" -X POST -H 'Content-Type: application/json' -b cookies.txt --data-raw '{"Configuration":{"MainTitle":"{\"en\":\"Colours\",\"fr\":\"\",\"de\":\"\"}","Scaffold":[{"ID":"5DRhKsY2","Title":"Colours","TitleFr":"","TitleDe":"","Order":["d0mSUfpv"],"Ranks":[],"Selects":[{"ID":"d0mSUfpv","Title":"{\"en\":\"RGB\",\"fr\":\"\",\"de\":\"\"}","TitleDe":"","TitleFr":"","MaxN":1,"MinN":1,"Choices":["{\"en\":\"Red\"}","{\"en\":\"Green\"}","{\"en\":\"Blue\"}"],"ChoicesMap":{},"Hint":"{\"en\":\"\",\"fr\":\"\",\"de\":\"\"}","HintFr":"","HintDe":""}],"Texts":[],"Subjects":[]}]}}')
FORMID=$(echo "$RESP" | jq -r .FormID)
echo "add permissions - it's normal to have a timeout error after this command"
curl -k "$FRONTEND_URL/api/evoting/authorizations" -X PUT -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID '$ARGS.named')" -m 1

echo "initialize nodes"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors" -X POST -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID --arg Proxy http://localhost:2001 '$ARGS.named')"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors" -X POST -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID --arg Proxy http://localhost:2003 '$ARGS.named')"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors" -X POST -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID --arg Proxy http://localhost:2005 '$ARGS.named')"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors" -X POST -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID --arg Proxy http://localhost:2007 '$ARGS.named')"

echo "set node up"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors/$FORMID" -X PUT -H 'Content-Type: application/json' -b cookies.txt --data-raw '{"Action":"setup","Proxy":"http://localhost:2001"}'

echo "open election"
sleep 8
curl -k "$FRONTEND_URL/api/evoting/forms/$FORMID" -X PUT -b cookies.txt -H 'Content-Type: application/json' --data-raw '{"Action":"open"}' >/dev/null
echo "Form with ID $FORMID has been set up"

echo "Adding $REACT_APP_SCIPER_ADMIN to voters"
tmpfile=$(mktemp)
echo -n "$REACT_APP_SCIPER_ADMIN" >"$tmpfile"
(cd web/backend && npx ts-node src/cli.ts addVoters --election-id $FORMID --scipers-file "$tmpfile")
echo "Restarting backend to take into account voters"
"$SCRIPT_DIR/run_local.sh" backend
. "$SCRIPT_DIR/local_login.sh"

if [[ "$1" ]]; then
  echo "Casting $1 votes"
  for i in $(seq $1); do
    echo "Casting vote #$i"
    curl -sk "$FRONTEND_URL/api/evoting/forms/$FORMID/vote" -X POST -H 'Content-Type: Application/json' \
      -H "Origin: $FRONTEND_URL" -b cookies.txt \
      --data-raw '{"Ballot":[{"K":[54,152,33,11,201,233,212,157,204,176,136,138,54,213,239,198,79,55,71,26,91,244,98,215,208,239,48,253,195,53,192,94],"C":[105,121,87,164,68,242,166,194,222,179,253,231,213,63,34,66,212,41,214,175,178,83,229,156,255,38,55,234,168,222,81,185]}],"UserID":null}' \
      >/dev/null
    sleep 1
  done
fi
