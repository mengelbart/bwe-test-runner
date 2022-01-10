#!/bin/bash

/setup.sh

set -x

_term() {
  echo "Caught SIGTERM signal!"
  kill -TERM "$child" 2>/dev/null
}

trap _term SIGTERM

if [ "$ROLE" == "sender" ]; then
    echo "Starting iperf3 sender..."
    iperf3 -c $RECEIVER "$@" &
    child=$!
    wait "$child"
else
    echo "Starting iperf3 receiver."
    iperf3 -s "$@" &
    child=$!
    wait "$child"
fi

