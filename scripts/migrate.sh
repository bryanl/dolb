#!/bin/bash

if [[ $MIGRATE_URL == "" ]]; then
  echo "set MIGRATE_URL"
  exit 1
fi

MYDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo $MYDIR

migrate -path "${MYDIR}/../db/migrations" $@
