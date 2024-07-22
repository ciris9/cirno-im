package cim

import (
	"cirno-im/constants"
	"cirno-im/logger"
	"errors"
	"fmt"
	"sync"
	"time"
)

type ChannelImpl struct {
	sync.Mutex
	Conn
	id        string
	writeChan chan []byte
	once      sync.Once
	writeWait time.Duration
	readWait  time.Duration
	closed    *Event
}

func NewChannel(id string, conn Conn) Channel {
	log := logger.WithFields(logger.Fields{
		"module": "channel",
		"id":     id,
	})
	ch := &ChannelImpl{
		Mutex:     sync.Mutex{},
		Conn:      conn,
		id:        id,
		writeChan: make(chan []byte, 5),
		once:      sync.Once{},
		writeWait: constants.DefaultWriteWait,
		readWait:  constants.DefaultReadWait,
	}
	go func() {
		err := ch.writeLoop()
		if err != nil {
			log.Info(err)
		}
	}()
	return ch
}
func (ch *ChannelImpl) writeLoop() error {
	for {
		select {
		case payload, ok := <-ch.writeChan:
			if !ok {
				return errors.New("channel closed")
			}
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}
			chanLen := len(ch.writeChan)
			for i := 0; i < chanLen; i++ {
				payload = <-ch.writeChan
				err = ch.WriteFrame(OpBinary, payload)
				if err != nil {
					return err
				}
			}
			err = ch.Conn.Flush()
			if err != nil {
				return err
			}
		case <-ch.closed.Done():
			return nil
		}
	}
}

func (ch *ChannelImpl) ID() string {
	return ch.id
}

func (ch *ChannelImpl) Push(payload []byte) error {
	if ch.closed.HasFired() {
		return fmt.Errorf("channel %s has closed", ch.id)
	}
	ch.writeChan <- payload
	return nil
}

func (ch *ChannelImpl) WriteFrame(code OpCode, payload []byte) error {
	err := ch.Conn.SetWriteDeadline(time.Now().Add(ch.writeWait))
	if err != nil {
		return err
	}
	return ch.Conn.WriteFrame(code, payload)
}

func (ch *ChannelImpl) Close() error {
	ch.once.Do(func() {
		close(ch.writeChan)
		ch.closed.Fire()
	})
	return nil
}

func (ch *ChannelImpl) ReadLoop(listener MessageListener) error {
	log := logger.WithFields(logger.Fields{
		"struct": "ChannelImpl",
		"func":   "Readloop",
		"id":     ch.id,
	})
	ch.Lock()
	defer ch.Unlock()
	for {
		err := ch.SetReadDeadline(time.Now().Add(ch.readWait))
		if err != nil {
			return err
		}
		frame, err := ch.ReadFrame()
		if err != nil {
			return err
		}
		switch frame.GetOpCode() {
		case OpClose:
			return errors.New("remote side is close the connection")
		case OpPing:
			log.Trace("recv a ping and resp with a pong")
			err = ch.WriteFrame(OpPong, nil)
			if err != nil {
				return err
			}
		}
		payload := frame.GetPayload()
		if len(payload) == 0 {
			continue
		}
		go listener.Receive(ch, payload)
	}
}

func (ch *ChannelImpl) SetWriteWait(writeWait time.Duration) {
	if writeWait == 0 {
		return
	}
	ch.writeWait = writeWait
}

func (ch *ChannelImpl) SetReadWait(readWait time.Duration) {
	if readWait == 0 {
		return
	}
	ch.readWait = readWait
}
