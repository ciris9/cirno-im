package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"
	"net"
	"net/url"
	"time"
)

func connect(addr string) (*handler, error) {
	_, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	conn, _, _, err := ws.Dial(context.Background(), addr)
	if err != nil {
		return nil, err
	}

	h := handler{
		conn:      conn,
		close:     make(chan struct{}, 1),
		recv:      make(chan []byte, 10),
		heartbeat: time.Second * 10,
	}

	//read
	go func() {
		err := h.readLoop(conn)
		if err != nil {
			logrus.Warn(err)
		}
		// 通知上层
		h.close <- struct{}{}
	}()

	//heartbeat
	go func() {
		err := h.heartBeatLoop()
		if err != nil {
			logrus.Warn(err)
		}
		h.close <- struct{}{}
	}()

	return &h, nil
}

type handler struct {
	conn      net.Conn
	close     chan struct{}
	recv      chan []byte
	heartbeat time.Duration
}

func (h *handler) heartBeatLoop() error {
	logrus.Info("heartBeatLoop start")
	tick := time.NewTicker(h.heartbeat)
	for range tick.C {
		logrus.Info("ping")
		if err := wsutil.WriteClientMessage(h.conn, ws.OpPing, nil); err != nil {
			logrus.Warn(err)
			return err
		}
	}
	return nil
}

func (h *handler) readLoop(conn net.Conn) error {
	logrus.Info("readloop started")
	for {
		frame, err := ws.ReadFrame(conn)
		if err != nil {
			return err
		}
		if frame.Header.OpCode == ws.OpPong {
			_ = h.conn.SetReadDeadline(time.Now().Add(h.heartbeat * 3))
			continue
		}
		if frame.Header.OpCode == ws.OpClose {
			return errors.New("remote side close the channel")
		}
		if frame.Header.OpCode == ws.OpText {
			h.recv <- frame.Payload
		}
	}
}

type StartOptions struct {
	address string
	user    string
}

func run(ctx context.Context, opts *StartOptions) error {
	url := fmt.Sprintf("%s?user=%s", opts.address, opts.user)
	logrus.Info("connect to ", url)
	//连接到服务，并返回hander对象
	h, err := connect(url)
	if err != nil {
		return err
	}
	go func() {
		// 读取消息并显示
		for msg := range h.recv {
			logrus.Info("Receive message:", string(msg))
		}
	}()

	tk := time.NewTicker(time.Second * 6)
	for {
		select {
		case <-tk.C:
			//每6秒发送一个消息
			err := h.sendText("hello")
			if err != nil {
				logrus.Error(err)
			}
		case <-h.close:
			logrus.Printf("connection closed")
			return nil
		}
	}
}

func (h *handler) sendText(msg string) error {
	logrus.Info("send message :", msg)
	return wsutil.WriteClientText(h.conn, []byte(msg))
}
