#!/bin/bash

iptables -t nat -A POSTROUTING -j MASQUERADE

ip route add 172.26.0.0/16 via 172.25.0.2
ip route add 172.27.0.0/16 via 172.25.0.3

/bin/bash

