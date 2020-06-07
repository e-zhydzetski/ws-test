# Web-socket test util

[![Build Status](https://cloud.drone.io/api/badges/e-zhydzetski/ws-test/status.svg)](https://cloud.drone.io/e-zhydzetski/ws-test)
[![Go Report Card](https://goreportcard.com/badge/github.com/e-zhydzetski/ws-test)](https://goreportcard.com/report/github.com/e-zhydzetski/ws-test)
[![Docker](https://img.shields.io/docker/pulls/zhydzetski/ws-test)](https://hub.docker.com/r/zhydzetski/ws-test)

## About
Tool for test web-socket tuned infrastructure.  

### Binary/container working modes:
* Server
  * EPoll-based, unix systems only
  * Goroutine-based
* Client
  * Goroutine-based

### Communication protocol
Messages format: [proto](https://github.com/e-zhydzetski/ws-test/blob/master/api/messages.proto)

1. Server listen specified tcp port
2. Client simultaneously opens configured number of web-socket connections
3. For each established connection:
    1. Client sends Msg with ClientID, that contains index of client's connection
    2. Server periodically sends Msg with ServerPing
    3. Client answers by Msg with ClientPong
4. On the client connect error, retries with random-based exponential backoff

## Install
* Build binary: `go get -u github.com/e-zhydzetski/ws-test/cmd/ws-test`
* Pull docker image: `docker pull zhydzetski/ws-test:latest`

## Usage
```
$ docker run --rm zhydzetski/ws-test
 Build time: 20200607111259
 Usage of ./ws-test:
   -client-connect-addr string
         client connect to address (default "ws://127.0.0.1:8888")
   -client-threads int
         parallel client sessions (default 100)
   -debug string
         debug server listen address. Empty - disabled. Use :8000 for http://127.0.0.1:8000/debug/pprof
   -server
         server mode, client otherwise
   -server-epoll
         epoll-based server, goroutines otherwise
   -server-listen-addr string
         server listen address (default ":8888")
   -server-ping-interval duration
         server ping interval (default 10s)
   -server-ws-pool int
         max size of server web-socket goroutine pool (default 10)
```
... is equal to `./ws-test --help`

Working example with [docker-compose](https://github.com/e-zhydzetski/ws-test/blob/master/docker-compose.yml)
```
$ ./docker-compose up
Starting ws-test_server_1 ... done
Recreating ws-test_client_1 ... done
Attaching to ws-test_server_1, ws-test_client_1
server_1  | 2020/06/07 12:15:58 Debug server listen on :8000
server_1  | 2020/06/07 12:15:58 epoll websocket is listening on [::]:8888
server_1  | 2020/06/07 12:15:58 WS pool size: 10
server_1  | 2020/06/07 12:15:59 accept error: schedule error: timed out; retrying in 5ms
server_1  | 2020/06/07 12:15:59 accept error: schedule error: timed out; retrying in 5ms
client_1  | 2020/06/07 12:16:00 Interval: 26.792042ms ; Connects: 100 ; Writes: 100 ; Reads: 100
server_1  | 2020/06/07 12:16:00 Interval: 21.241459ms ; Connects: 100 ; Writes: 100 ; Reads: 100
client_1  | 2020/06/07 12:16:10 Interval: 19.680935ms ; Connects: 0 ; Writes: 100 ; Reads: 100
server_1  | 2020/06/07 12:16:10 Interval: 19.915359ms ; Connects: 0 ; Writes: 100 ; Reads: 100
Gracefully stopping... (press Ctrl+C again to force)
Stopping ws-test_client_1   ... done
Stopping ws-test_server_1   ... done
```

## Based on
* web-socket library: https://github.com/gobwas/ws
* protocol buffers library: https://github.com/gogo/protobuf