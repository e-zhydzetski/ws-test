package client

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/sync/errgroup"

	"github.com/e-zhydzetski/ws-test/api"
	"github.com/e-zhydzetski/ws-test/internal/util"
)

const ioTimeout = 1 * time.Second // TODO make configurable

func NewClient(ctx context.Context, addr string, threads int) error {
	rand.Seed(time.Now().Unix())

	g, ctx := errgroup.WithContext(ctx)

	monitor := util.NewMonitor(ctx, g)

	g.Go(func() error {
		wg := &sync.WaitGroup{}
		wg.Add(threads)

		for i := 0; i < threads; i++ {
			go func(idx int) {
				err := func() error {
					clientID := strconv.Itoa(idx)
					defer wg.Done()

					threadCtx, cancel := context.WithCancel(ctx)
					defer cancel()

					var conn net.Conn
					retryAfter := time.Duration(rand.Intn(1000)) * time.Millisecond
					for {
						var err error
						conn, _, _, err = ws.Dial(threadCtx, addr)
						if err == nil {
							break
						}
						monitor.Error(fmt.Errorf("connect to server error: %w; retry in %s", err, retryAfter))
						select {
						case <-time.After(retryAfter):
						case <-threadCtx.Done():
							return threadCtx.Err()
						}
						retryAfter = 2 * retryAfter
					}

					go func() {
						<-threadCtx.Done()
						_ = conn.Close()
					}()

					err := writeMessage(conn, &api.ClientID{
						Id: clientID,
					})
					if err != nil {
						return fmt.Errorf("client index send error: %w", err)
					}
					monitor.Connect()

					for {
						msg, err := readMessage(conn)
						if err != nil {
							return fmt.Errorf("server ping receive error: %w", err)
						}
						ping, ok := msg.(*api.ServerPing)
						if !ok || ping.GetClientId() != clientID {
							return fmt.Errorf("invalid server ping message")
						}
						monitor.Read()

						err = writeMessage(conn, &api.ClientPong{
							ClientId: clientID,
						})
						if err != nil {
							return fmt.Errorf("client pong send error: %w", err)
						}
						monitor.Write()
					}
				}()
				if err != nil {
					monitor.Error(err)
				}
			}(i)
		}

		wg.Wait()
		return errors.New("all client worker threads finished")
	})

	return g.Wait()
}

func readMessage(conn net.Conn) (proto.Message, error) {
	_ = conn.SetReadDeadline(time.Time{})
	bytes, err := wsutil.ReadServerBinary(conn)
	if err != nil {
		return nil, err
	}
	return util.UnmarshalProtoMessage(bytes)
}

func writeMessage(conn net.Conn, msg proto.Message) error {
	bytes, err := util.MarshalProtoMessage(msg)
	if err != nil {
		return err
	}
	_ = conn.SetWriteDeadline(time.Now().Add(ioTimeout))
	return wsutil.WriteClientBinary(conn, bytes)
}
