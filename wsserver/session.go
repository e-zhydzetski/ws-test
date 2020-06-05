package wsserver

import (
	"context"
	"github.com/e-zhydzetski/ws-test/util"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"sync"
)

type Session interface {
	Context() context.Context
	Close(err error)
	Write(msg proto.Message) error
	OnRead(f func(msg proto.Message) error)
}

func newSession(ctx context.Context, conn net.Conn) *session {
	ctx, cancel := context.WithCancel(ctx)

	return &session{
		name:   nameConn(conn),
		ctx:    ctx,
		cancel: cancel,
		mx:     &sync.Mutex{},
		conn:   conn,
	}
}

type session struct {
	name     string
	ctx      context.Context
	cancel   context.CancelFunc
	mx       *sync.Mutex
	conn     net.Conn
	readFunc func(msg proto.Message) error
}

func (s *session) Context() context.Context {
	return s.ctx
}

func (s *session) Close(err error) {
	_ = s.conn.Close()
	s.cancel()
}

func (s *session) Write(msg proto.Message) error {
	bytes, err := util.MarshalProtoMessage(msg)
	if err != nil {
		return err
	}
	s.mx.Lock()
	defer s.mx.Unlock()
	return wsutil.WriteServerBinary(s.conn, bytes)
}

// called only by server, not thread safe
func (s *session) receive() error {
	bytes, err := wsutil.ReadClientBinary(s.conn)
	if err != nil {
		return err
	}
	if s.readFunc == nil {
		log.Println("No read func on", s.name)
		return nil
	}
	msg, err := util.UnmarshalProtoMessage(bytes)
	if err != nil {
		return err
	}
	return s.readFunc(msg)
}

func (s *session) OnRead(f func(msg proto.Message) error) {
	s.readFunc = f
}
