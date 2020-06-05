package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/e-zhydzetski/ws-test/util"
	"log"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // G108, http server starts only by --debug flag
	"os"
	"time"
)

var buildTime string //nolint:gochecknoglobals // build tag

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Build time:", buildTime)
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	var debug string
	flag.StringVar(&debug, "debug", "", "debug server listen address. Empty - disabled. Use :8000 for http://127.0.0.1:8000/debug/pprof")

	var serverMode bool
	flag.BoolVar(&serverMode, "server", false, "server mode, client otherwise")
	var epollServer bool
	flag.BoolVar(&epollServer, "server-epoll", false, "epoll-based server, goroutines otherwise")
	var listenAddr string
	flag.StringVar(&listenAddr, "server-listen-addr", ":8888", "server listen address")
	var serverWSPool int
	flag.IntVar(&serverWSPool, "server-ws-pool", 10, "max size of server web-socket goroutine pool")
	var serverPingInterval time.Duration
	flag.DurationVar(&serverPingInterval, "server-ping-interval", 10*time.Second, "server ping interval")

	var connectAddr string
	flag.StringVar(&connectAddr, "client-connect-addr", "ws://127.0.0.1:8888", "client connect to address")
	var clientThreads int
	flag.IntVar(&clientThreads, "client-threads", 100, "parallel client sessions")

	flag.Parse()

	if debug != "" {
		log.Println("Debug server listen on " + debug)
		go func() {
			err := http.ListenAndServe(debug, nil)
			if err != nil {
				log.Println("Debug server error:", err)
			}
		}()
	}

	ctx := context.Background()
	ctx = util.GracefulContext(ctx)

	if err := func() error {
		if serverMode {
			if epollServer {
				return NewEPollServer(ctx, listenAddr, serverWSPool, serverPingInterval)
			}
			return NewSimpleServer(ctx, listenAddr, serverPingInterval)
		}
		return NewClient(ctx, connectAddr, clientThreads)
	}(); err != nil {
		log.Println("Error:", err)
	}
}
