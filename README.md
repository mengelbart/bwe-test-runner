# BWE Test Runner

This is a simple application to setup some docker containers to run the test
cases described in [RFC 8867 (Test Cases for Evaluating Congestion Control for
Interactive Real-Time Media)](https://www.rfc-editor.org/rfc/rfc8867.html).

## Dependencies

* Docker

## Running

Run `go run main.go run` to execute a test. This will start some docker
containers as described in the next section.

## Network topology

The routers will setup NAT forwarding and routing to the sender/receiver, so
that the endpoints can talk to each other. The endpoints need to run the setup
script to add routes to the other nets.

This setup ensures, that communication is done via the two routers. Once all
containers are started, the controller application will start adding network
impairments to the sharednet-interfaces of leftrouter and rightrouter using
`tc-netem` and `tc-tbf`.

![network setup](/network.png)

## Test cases

Currently only the first testcase from [RFC
8867](https://www.rfc-editor.org/rfc/rfc8867.html) is implemented.

## Implementations

### Implement a test application

