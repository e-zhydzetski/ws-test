package wsserver

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/mailru/easygo/netpoll"
	"golang.org/x/sync/errgroup"
)

type SessionRegistrar func(session Session)

//nolint:funlen // TODO refactor
func StartEPollServer(ctx context.Context, g *errgroup.Group, listenAddr string, poolSize int, registrar SessionRegistrar) {
	g.Go(func() error { // modified part of server.ListenAndServer
		poller, err := netpoll.New(nil)
		if err != nil {
			return err
		}

		ln, err := net.Listen("tcp", listenAddr)
		if err != nil {
			return err
		}
		log.Printf("epoll websocket is listening on %s", ln.Addr().String())

		goPool := NewPool(poolSize)
		log.Println("WS pool size:", poolSize)

		acceptDesc := netpoll.Must(netpoll.HandleListener(
			ln, netpoll.EventRead|netpoll.EventOneShot,
		))
		err = poller.Start(acceptDesc, func(e netpoll.Event) {
			accept := make(chan error, 1)
			err := goPool.ScheduleTimeout(time.Millisecond, func() {
				conn, err := ln.Accept()
				if err != nil {
					accept <- err
					return
				}

				accept <- nil

				_, err = ws.Upgrade(conn)
				if err != nil {
					log.Printf("%s: upgrade error: %v", nameConn(conn), err)
					conn.Close()
					return
				}
				sess := newSession(ctx, connWithTimeout{
					Conn: conn,
					wt:   1 * time.Second, // TODO make configurable
					rt:   1 * time.Second, // TODO make configurable
				})
				registrar(sess)
				desc := netpoll.Must(netpoll.HandleReadOnce(conn))
				err = poller.Start(desc, func(ev netpoll.Event) {
					if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
						// connection closed
						_ = poller.Stop(desc)
						sess.Close(nil)
						return
					}
					goPool.Schedule(func() {
						if err := sess.receive(); err != nil {
							_ = poller.Stop(desc)
							sess.Close(err)
							return
						}
						err = poller.Resume(desc)
						if err != nil {
							log.Printf("resume error: %v", err)
							_ = poller.Stop(desc)
							sess.Close(err)
							return
						}
					})
				})
				if err != nil {
					sess.Close(err)
					return
				}
			})
			if err == nil {
				err = <-accept
			}
			if err != nil {
				if err == ErrScheduleTimeout {
					goto cooldown
				}
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					goto cooldown
				}

				log.Printf("accept error: %v", err)
				return

			cooldown:
				delay := 5 * time.Millisecond
				log.Printf("accept error: %v; retrying in %s", err, delay)
				time.Sleep(delay)
			}

			err = poller.Resume(acceptDesc)
			if err != nil {
				log.Printf("resume error: %v", err)
				return
			}
		})
		if err != nil {
			return err
		}
		<-ctx.Done()
		_ = ln.Close()
		return ctx.Err()
	})
}
