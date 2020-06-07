package wsserver

import (
	"net"
	"time"
)

func nameConn(conn net.Conn) string {
	return conn.LocalAddr().String() + " > " + conn.RemoteAddr().String()
}

type connWithTimeout struct {
	net.Conn
	rt time.Duration
	wt time.Duration
}

func (d connWithTimeout) Write(p []byte) (int, error) {
	if d.wt != 0 {
		if err := d.Conn.SetWriteDeadline(time.Now().Add(d.wt)); err != nil {
			return 0, err
		}
	}
	return d.Conn.Write(p)
}

func (d connWithTimeout) Read(p []byte) (int, error) {
	if d.rt != 0 {
		if err := d.Conn.SetReadDeadline(time.Now().Add(d.rt)); err != nil {
			return 0, err
		}
	}
	return d.Conn.Read(p)
}
