package tcp

import (
	"bufio"
	cim "cirno-im"
	"cirno-im/wire/endian"
	"io"
	"net"
)

type Frame struct {
	OpCode  cim.OpCode
	Payload []byte
}

func (f *Frame) SetOpCode(code cim.OpCode) {
	f.OpCode = code
}

func (f *Frame) GetOpCode() cim.OpCode {
	return f.OpCode
}

func (f *Frame) GetPayload() []byte {
	return f.Payload
}

func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

type TcpConn struct {
	net.Conn
	rd *bufio.Reader
	wr *bufio.Writer
}

func NewConn(conn net.Conn) *TcpConn {
	return &TcpConn{
		Conn: conn,
		rd:   bufio.NewReaderSize(conn, 4096),
		wr:   bufio.NewWriterSize(conn, 1024),
	}
}

func NewConnWithRW(conn net.Conn, rd *bufio.Reader, wr *bufio.Writer) *TcpConn {
	return &TcpConn{
		Conn: conn,
		rd:   rd,
		wr:   wr,
	}
}
func (c *TcpConn) ReadFrame() (cim.Frame, error) {
	opCode, err := endian.ReadUint8(c.Conn)
	if err != nil {
		return nil, err
	}
	payload, err := endian.ReadBytes(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{
		OpCode:  cim.OpCode(opCode),
		Payload: payload,
	}, nil
}

func (c *TcpConn) WriteFrame(OpCode cim.OpCode, data []byte) error {
	return WriteFrame(c.Conn, OpCode, data)
}

func (c *TcpConn) Flush() error {
	return nil
}

func WriteFrame(w io.Writer, opCode cim.OpCode, payload []byte) error {
	err := endian.WriteUint8(w, uint8(opCode))
	if err != nil {
		return err
	}
	err = endian.WriteBytes(w, payload)
	if err != nil {
		return err
	}
	return nil
}
