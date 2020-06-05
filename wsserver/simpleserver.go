package wsserver

import (
	"context"
	"github.com/gobwas/ws"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"time"
)

func StartSimpleServer(ctx context.Context, g *errgroup.Group, listenAddr string, registrar SessionRegistrar) {
	g.Go(func() error {
		ln, err := net.Listen("tcp", listenAddr)
		if err != nil {
			return err
		}
		log.Printf("simple websocket is listening on %s", ln.Addr().String())

		go func() {
			<-ctx.Done()
			_ = ln.Close()
		}()
		for {
			conn, err := ln.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					log.Printf("accept tmp error: %v", err)
					continue
				}
				return err
			}

			_, err = ws.Upgrade(conn)
			if err != nil {
				log.Printf("%s: upgrade error: %v", nameConn(conn), err)
				conn.Close()
				continue
			}

			sess := newSession(ctx, connWithTimeout{
				Conn: conn,
				wt:   1 * time.Second, // TODO make configurable
				rt:   0,               // can't use read timeout in wait model (without events)
			})
			registrar(sess)
			go func() {
				for {
					if err := sess.receive(); err != nil {
						sess.Close(err)
						break
					}
				}
			}()
		}
	})
}
