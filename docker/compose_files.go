package docker

const composeFileStringOne = `
version: "3.8"

services:
  leftrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: leftrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.2
      leftnet:
        ipv4_address: 172.26.0.2
    cap_add:
      - NET_ADMIN

  rightrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: rightrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.3
      rightnet:
        ipv4_address: 172.27.0.2
    cap_add:
      - NET_ADMIN

  sender_0:
    image: $SENDER_0
    tty: true
    container_name: sender_0
    hostname: sender_0
    environment:
      ROLE: 'sender'
      ARGS: $SENDER_0_ARGS
      RECEIVER: '172.27.0.3'
    volumes:
      - ./$OUTPUT/forward_0/send_log:/log
      - ./input:/input:ro
    networks:
      leftnet:
        ipv4_address: 172.26.0.3
    cap_add:
      - NET_ADMIN

  receiver_0:
    image: $RECEIVER_0
    tty: true
    container_name: receiver_0
    hostname: receiver_0
    environment:
      ROLE: 'receiver'
      ARGS: $RECEIVER_0_ARGS
      SENDER: '172.26.03'
    volumes:
      - ./$OUTPUT/forward_0/receive_log:/log
      - ./$OUTPUT/forward_0/sink:/output
    networks:
      rightnet:
        ipv4_address: 172.27.0.3
    cap_add:
      - NET_ADMIN

networks:
  sharednet:
    name: sharednet
    ipam:
      driver: default
      config:
        - subnet: 172.25.0.0/16
  leftnet:
    name: leftnet
    ipam:
      driver: default
      config:
        - subnet: 172.26.0.0/16
  rightnet:
    name: rightnet
    ipam:
      driver: default
      config:
        - subnet: 172.27.0.0/16
`

const composeFileStringTwo = `
version: "3.8"

services:
  leftrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: leftrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.2
      leftnet:
        ipv4_address: 172.26.0.2
    cap_add:
      - NET_ADMIN

  rightrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: rightrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.3
      rightnet:
        ipv4_address: 172.27.0.2
    cap_add:
      - NET_ADMIN

  sender_0:
    image: $SENDER_0
    tty: true
    container_name: sender_0
    hostname: sender_0
    environment:
      ROLE: 'sender'
      ARGS: $SENDER_0_ARGS
      RECEIVER: '172.27.0.3'
    volumes:
      - ./$OUTPUT/forward_0/send_log:/log
      - ./input:/input:ro
    networks:
      leftnet:
        ipv4_address: 172.26.0.3
    cap_add:
      - NET_ADMIN

  receiver_0:
    image: $RECEIVER_0
    tty: true
    container_name: receiver_0
    hostname: receiver_0
    environment:
      ROLE: 'receiver'
      ARGS: $RECEIVER_0_ARGS
      SENDER: '172.26.03'
    volumes:
      - ./$OUTPUT/forward_0/receive_log:/log
      - ./$OUTPUT/sink:/output
    networks:
      rightnet:
        ipv4_address: 172.27.0.3
    cap_add:
      - NET_ADMIN

  sender_1:
    image: $SENDER_1
    tty: true
    container_name: sender_1
    hostname: sender_1
    environment:
      ROLE: 'sender'
      ARGS: $SENDER_1_ARGS
      RECEIVER: '172.27.0.4'
    volumes:
      - ./$OUTPUT/forward_1/send_log:/log
      - ./input:/input:ro
    networks:
      leftnet:
        ipv4_address: 172.26.0.4
    cap_add:
      - NET_ADMIN

  receiver_1:
    image: $RECEIVER_1
    tty: true
    container_name: receiver_1
    hostname: receiver_1
    environment:
      ROLE: 'receiver'
      ARGS: $RECEIVER_1_ARGS
      SENDER: '172.26.0.4'
    volumes:
      - ./$OUTPUT/forward_1/receive_log:/log
      - ./$OUTPUT/forward_1/sink:/output
    networks:
      rightnet:
        ipv4_address: 172.27.0.4
    cap_add:
      - NET_ADMIN

networks:
  sharednet:
    name: sharednet
    ipam:
      driver: default
      config:
        - subnet: 172.25.0.0/16
  leftnet:
    name: leftnet
    ipam:
      driver: default
      config:
        - subnet: 172.26.0.0/16
  rightnet:
    name: rightnet
    ipam:
      driver: default
      config:
        - subnet: 172.27.0.0/16
`

const composeFileStringThree = `
version: "3.8"

services:
  leftrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: leftrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.2
      leftnet:
        ipv4_address: 172.26.0.2
    cap_add:
      - NET_ADMIN

  rightrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: rightrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.3
      rightnet:
        ipv4_address: 172.27.0.2
    cap_add:
      - NET_ADMIN

  sender_0:
    image: $SENDER_0
    tty: true
    container_name: sender_0
    hostname: sender_0
    environment:
      ROLE: 'sender'
      ARGS: $SENDER_0_ARGS
      RECEIVER: '172.27.0.3'
    volumes:
      - ./$OUTPUT/forward_0/send_log:/log
      - ./input:/input:ro
    networks:
      leftnet:
        ipv4_address: 172.26.0.3
    cap_add:
      - NET_ADMIN

  receiver_0:
    image: $RECEIVER_0
    tty: true
    container_name: receiver_0
    hostname: receiver_0
    environment:
      ROLE: 'receiver'
      ARGS: $RECEIVER_0_ARGS
      SENDER: '172.26.03'
    volumes:
      - ./$OUTPUT/forward_0/receive_log:/log
      - ./$OUTPUT/forward_0/sink:/output
    networks:
      rightnet:
        ipv4_address: 172.27.0.3
    cap_add:
      - NET_ADMIN

  sender_1:
    image: $SENDER_1
    tty: true
    container_name: sender_1
    hostname: sender_1
    environment:
      ROLE: 'sender'
      ARGS: $SENDER_0_ARGS
      RECEIVER: '172.26.0.4'
    volumes:
      - ./$OUTPUT/backward_0/send_log:/log
      - ./input:/input:ro
    networks:
      rightnet:
        ipv4_address: 172.27.0.4
    cap_add:
      - NET_ADMIN

  receiver_1:
    image: $RECEIVER_1
    tty: true
    container_name: receiver_1
    hostname: receiver_1
    environment:
      ROLE: 'receiver'
      ARGS: $RECEIVER_1_ARGS
      SENDER: '172.27.04'
    volumes:
      - ./$OUTPUT/backward_0/receive_log:/log
      - ./$OUTPUT/backward_0/sink:/output
    networks:
      leftnet:
        ipv4_address: 172.26.0.4
    cap_add:
      - NET_ADMIN

networks:
  sharednet:
    name: sharednet
    ipam:
      driver: default
      config:
        - subnet: 172.25.0.0/16
  leftnet:
    name: leftnet
    ipam:
      driver: default
      config:
        - subnet: 172.26.0.0/16
  rightnet:
    name: rightnet
    ipam:
      driver: default
      config:
        - subnet: 172.27.0.0/16
`

const composeFileStringSix = `
version: "3.8"

services:
  leftrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: leftrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.2
      leftnet:
        ipv4_address: 172.26.0.2
    cap_add:
      - NET_ADMIN

  rightrouter:
    image: engelbart/router
    tty: true
    command: bash
    container_name: rightrouter
    networks:
      sharednet:
        ipv4_address: 172.25.0.3
      rightnet:
        ipv4_address: 172.27.0.2
    cap_add:
      - NET_ADMIN

  sender_0:
    image: $SENDER_0
    tty: true
    container_name: sender_0
    hostname: sender_0
    environment:
      ROLE: 'sender'
      ARGS: $SENDER_0_ARGS
      RECEIVER: '172.27.0.3'
    volumes:
      - ./$OUTPUT/forward_0/send_log:/log
      - ./input:/input:ro
    networks:
      leftnet:
        ipv4_address: 172.26.0.3
    cap_add:
      - NET_ADMIN

  receiver_0:
    image: $RECEIVER_0
    tty: true
    container_name: receiver_0
    hostname: receiver_0
    environment:
      ROLE: 'receiver'
      ARGS: $RECEIVER_0_ARGS
      SENDER: '172.26.03'
    volumes:
      - ./$OUTPUT/forward_0/receive_log:/log
      - ./$OUTPUT/forward_0/sink:/output
    networks:
      rightnet:
        ipv4_address: 172.27.0.3
    cap_add:
      - NET_ADMIN

  tcp_sender:
    image: engelbart/endpoint
    tty: true
    container_name: tcp_sender
    hostname: tcp_sender
    environment:
      ARGS:
      ROLE: 'sender'
      RECEIVER: '172.27.0.4'
    volumes:
      - ./$OUTPUT/forward_1/send_log:/log
    networks:
      leftnet:
        ipv4_address: 172.26.0.4
    cap_add:
      - NET_ADMIN
    entrypoint: /iperf.sh -i 0.2 -C cubic -J --logfile /log/tcp.log -t 100

  tcp_receiver:
    image: engelbart/endpoint
    tty: true
    container_name: tcp_receiver
    hostname: tcp_receiver
    environment:
      ROLE: 'receiver'
      SENDER: '172.26.0.4'
    volumes:
      - ./$OUTPUT/forward_1/receive_log:/log
    networks:
      rightnet:
        ipv4_address: 172.27.0.4
    cap_add:
      - NET_ADMIN
    entrypoint: /iperf.sh -1 -i 0.2 -J --logfile /log/tcp.log

networks:
  sharednet:
    name: sharednet
    ipam:
      driver: default
      config:
        - subnet: 172.25.0.0/16
  leftnet:
    name: leftnet
    ipam:
      driver: default
      config:
        - subnet: 172.26.0.0/16
  rightnet:
    name: rightnet
    ipam:
      driver: default
      config:
        - subnet: 172.27.0.0/16
`
