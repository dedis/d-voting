#!/bin/bash
if [[ "$1" == "" ]]; then
  echo "please run $0 with form-ID"
  exit 1
fi

votes=${2:-120}

npx ts-node src/cli.ts vote -f https://dvoting.c4dt.org \
	-e $1 \
	-b $votes -a 111111
