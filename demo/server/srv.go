package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
)

// Server is a websocket Server
type Server struct {
	once    sync.Once
	id      string
	address string
	sync.RWMutex
	// 会话列表
	users map[string]net.Conn
}

// NewServer NewServer
func NewServer(id, address string) *Server {
	return newServer(id, address)
}

func newServer(id, address string) *Server {
	return &Server{
		id:      id,
		address: address,
		users:   make(map[string]net.Conn, 100),
	}
}

// Start server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	log := logrus.WithFields(logrus.Fields{
		"module": "Server",
		"listen": s.address,
		"id":     s.id,
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// step1. 升级
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}
		//step2. 读取userId
		user := r.URL.Query().Get("user")
		if user == "" {
			conn.Close()
			log.Println(err)
			return
		}
		log.Println("user:", user)
		//step3. 添加到会话管理中
		old, ok := s.addUser(user, conn)
		if ok {
			// 断开旧的连接
			old.Close()
		}
		log.Infof("user %s in", user)

		go func(user string, conn net.Conn) {
			//step4. 读取消息
			err := s.readLoop(user, conn)
			if err != nil {
				log.Error(err)
			}
			conn.Close()
			//step5. 连接断开，删除用户
			s.delUser(user)

			log.Infof("connection of %s closed", user)
		}(user, conn)

	})
	log.Infoln("started")
	return http.ListenAndServe(s.address, mux)
}

func (s *Server) addUser(user string, conn net.Conn) (net.Conn, bool) {
	s.Lock()
	defer s.Unlock()
	old, ok := s.users[user] //返回旧的连接
	s.users[user] = conn     //缓存
	return old, ok
}

func (s *Server) delUser(user string) {
	s.Lock()
	defer s.Unlock()
	delete(s.users, user)
}

// Shutdown Shutdown
func (s *Server) Shutdown() {
	s.once.Do(func() {
		s.Lock()
		defer s.Unlock()
		for _, conn := range s.users {
			conn.Close()
		}
	})
}

func (s *Server) readLoop(user string, conn net.Conn) error {
	readWait := time.Minute * 2
	for {
		_ = conn.SetReadDeadline(time.Now().Add(readWait))
		frame, err := ws.ReadFrame(conn)
		if err != nil {
			return err
		}
		if frame.Header.OpCode == ws.OpPing {
			_ = wsutil.WriteServerMessage(conn, ws.OpPong, nil)
			continue
		}
		if frame.Header.OpCode == ws.OpClose {
			return errors.New("remote side close the conn")
		}

		if frame.Header.Masked {
			ws.Cipher(frame.Payload, frame.Header.Mask, 0)
		}
		// 接收文本帧内容
		if frame.Header.OpCode == ws.OpText {
			go s.handle(user, string(frame.Payload))
		}
		//处理二进制内容
		if frame.Header.OpCode == ws.OpBinary {
			go s.handleBinary(user, frame.Payload)
		}
	}
}

// 广播消息
func (s *Server) handle(user string, message string) {
	logrus.Infof("recv message %s from %s", message, user)
	s.Lock()
	defer s.Unlock()
	broadcast := fmt.Sprintf("%s -- FROM %s", message, user)
	for u, conn := range s.users {
		if u == user { // 不发给自己
			continue
		}
		logrus.Infof("send to %s : %s", u, broadcast)
		err := s.writeText(conn, broadcast)
		if err != nil {
			logrus.Errorf("write to %s failed, error: %v", user, err)
		}
	}
}

func (s *Server) writeText(conn net.Conn, message string) error {
	// 创建文本帧数据
	f := ws.NewTextFrame([]byte(message))
	return ws.WriteFrame(conn, f)
}

type ServerStartOptions struct {
	id     string
	listen string
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	server := NewServer(opts.id, opts.listen)
	defer server.Shutdown()
	return server.Start()
}

const (
	CommandPing = 100
	CommandPong = 101
)

// 自定义业务层协议
// 消息指令 Command	消息长度 Length	消息载体 Payload
// 2bytes	4bytes	n bytes
func (s *Server) handleBinary(user string, message []byte) {
	logrus.Infof("recv message %s from %s", message, user)
	s.RLock()
	defer s.RUnlock()
	i := 0
	command := binary.BigEndian.Uint16(message[i : i+2])
	i += 2
	payloadLen := binary.BigEndian.Uint32(message[i : i+4])
	logrus.Infof("command: %v payloadLen: %v", command, payloadLen)
	if command == CommandPing {
		u := s.users[user]
		if err := wsutil.WriteServerBinary(u, []byte{0, CommandPong, 0, 0, 0, 0}); err != nil {
			logrus.Error(err)
		}
	}
}
