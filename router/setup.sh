#!/bin/bash

iptables -t nat -A POSTROUTING -j MASQUERADE

ip route add 172.26.0.0/16 via 172.25.0.2
ip route add 172.27.0.0/16 via 172.25.0.3

_term() {
  echo "Caught SIGTERM signal!"
  kill -TERM "$child" 2>/dev/null
}

trap _term SIGTERM

sleep infinity &

child=$!
wait

