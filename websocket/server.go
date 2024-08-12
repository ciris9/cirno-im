package websocket

import (
	"bufio"
	cim "cirno-im"
	"github.com/gobwas/ws"
	"net"
)

// Server is a websocket implement of the Server
type Upgrader struct {
}

// NewServer NewServer
func NewServer(listen string, service cim.ServiceRegistration, options ...cim.ServerOption) cim.Server {
	return cim.NewServer(listen, service, new(Upgrader), options...)
}

func (u *Upgrader) Name() string {
	return "websocket.Server"
}

func (u *Upgrader) Upgrade(rawconn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (cim.Conn, error) {
	_, err := ws.Upgrade(rawconn)
	if err != nil {
		return nil, err
	}
	conn := NewConnWithRW(rawconn, rd, wr)
	return conn, nil
}
