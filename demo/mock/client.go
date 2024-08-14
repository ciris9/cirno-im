package mock

import (
	"context"
	"net"
	"time"

	"cirno-im"
	"cirno-im/logger"
	"cirno-im/tcp"
	"cirno-im/websocket"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// ClientDemo Client demo
type ClientDemo struct {
}

func (c *ClientDemo) Start(userID, protocol, addr string) {
	var cli cim.Client

	// step1: 初始化客户端
	if protocol == "ws" {
		cli = websocket.NewClient(userID, "client", websocket.ClientOptions{})
		// set dialer
		cli.SetDialer(&WebsocketDialer{})
	} else if protocol == "tcp" {
		cli = tcp.NewClient("test1", "client", tcp.ClientOptions{})
		cli.SetDialer(&TCPDialer{})
	}

	// step2: 建立连接
	err := cli.Connect(addr)
	if err != nil {
		logger.Error(err)
	}
	count := 5
	go func() {
		// step3: 发送消息然后退出
		for i := 0; i < count; i++ {
			err := cli.Send([]byte("hello"))
			if err != nil {
				logger.Error(err)
				return
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	// step4: 接收消息
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Info(err)
			break
		}
		if frame.GetOpCode() != cim.OpBinary {
			continue
		}
		recv++
		logger.Warnf("%s receive message [%s]", cli.ID(), frame.GetPayload())
		if recv == count { // 接收完消息
			break
		}
	}
	//退出
	cli.Close()
}

// WebsocketDialer WebsocketDialer
type WebsocketDialer struct {
}

// DialAndHandshake DialAndHandshake
func (d *WebsocketDialer) DialAndHandshake(ctx cim.DialerContext) (net.Conn, error) {
	// 1 调用ws.Dial拨号
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	// 2. 发送用户认证信息，示例就是userid
	err = wsutil.WriteClientBinary(conn, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}
	// 3. return conn
	return conn, nil
}

// TCPDialer TCPDialer
type TCPDialer struct {
}

// DialAndHandshake DialAndHandshake
func (d *TCPDialer) DialAndHandshake(ctx cim.DialerContext) (net.Conn, error) {
	logger.Info("start dial: ", ctx.Address)
	// 1 调用net.Dial拨号
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	// 2. 发送用户认证信息，示例就是userid
	err = tcp.WriteFrame(conn, cim.OpBinary, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}
	// 3. return conn
	return conn, nil
}
