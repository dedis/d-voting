#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
. "$SCRIPT_DIR/local_login.sh"

echo "add proxies";
curl -sk "$FRONTEND_URL/api/proxies/" -X POST -H 'Content-Type: application/json' -b cookies.txt --data-raw '{"NodeAddr":"localhost:2000","Proxy":"http://localhost:2001"}';
curl -sk "$FRONTEND_URL/api/proxies/" -X POST -H 'Content-Type: application/json' -b cookies.txt --data-raw '{"NodeAddr":"localhost:2002","Proxy":"http://localhost:2003"}';
curl -sk "$FRONTEND_URL/api/proxies/" -X POST -H 'Content-Type: application/json' -b cookies.txt --data-raw '{"NodeAddr":"localhost:2004","Proxy":"http://localhost:2005"}';
curl -sk "$FRONTEND_URL/api/proxies/" -X POST -H 'Content-Type: application/json' -b cookies.txt --data-raw '{"NodeAddr":"localhost:2006","Proxy":"http://localhost:2007"}';
