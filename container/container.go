package container

import (
	cim "cirno-im"
	"cirno-im/logger"
	"cirno-im/naming"
	"cirno-im/wire"
	"cirno-im/wire/pkt"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	stateUninitialized = iota
	stateInitalized
	stateStart
	stateClosed
)

const (
	StateYoung = "young"
	StateAdult = "adult"
)

const (
	KeyServiceState = "service_state"
)

type Container struct {
	sync.RWMutex
	Naming naming.Naming
	Srv    cim.Server

	state      uint32
	srvClients map[string]ClientMap
	selector   Selector
	dialer     cim.Dialer
	deps       map[string]struct{}
}

var log = logger.WithField("module", "container")

var c = &Container{
	state:    0,
	selector: &HashSelector{},
	deps:     map[string]struct{}{},
}

func Default() *Container {
	return c
}

func Init(srv cim.Server, deps ...string) error {
	if !atomic.CompareAndSwapUint32(&c.state, stateUninitialized, stateInitalized) {
		return errors.New("container already inited")
	}

	c.Srv = srv
	for _, dep := range deps {
		if _, ok := c.deps[dep]; !ok {
			continue
		}
		c.deps[dep] = struct{}{}
	}
	log.WithField("func", "Init").Infof("srv %s:%s - deps %v", srv.ServiceID(), srv.ServiceName(), c.deps)
	c.srvClients = make(map[string]ClientMap, len(deps))
	return nil
}

func SetDialer(dialer cim.Dialer) {
	c.dialer = dialer
}

func EnableMonitor(listen string) error {
	return nil
}

func SetSelector(selector Selector) {
	c.selector = selector
}

func SetServiceName(name naming.Naming) {
	c.Naming = name
}

func Start() error {
	if c.Naming == nil {
		return errors.New("naming is nil")
	}
	if !atomic.CompareAndSwapUint32(&c.state, stateInitalized, stateStart) {
		return errors.New("container already started")
	}

	// 1. 启动server
	go func(srv cim.Server) {
		err := srv.Start()
		if err != nil {
			log.Errorln(err.Error())
		}
	}(c.Srv)

	// 2. 与依赖的服务建立连接
	for service := range c.deps {
		go func(service string) {
			err := connectToService(service)
			if err != nil {
				log.Errorln(err.Error())
			}
		}(service)
	}

	//3.服务注册
	if c.Srv.PublicAddress() != "" && c.Srv.PublicPort() != 0 {
		err := c.Naming.Register(c.Srv)
		if err != nil {
			log.Errorln(err.Error())
		}
	}

	// 等待中断信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Infoln("exit signal:", <-ch)
	return shutdown()
}

func Push(serviceName string, packet *pkt.LogicPkt) error {
	if packet == nil {
		return errors.New("packt is nil")
	}
	if packet.Command == "" {
		return errors.New("command is nil")
	}
	if packet.ChannelID == "" {
		return errors.New("channelID is nil")
	}
	return ForwardWithSelector(serviceName, packet, c.selector)
}

func ForwardWithSelector(serviceName string, packet *pkt.LogicPkt, selector Selector) error {
	cli, err := lookup(serviceName, &packet.Header, selector)
	if err != nil {
		return err
	}
	packet.AddStringMeta(wire.MetaDestServer, c.Srv.ServiceID())
	log.Debugf("forward message to %v with %s", cli.ServiceID(), &packet.Header)
	return cli.Send(pkt.Marshal(packet))
}

func connectToService(service string) error {
	return nil
}

func shutdown() error {
	if !atomic.CompareAndSwapUint32(&c.state, stateStart, stateClosed) {
		return errors.New("container already shutdown")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	//close server elegantly
	err := c.Srv.ShutDown(ctx)
	if err != nil {
		return err
	}

	// deregiste service from service registrer center
	err = c.Naming.Deregister(c.Srv.ServiceID())
	if err != nil {
		return err
	}
	// unsubscribe service change
	for dep := range c.deps {
		_ = c.Naming.Unsubscribe(dep)
	}

	log.Infoln("shutdown")
	return nil
}

func lookup(serviceName string, header *pkt.Header, selector Selector) (cim.Client, error) {
	clients,ok:=c.srvClients[serviceName]
	if !ok{
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	srvs:=clients.Services(KeyServiceState,StateAdult)
	if len(srvs) == 0 {
		return nil, fmt.Errorf("no service found")
	}
	id:=selector.Lookup(header,srvs)
	if cli,ok:=clients.Get(id);ok{
		return cli, nil
	}
	return nil,errors.New("no client found")
}