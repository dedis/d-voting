#!/bin/bash

if [ -z "$COOKIE" ]; then
  echo "'COOKIE' variable is not set";
  exit 1;
fi

echo "add proxies";
curl -sk 'https://127.0.0.1:3000/api/proxies/' -X POST -H 'Content-Type: application/json' -H "Cookie: connect.sid=$COOKIE; redirect=/" --data-raw '{"NodeAddr":"localhost:2000","Proxy":"http://localhost:2001"}';
curl -sk 'https://127.0.0.1:3000/api/proxies/' -X POST -H 'Content-Type: application/json' -H "Cookie: connect.sid=$COOKIE; redirect=/" --data-raw '{"NodeAddr":"localhost:2002","Proxy":"http://localhost:2003"}';
curl -sk 'https://127.0.0.1:3000/api/proxies/' -X POST -H 'Content-Type: application/json' -H "Cookie: connect.sid=$COOKIE; redirect=/" --data-raw '{"NodeAddr":"localhost:2004","Proxy":"http://localhost:2005"}';
curl -sk 'https://127.0.0.1:3000/api/proxies/' -X POST -H 'Content-Type: application/json' -H "Cookie: connect.sid=$COOKIE; redirect=/" --data-raw '{"NodeAddr":"localhost:2006","Proxy":"http://localhost:2007"}';
