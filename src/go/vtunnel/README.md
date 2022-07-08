# vtunnel

Vtunnel is a virtual tunnel that leverages AF_VSOCK virtual sockets. Vtunnel runs two processes:
 - host process that runs on the host machine
 - peer process that runs inside a Hyper-V VM (e.g. WSL).

## Host

The host process can be configured with an upstream HTTP/TCP server to forward the requests to. The tunnel accepts the incoming vsock requests over AF_VSOCK and for every request it creates a connection to the given upstream server to pipe the data forward.

## Peer

The Peer process starts a TCP server inside the Hyper-V VM and listens for all the incoming requests; once a request is received it forwards it over the AF_SOCK to the host.

```mermaid
flowchart LR;
 subgraph Host["HOST"]
 UpstreamServer("Upstream HTTP/TCP Server")
 HostProcess("Vtunnel Host")
 UpstreamServer  <---> |over TCP| HostProcess
 end
 subgraph VM["WSL VM"]
 Peer("Vtunnel Peer")
 Client("client")
 Peer <--->  |over TCP| Client
 end
 HostProcess <---> |AF_VSOCK| Peer
```
## E2E test

You can simply run the e2e test:
```pwsh
 go test -v .\test\e2e\connectivity_test.go
```

## Manual Testing

 - You will need to build the binaries for both Host and Peer processes:
```bash
 GOOS=windows go build
 GOOS=linux go build
```
- Creat a configuration file, below is an example of a `config.yaml`.
 **Note** same configuration file can be used for both Peer and Host processes.
 ```yaml
 tunnel:
  - handshakePort: 9090
    vsockHostPort: 8989
    peerAddress: 127.0.0.1
    peerPort: 3030
    upstreamServerAddress: 127.0.0.1:4444
 ```
 - Move the `vtunnel` executable to the Hyper-V VM and run the Peer process:
 ```bash
 ./vtunnel peer --config-path config.yaml
 ```
 - Use netcat or a similar approach to run a HTTP/TCP server on the host machine:
 ```pwsh
 python3 -m http.server 4444 --bind 127.0.0.1
 ```
 - Run the host process on windows:
 ```pwsh
 .\vtunnel.exe host --config-path config.yaml
 ```
 - Using Curl or similar utilities send a request to the Peer TCP server inside the VM.
 ```bash
 curl localhost:3030
 ```
