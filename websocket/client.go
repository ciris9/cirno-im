package websocket

import (
	"cirno-im"
	"cirno-im/constants"
	"cirno-im/logger"
	"errors"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type ClientOptions struct {
	Heartbeat time.Duration
	ReadWait  time.Duration
	WriteWait time.Duration
}

type Client struct {
	sync.Mutex
	cim.Dialer
	once    sync.Once
	id      string
	name    string
	conn    net.Conn
	state   int32
	options ClientOptions
	dc      *cim.DialerContext
}

func NewClient(id, name string, opts ClientOptions) *Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = constants.DefaultWriteWait
	}
	if opts.ReadWait == 0 {
		opts.ReadWait = constants.DefaultReadWait
	}
	client := &Client{
		id:      id,
		name:    name,
		options: opts,
	}
	return client
}

func (cli *Client) ID() string {
	return cli.id
}

func (cli *Client) Name() string {
	return cli.name
}

func (cli *Client) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if !atomic.CompareAndSwapInt32(&cli.state, 0, 1) {
		return errors.New("client has connected")
	}
	conn, err := cli.Dialer.DialAndHandshake(cim.DialerContext{
		Id:      cli.id,
		Name:    cli.name,
		Address: addr,
		Timeout: constants.DefaultLoginWait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&cli.state, 1, 0)
		return err
	}
	if conn == nil {
		return errors.New("conn is nil")
	}
	cli.conn = conn
	if cli.options.Heartbeat > 0 {
		go func() {
			err = cli.heartbeatLoop(conn)
			if err != nil {
				logger.Error("heartbeat loop err:", err)
			}
		}()
	}
	return nil
}

func (cli *Client) SetDialer(dialer cim.Dialer) {
	cli.Dialer = dialer
}

func (cli *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&cli.state) == 0 {
		return errors.New("connection is nil")
	}
	cli.Lock()
	defer cli.Unlock()
	err := cli.conn.SetWriteDeadline(time.Now().Add(cli.options.WriteWait))
	if err != nil {
		return err
	}
	return wsutil.WriteClientMessage(cli.conn, ws.OpBinary, payload)
}

func (cli *Client) Close() {
	cli.once.Do(func() {
		if cli.conn != nil {
			return
		}
		err := wsutil.WriteClientMessage(cli.conn, ws.OpClose, nil)
		if err != nil {
			logger.Error("write close:", err)
			return
		}
		err = cli.conn.Close()
		if err != nil {
			logger.Error("close conn:", err)
			return
		}
	})
}

func (cli *Client) Read() (cim.Frame, error) {
	if cli.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if cli.options.Heartbeat > 0 {
		err := cli.conn.SetReadDeadline(time.Now().Add(cli.options.ReadWait))
		if err != nil {
			return nil, err
		}
	}
	frame, err := ws.ReadFrame(cli.conn)
	if err != nil {
		return nil, err
	}
	if frame.Header.OpCode == ws.OpClose {
		return nil, errors.New("remote side closed the channel")
	}
	return &Frame{raw: frame}, nil
}

func (cli *Client) heartbeatLoop(conn net.Conn) error {
	ticker := time.NewTicker(cli.options.Heartbeat)
	for range ticker.C {
		if err := cli.ping(conn); err != nil {
			return err
		}
	}
	return nil
}

func (cli *Client) ping(conn net.Conn) error {
	cli.Lock()
	defer cli.Unlock()
	err := conn.SetWriteDeadline(time.Now().Add(cli.options.WriteWait))
	if err != nil {
		return err
	}
	logger.Tracef("%s send ping to server", cli.id)
	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}
