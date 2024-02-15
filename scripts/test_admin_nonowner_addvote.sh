#!/bin/bash

# This script tests that an admin who is not the owner of a form
# cannot add voters to the form.
# It also tests that the admin who created the form can actually add
# voters to the form.

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
"$SCRIPT_DIR/run_local.sh"

. "$SCRIPT_DIR/local_vars.sh"
SECOND_ADMIN=123321
echo "Adding $SECOND_ADMIN to admin"
(cd web/backend && npx ts-node src/cli.ts addAdmin --sciper $SECOND_ADMIN | grep -v Executing)

"$SCRIPT_DIR/local_proxies.sh"
"$SCRIPT_DIR/local_forms.sh"

. "$SCRIPT_DIR/formid.env"

tmp_dir=$(mktemp -d)
trap 'rm -rf -- "tmpdir"' EXIT

tmp_cookie_owner="$tmp_dir/cookie_owner"
curl -k "$FRONTEND_URL/api/get_dev_login/$REACT_APP_SCIPER_ADMIN" -X GET -c "$tmp_cookie_owner" -o /dev/null -s
tmp_cookie_nonowner="$tmp_dir/cookie_nonowner"
curl -k "$FRONTEND_URL/api/get_dev_login/$SECOND_ADMIN" -X GET -c "$tmp_cookie_nonowner" -o /dev/null -s

echo "This should fail with an error that we're not allowed"
tmp_output="$tmp_dir/output"
curl -s 'http://localhost:3000/api/add_role' \
  -H 'Content-Type: application/json' \
  --data-raw "{\"userId\":444555,\"subject\":\"$FORMID\",\"permission\":\"vote\"}" \
  -b "$tmp_cookie_nonowner" 2>&1 | tee "$tmp_output"
echo

if ! grep -q "not owner of form" "$tmp_output"; then
  echo
  echo "ERROR: Reply should be 'not owner of form'"
  exit 1
fi

echo "This should pass for the owner of the form"
curl 'http://localhost:3000/api/add_role' \
  -H 'Content-Type: application/json' \
  --data-raw "{\"userId\":444555,\"subject\":\"$FORMID\",\"permission\":\"vote\"}" \
  -b "$tmp_cookie_owner"
echo
