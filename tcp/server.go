package tcp

import (
	cim "cirno-im"
	"cirno-im/constants"
	"cirno-im/logger"
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"net"
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
	cim.ServiceRegistration
	cim.ChannelMap
	cim.Acceptor
	cim.MessageListener
	cim.StateListener
	once    sync.Once
	options ServerOptions
	quit    *cim.Event
}

func NewServer(listen string, service cim.ServiceRegistration) cim.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		ChannelMap:          cim.NewChannels(100),
		quit:                cim.NewEvent(),
		options: ServerOptions{
			loginWait: constants.DefaultLoginWait,
			readWait:  constants.DefaultReadWait,
			writeWait: constants.DefaultWriteWait,
		},
	}
}

func (s *Server) Start() error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})
	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}

	listen, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	log.Info("started")
	for {
		rawConn, err := listen.Accept()
		if err != nil {
			if err1 := rawConn.Close(); err1 != nil {
				log.Warn(err1)
			}
			continue
		}

		go func(rawConn net.Conn) {
			conn := NewConn(rawConn)
			id, err := s.Accept(conn, s.options.loginWait)
			if err != nil {
				err1 := conn.WriteFrame(cim.OpClose, []byte(err.Error()))
				if err1 != nil {
					log.Warn(err1)
					return
				}
				err2 := conn.Close()
				if err2 != nil {
					log.Warn(err2)
					return
				}
				return
			}
			if _, ok := s.Get(id); ok {
				log.Warnf("channel %s exists", id)
				err1 := conn.WriteFrame(cim.OpClose, []byte(err.Error()))
				if err1 != nil {
					log.Warn(err1)
					return
				}
				err2 := conn.Close()
				if err2 != nil {
					log.Warn(err2)
					return
				}
				return
			}
			channel := cim.NewChannel(id, conn)
			channel.SetReadWait(s.options.readWait)
			channel.SetWriteWait(s.options.writeWait)
			s.Add(channel)
			log.Info("accept ", channel)
			if err1 := channel.ReadLoop(s.MessageListener); err1 != nil {
				log.Warn(err1)
			}
			s.Remove(channel.ID())
			if err2 := s.DisConnect(channel.ID()); err2 != nil {
				log.Warn(err2)
			}
			if err3 := channel.Close(); err3 != nil {
				log.Warn(err3)
			}
		}(rawConn)
		select {
		case <-s.quit.Done():
			return fmt.Errorf("listen exited")
		default:
		}
	}
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
	channel, ok := s.ChannelMap.Get(id)
	if !ok {
		return errors.New("channel not found")
	}
	return channel.Push(data)
}

func (s *Server) ShutDown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"id":     s.ServiceID(),
	})
	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()
		channels := s.ChannelMap.All()
		for _, channel := range channels {
			err := channel.Close()
			if err != nil {
				log.Warn(err)
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
