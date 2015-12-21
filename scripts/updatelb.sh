#!/bin/bash

if [[ $FLEETCTL_TUNNEL == "" ]]; then
  echo "required FLEETCTL_TUNNEL env var"
  exit 1
fi

for i in "$@"; do
  agent="dolb-agent@$i"
  echo "restarting ${agent}"
  fleetctl stop $agent
  sleep 2
  fleetctl start $agent
  while true; do
    fleetctl status $agent | grep "active (running)" &> /dev/null
    if [[ $? == 0 ]]; then
      up=1
      break
    fi
  done
done

