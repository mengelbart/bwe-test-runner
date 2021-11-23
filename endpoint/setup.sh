#!/bin/bash

set -x

IP=$(hostname -I | cut -f1 -d" ")
GATEWAY="${IP%.*}.2"

ip route add 172.25.0.0/16 via $GATEWAY
ip route add 172.26.0.0/16 via $GATEWAY
ip route add 172.27.0.0/16 via $GATEWAY

