package websocket

import (
	cim "cirno-im"
	"github.com/gobwas/ws"
	"net"
)

type Frame struct {
	raw ws.Frame
}

func (f *Frame) SetOpCode(opCode cim.OpCode) {
	f.raw.Header.OpCode = ws.OpCode(opCode)
}

func (f *Frame) GetOpCode() cim.OpCode {
	return cim.OpCode(f.raw.Header.OpCode)
}

func (f *Frame) SetPayload(payload []byte) {
	f.raw.Payload = payload
}

func (f *Frame) GetPayload() []byte {
	//对于websocket协议，client发送的数据帧会进行masking(掩码化)处理，所以需要对数据进行解码
	if f.raw.Header.Masked {
		ws.Cipher(f.raw.Payload, f.raw.Header.Mask, 0)
	}
	f.raw.Header.Masked = false
	return f.raw.Payload
}

type WsConn struct {
	net.Conn
}

func NewConn(conn net.Conn) *WsConn {
	return &WsConn{conn}
}

func (c *WsConn) ReadFrame() (cim.Frame, error) {
	f, err := ws.ReadFrame(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{f}, nil
}
func (c *WsConn) WriteFrame(code cim.OpCode, payload []byte) error {
	f := ws.NewFrame(ws.OpCode(code), true, payload)
	return ws.WriteFrame(c.Conn, f)
}

func (c *WsConn) Flush() error {
	return nil
}
