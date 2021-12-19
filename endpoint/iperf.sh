#!/bin/bash

/setup.sh

set -x

if [ "$ROLE" == "sender" ]; then
    echo "Starting iperf3 sender..."
    iperf3 -c $RECEIVER "$@"
else
    echo "Starting iperf3 receiver."
    iperf3 -s "$@"
fi

