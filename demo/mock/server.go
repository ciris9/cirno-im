package mock

import (
	"cirno-im"
	"cirno-im/logger"
	"cirno-im/naming"
	"cirno-im/tcp"
	"cirno-im/websocket"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"time"
)

type ServerDemo struct{}

func (s *ServerDemo) Start(id, protocol, addr string) {
	go func() {
		logger.Println(http.ListenAndServe(":6060", nil))
	}()
	var srv cim.Server
	service := &naming.DefaultService{
		Id:       id,
		Protocol: protocol,
	}
	if protocol == "ws" {
		srv = websocket.NewServer(addr, service)
	} else if protocol == "tcp" {
		srv = tcp.NewServer(addr, service)
	}

	handler := &ServerHandler{}

	srv.SetReadWait(time.Minute)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err := srv.Start()
	if err != nil {
		panic(err)
	}
}

// ServerHandler ServerHandler
type ServerHandler struct {
}

// Accept this connection
func (h *ServerHandler) Accept(conn cim.Conn, timeout time.Duration) (string, error) {
	// 1. 读取：客户端发送的鉴权数据包
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}
	logger.Info("recv", frame.GetOpCode())
	// 2. 解析：数据包内容就是userId
	userID := string(frame.GetPayload())
	// 3. 鉴权：这里只是为了示例做一个fake验证，非空
	if userID == "" {
		return "", errors.New("user id is invalid")
	}
	return userID, nil
}

// Receive default listener
func (h *ServerHandler) Receive(ag cim.Agent, payload []byte) {
	ack := string(payload) + " from server "
	_ = ag.Push([]byte(ack))
}

// DisConnect default listener
func (h *ServerHandler) DisConnect(id string) error {
	logger.Warnf("disconnect %s", id)
	return nil
}
