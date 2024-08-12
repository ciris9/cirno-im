package tcp

import (
	"bufio"
	cim "cirno-im"
	"net"
)

// Server is a websocket implement of the Server
type UpGrader struct {
}

// NewServer NewServer
func NewServer(listen string, service cim.ServiceRegistration, options ...cim.ServerOption) cim.Server {
	return cim.NewServer(listen, service, new(UpGrader), options...)
}

func (u *UpGrader) Name() string {
	return "tcp.Server"
}

func (u *UpGrader) Upgrade(rawconn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (cim.Conn, error) {
	conn := NewConnWithRW(rawconn, rd, wr)
	return conn, nil
}
