#!/bin/bash

if [[ $FLEETCTL_TUNNEL == "" ]]; then
  echo "required FLEETCTL_TUNNEL env var"
  exit 1
fi

name=$1

for i in "${@:2}"; do
  unit="${name}@$i"
  echo "restarting ${unit}"
  fleetctl stop $unit
  sleep 2
  fleetctl start $unit
  while true; do
    fleetctl status $unit | grep "active (running)" &> /dev/null
    if [[ $? == 0 ]]; then
      up=1
      break
    fi
  done
done

