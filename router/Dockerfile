FROM ubuntu:20.04

RUN apt-get update && \
  apt-get install -y \
  net-tools \
  iptables \
  iproute2 \
  iputils-ping \
  iperf3 \
  && \
  apt-get clean


COPY setup.sh .
RUN chmod +x setup.sh

ENTRYPOINT [ "./setup.sh" ]

