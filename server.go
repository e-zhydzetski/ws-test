package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/e-zhydzetski/ws-test/api"
	"github.com/e-zhydzetski/ws-test/util"
	"github.com/e-zhydzetski/ws-test/wsserver"
	"github.com/golang/protobuf/proto"
	"golang.org/x/sync/errgroup"
	"time"
)

func NewEPollServer(ctx context.Context, listenAddr string, poolSize int, pingInterval time.Duration) error {
	g, ctx := errgroup.WithContext(ctx)

	monitor := util.NewMonitor(ctx, g)

	wsserver.StartEPollServer(ctx, g, listenAddr, poolSize, newServerSessionRegistrar(monitor, pingInterval))
	return g.Wait()
}

func NewSimpleServer(ctx context.Context, listenAddr string, pingInterval time.Duration) error {
	g, ctx := errgroup.WithContext(ctx)

	monitor := util.NewMonitor(ctx, g)

	wsserver.StartSimpleServer(ctx, g, listenAddr, newServerSessionRegistrar(monitor, pingInterval))
	return g.Wait()
}

func newServerSessionRegistrar(monitor *util.Monitor, pingInterval time.Duration) func(wsserver.Session) {
	return func(sess wsserver.Session) {
		clientMsgs := make(chan proto.Message)

		go func() {
			err := func() error {
				var clientID string

				select {
				case msg := <-clientMsgs:
					id, ok := msg.(*api.ClientID)
					if !ok {
						return errors.New("invalid client id message")
					}
					clientID = id.GetId()
					monitor.Connect()
				case <-sess.Context().Done():
					return sess.Context().Err()
				}

				for {
					err := sess.Write(&api.ServerPing{
						ClientId: clientID,
					})
					if err != nil {
						return err
					}
					monitor.Write()

					select {
					case msg := <-clientMsgs:
						pong, ok := msg.(*api.ClientPong)
						if !ok || pong.GetClientId() != clientID {
							return fmt.Errorf("invalid client pong message")
						}
						monitor.Read()
					case <-sess.Context().Done():
						return sess.Context().Err()
					}

					select {
					case <-time.After(pingInterval):
					case <-sess.Context().Done():
						return sess.Context().Err()
					}
				}
			}()
			if err != nil {
				monitor.Error(err)
			}
		}()

		sess.OnRead(func(msg proto.Message) error {
			clientMsgs <- msg
			return nil
		})
	}
}
