#!/bin/bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
. "$SCRIPT_DIR/local_login.sh"

echo "add form"
RESP=$(curl -sk "$FRONTEND_URL/api/evoting/forms" -X POST -H 'Content-Type: application/json' -b cookies.txt --data-raw $'{"Configuration":{"Title":{"En":"Colours","Fr":"","De":""},"Scaffold":[{"ID":"A7GsJxVJ","Title":{"En":"Colours","Fr":"","De":""},"Order":["GhidLIfw"],"Ranks":[],"Selects":[{"ID":"GhidLIfw","Title":{"En":"RGB","Fr":"","De":"RGB"},"MaxN":3,"MinN":1,"Choices":["{\\"en\\":\\"Red\\",\\"de\\":\\"Rot\\"}","{\\"en\\":\\"Green\\",\\"de\\":\\"Gr\xfcn\\"}","{\\"en\\":\\"Blue\\",\\"de\\":\\"Blau\\"}"],"Hint":{"En":"","Fr":"","De":"RGB"}}],"Texts":[],"Subjects":[]}]}}')
FORMID=$(echo "$RESP" | jq -r .FormID)

echo "add permissions - it's normal to have a timeout error after this command"
curl -k "$FRONTEND_URL/api/evoting/authorizations" -X PUT -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID '$ARGS.named')" -m 1

echo "initialize nodes"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors" -X POST -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID --arg Proxy http://localhost:2001 '$ARGS.named')"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors" -X POST -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID --arg Proxy http://localhost:2003 '$ARGS.named')"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors" -X POST -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID --arg Proxy http://localhost:2005 '$ARGS.named')"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors" -X POST -H 'Content-Type: application/json' -b cookies.txt --data "$(jq -cn --arg FormID $FORMID --arg Proxy http://localhost:2007 '$ARGS.named')"
sleep 2

echo "set node up"
curl -k "$FRONTEND_URL/api/evoting/services/dkg/actors/$FORMID" -X PUT -H 'Content-Type: application/json' -b cookies.txt --data-raw '{"Action":"setup","Proxy":"http://localhost:2001"}'
sleep 8

echo "open election"
curl -k "$FRONTEND_URL/api/evoting/forms/$FORMID" -X PUT -b cookies.txt -H 'Content-Type: application/json' --data-raw '{"Action":"open"}' >/dev/null
echo "Form with ID $FORMID has been set up"

echo "The following forms are available:"
curl -sk "$FRONTEND_URL/api/evoting/forms" -X GET -H 'Content-Type: application/json' -b cookies.txt
echo

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
      --data-raw '{"Ballot":[{"K":[216,252,154,214,23,73,218,203,111,141,124,186,222,48,108,44,151,176,234,112,44,42,242,255,168,82,143,252,103,34,171,20],"C":[172,150,64,201,211,61,72,9,170,205,101,70,226,171,48,39,111,222,242,2,231,221,139,13,189,101,87,151,120,87,183,199]}],"UserID":null}' \
      >/dev/null
    sleep 1
  done
fi
