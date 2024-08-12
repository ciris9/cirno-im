package tcp

import (
	"cirno-im/constants"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"cirno-im"
	"cirno-im/logger"
)

// ClientOptions ClientOptions
type ClientOptions struct {
	Heartbeat time.Duration //登录超时
	ReadWait  time.Duration //读超时
	WriteWait time.Duration //写超时
}

// Client is a websocket implement of the terminal
type Client struct {
	sync.Mutex
	cim.Dialer
	once    sync.Once
	id      string
	name    string
	conn    cim.Conn
	state   int32
	options ClientOptions
	Meta    map[string]string
}

// NewClient NewClient
func NewClient(id, name string, opts ClientOptions) cim.Client {
	return NewClientWithProps(id, name, make(map[string]string), opts)
}

func NewClientWithProps(id, name string, meta map[string]string, opts ClientOptions) cim.Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = constants.DefaultWriteWait
	}
	if opts.ReadWait == 0 {
		opts.ReadWait = constants.DefaultReadWait
	}

	cli := &Client{
		id:      id,
		name:    name,
		options: opts,
		Meta:    meta,
	}
	return cli
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Name() string {
	return c.name
}

// Connect to server
func (c *Client) Connect(addr string) error {
	// 这里是一个CAS原子操作，对比并设置值，是并发安全的。
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}

	rawconn, err := c.Dialer.DialAndHandshake(cim.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: constants.DefaultLoginWait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if rawconn == nil {
		return fmt.Errorf("conn is nil")
	}
	c.conn = NewConn(rawconn)

	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbeatloop()
			if err != nil {
				logger.WithField("module", "tcp.client").Warn("heartbeatloop stopped - ", err)
			}
		}()
	}
	return nil
}

// SetDialer 设置握手逻辑
func (c *Client) SetDialer(dialer cim.Dialer) {
	c.Dialer = dialer
}

// Send data to connection
func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return fmt.Errorf("connection is nil")
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.WriteFrame(cim.OpBinary, payload)
	if err != nil {
		return err
	}
	return c.conn.Flush()
}

// Close 关闭
func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}
		// graceful close connection
		_ = c.conn.WriteFrame(cim.OpClose, nil)
		c.conn.Flush()

		c.conn.Close()
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

func (c *Client) Read() (cim.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if c.options.Heartbeat > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := c.conn.ReadFrame()
	if err != nil {
		return nil, err
	}
	if frame.GetOpCode() == cim.OpClose {
		return nil, errors.New("remote side close the channel")
	}
	return frame, nil
}

func (c *Client) heartbeatloop() error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		// 发送一个ping的心跳包给服务端
		if err := c.ping(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ping() error {
	logger.WithField("module", "tcp.client").Tracef("%s send ping to server", c.id)

	err := c.conn.WriteFrame(cim.OpPing, nil)
	if err != nil {
		return err
	}
	return c.conn.Flush()
}

// ID return id
func (c *Client) ServiceID() string {
	return c.id
}

// Name Name
func (c *Client) ServiceName() string {
	return c.name
}
func (c *Client) GetMetadata() map[string]string { return c.Meta }
