SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
. "$SCRIPT_DIR/local_vars.sh"

if ! [[ -f cookies.txt ]]; then
  curl -k "$FRONTEND_URL/api/get_dev_login/$REACT_APP_SCIPER_ADMIN" -X GET -c cookies.txt -o /dev/null -s
fi
