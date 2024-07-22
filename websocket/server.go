package websocket

import (
	cim "cirno-im"
	"cirno-im/constants"
	"cirno-im/logger"
	"cirno-im/naming"
	"context"
	"errors"
	"github.com/gobwas/ws"
	"github.com/segmentio/ksuid"
	"net/http"
	"sync"
	"time"
)

type defaultAcceptor struct{}

func (da *defaultAcceptor) Accept(conn cim.Conn, timeout time.Duration) (string, error) {
	return ksuid.New().String(), nil
}

type ServerOptions struct {
	loginWait time.Duration
	readWait  time.Duration
	writeWait time.Duration
}

type Server struct {
	listen string
	naming.ServiceRegistration
	cim.ChannelMap
	cim.Acceptor
	cim.MessageListener
	cim.StateListener
	once    sync.Once
	options ServerOptions
}

func NewServer(listen string, service naming.ServiceRegistration) cim.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		options: ServerOptions{
			loginWait: constants.DefaultLoginWait,
			readWait:  constants.DefaultReadWait,
			writeWait: constants.DefaultWriteWait,
		},
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	log := logger.WithFields(logger.Fields{
		"module": "ws.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	if s.StateListener == nil {
		return errors.New("StateListener is nil")
	}
	if s.ChannelMap == nil {
		s.ChannelMap = cim.NewChannels(100)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//step 1 升级
		rawConn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			resp(w, http.StatusBadRequest, err.Error())
			return
		}

		//step 2 包装con
		conn := NewConn(rawConn)
		id, err := s.Accept(conn, s.options.loginWait)
		if err != nil {
			err1 := conn.WriteFrame(cim.OpClose, []byte(err.Error()))
			if err1 != nil {
				logger.Error(err.Error())
			}
			return
		}

		//step 3
		if _, ok := s.Get(id); ok {
			log.Warnf("channel %s existed", id)
			err = conn.WriteFrame(cim.OpClose, []byte("channel is is repeated"))
			if err != nil {
				log.Warn(err)
				return
			}
			err = conn.Close()
			if err != nil {
				log.Warn(err)
				return
			}
		}

		//step 4
		channel := cim.NewChannel(id, conn)
		channel.SetWriteWait(s.options.writeWait)
		channel.SetReadWait(s.options.readWait)
		s.Add(channel)

		go func(ch cim.Channel) {
			err := ch.ReadLoop(s.MessageListener)
			if err != nil {
				logger.Error(err.Error())
			}
			s.Remove(ch.ID())
			err = s.DisConnect(ch.ID())
			if err != nil {
				log.Warn(err)
				return
			}
			err = ch.Close()
			if err != nil {
				log.Warn(err)
				return
			}
		}(channel)
	})
	log.Info("started")
	return http.ListenAndServe(s.listen, mux)
}

func resp(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	if msg != "" {
		_, _ = w.Write([]byte(msg))
	}
	logger.Warnf("response with code: %d", code)
}

func (s *Server) SetAcceptor(acceptor cim.Acceptor) {
	s.Acceptor = acceptor
}

func (s *Server) SetMessageListener(messageListener cim.MessageListener) {
	s.MessageListener = messageListener
}

func (s *Server) SetStateListener(stateListener cim.StateListener) {
	s.StateListener = stateListener
}

func (s *Server) SetReadWait(readWait time.Duration) {
	s.options.readWait = readWait
}

func (s *Server) SetChannelMap(channelMap cim.ChannelMap) {
	s.ChannelMap = channelMap
}

func (s *Server) Push(id string, data []byte) error {
	if ch, ok := s.ChannelMap.Get(id); ok {
		return ch.Push(data)
	} else {
		return errors.New("channel not found")
	}
}

func (s *Server) ShutDown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "ws.server",
		"id":     s.ServiceID(),
	})
	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()
		channels := s.ChannelMap.All()
		for index, ch := range channels {
			err := ch.Close()
			if err != nil {
				log.Warnf("index: %d close channel error:%s", index, err.Error())
			}
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
	})
	return nil
}
