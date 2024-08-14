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

// ReadFrame ReadFrame
func (c *TcpConn) ReadFrame() (cim.Frame, error) {
	opcode, err := endian.ReadUint8(c.rd)
	if err != nil {
		return nil, err
	}
	payload, err := endian.ReadBytes(c.rd)
	if err != nil {
		return nil, err
	}
	return &Frame{
		OpCode:  cim.OpCode(opcode),
		Payload: payload,
	}, nil
}

// WriteFrame WriteFrame
func (c *TcpConn) WriteFrame(code cim.OpCode, payload []byte) error {
	return WriteFrame(c.wr, code, payload)
}

// Flush Flush
func (c *TcpConn) Flush() error {
	return c.wr.Flush()
}

// WriteFrame write a frame to w
func WriteFrame(w io.Writer, code cim.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
	return nil
}
