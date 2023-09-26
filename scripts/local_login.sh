export FRONTEND_URL=http://127.0.0.1:3000

if ! [[ -f cookies.txt ]]; then
  curl -k "$FRONTEND_URL/api/get_dev_login" -X GET -c cookies.txt
fi
