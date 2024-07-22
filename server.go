package cim

import (
	"context"
	"net"
	"time"
)

type Server interface {
	SetAcceptor(Acceptor)
	SetMessageListener(MessageListener)
	SetStateListener(StateListener)
	SetReadWait(time.Duration)
	SetChannelMap(ChannelMap)

	// Start 服务启动
	Start() error
	// Push 消息到指定的Channel内
	Push(string, []byte) error
	// ShutDown 服务下线
	ShutDown(ctx context.Context) error
}

type Acceptor interface {
	// Accept TODO 在Server的Start()方法中监听到连接之后，就要调用这个Accept方法让上层业务处理握手相关工作.
	Accept(Conn, time.Duration) (string, error)
}

type MessageListener interface {
	// Receive TODO 设置一个消息监听器。
	Receive(Agent, []byte)
}

type StateListener interface {
	// DisConnect TODO 设置一个状态监听器，将连接断开的事件上报给业务层，让业务层可以实现一些逻辑处理。
	DisConnect(string) error
}

type Agent interface {
	ID() string
	Push([]byte) error
}

// Conn todo 对net.Conn的二次封装，将读写操作叶枫装进Conn内
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}

// Channel todo Channel是对连接进一步的封装
type Channel interface {
	Conn
	Agent
	Close() error
	ReadLoop(listener MessageListener) error
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
}

type Client interface {
	ID() string
	Name() string
	Connect(string) error
	// SetDialer 设置一个拨号器，这个方法会在Connect中被调用，完成连接的建立和握手。
	SetDialer(Dialer)
	Send([]byte) error
	Read() (Frame, error)
	Close()
}

type Dialer interface {
	// DialAndHandshake 对调用以及握手的抽象
	DialAndHandshake(DialerContext) (net.Conn, error)
}

type DialerContext struct {
	Id      string
	Name    string
	Address string
	Timeout time.Duration
}

type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

// Frame 将TCP和WS抽象为同一个接口，对上层业务来说，只需要关心Opcode和Payload。
type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}
