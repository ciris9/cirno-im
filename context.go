package cim

import (
	"cirno-im/wire/pkt"
	"google.golang.org/protobuf/proto"
	"sync"
)

type Session interface {
	GetChannelID() string
	GetGateID() string
	GetAccount() string
	GetZero() string
	GetIsp() string
	GetRemoteIP() string
	GetDevice() string
	GetApp() string
	GetTags() []string
}

type Context interface {
	Dispatcher
	SessionStorage
	Header() *pkt.Header
	ReadBody(val proto.Message) error
	Session() Session
	RespWithError(status pkt.Status, err error) error
	Resp(status pkt.Status, body proto.Message) error
	Dispatch(body proto.Message, recvs ...*Location) error
}

type HandlerFunc func(ctx Context)

type HandlerChain []HandlerFunc

type ContextImpl struct {
	sync.Mutex
	Dispatcher
	SessionStorage

	handlers HandlerChain
	index    int
	request  *pkt.LogicPkt
	session  Session
}
