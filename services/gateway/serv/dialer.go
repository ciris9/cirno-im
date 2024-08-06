package serv

import (
	cim "cirno-im"
	"cirno-im/logger"
	"cirno-im/tcp"
	"cirno-im/wire/pkt"
	"google.golang.org/protobuf/proto"
	"net"
)

type TcpDialer struct {
	ServiceID string
}

func NewDialer(serviceId string) cim.Dialer {
	return &TcpDialer{
		ServiceID: serviceId,
	}
}

func (d *TcpDialer) DialAndHandshake(ctx cim.DialerContext) (net.Conn, error) {
	// 拨号建立连接
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	req := &pkt.InnerHandshakeRequest{ServiceID: d.ServiceID}
	logger.Infof("send req %v", req)
	//将自己的ServiceId发送给对方
	bts, _ := proto.Marshal(req)
	err = tcp.WriteFrame(conn, cim.OpBinary, bts)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
