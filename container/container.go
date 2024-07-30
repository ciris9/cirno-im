package container

import (
	"bytes"
	cim "cirno-im"
	"cirno-im/constants"
	"cirno-im/logger"
	"cirno-im/naming"
	"cirno-im/tcp"
	"cirno-im/wire"
	"cirno-im/wire/pkt"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
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

// Push message to server
func Push(server string, p *pkt.LogicPkt) error {
	p.AddStringMeta(wire.MetaDestServer, server)
	return c.Srv.Push(server, pkt.Marshal(p))
}

// Forward message to service
func Forward(serviceName string, packet *pkt.LogicPkt) error {
	if packet == nil {
		return errors.New("packet is nil")
	}
	if packet.Command == "" {
		return errors.New("command is empty in packet")
	}
	if packet.ChannelID == "" {
		return errors.New("ChannelId is empty in packet")
	}
	return ForwardWithSelector(serviceName, packet, c.selector)
}

// ForwardWithSelector forward data to the specified node of service which is chosen by selector
func ForwardWithSelector(serviceName string, packet *pkt.LogicPkt, selector Selector) error {
	cli, err := lookup(serviceName, &packet.Header, selector)
	if err != nil {
		return err
	}
	packet.AddStringMeta(wire.MetaDestServer, c.Srv.ServiceID())
	log.Debugf("forward message to %v with %s", cli.ServiceID(), &packet.Header)
	return cli.Send(pkt.Marshal(packet))
}

func connectToService(serviceName string) error {
	clients := NewClients(10)
	c.srvClients[serviceName] = clients
	delay := time.Second * 10
	err := c.Naming.Subscribe(serviceName, func(services []cim.ServiceRegistration) {
		for _, service := range services {
			if _, ok := clients.Get(service.ServiceID()); ok {
				continue
			}
			log.WithField("func", "connectToService").Infof("Watch a new service: %v", service)
			service.GetMetadata()[KeyServiceState] = StateYoung
			go func() {
				time.Sleep(delay)
				service.GetMetadata()[KeyServiceState] = StateAdult
			}()

			_, err := buildClient(clients, service)
			if err != nil {
				logger.Warn(err)
			}
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func buildClient(clients ClientMap, service cim.ServiceRegistration) (cim.Client, error) {
	c.Lock()
	defer c.Unlock()
	var (
		id       = service.ServiceID()
		name     = service.ServiceName()
		metadata = service.GetMetadata()
	)
	//检查连接是否已经存在
	if _, ok := clients.Get(id); ok {
		return nil, nil
	}
	//检查服务间是否使用tcp协议进行通讯
	if service.GetProtocol() != string(wire.ProtocolTCP) {
		return nil, errors.New("service is not a TCP protocol")
	}

	//构建客户端并且进行连接
	cli := tcp.NewClientWithProps(id, name, metadata, tcp.ClientOptions{
		Heartbeat: constants.DefaultHearBeat,
		ReadWait:  constants.DefaultReadWait,
		WriteWait: constants.DefaultWriteWait,
	})
	if c.dialer == nil {
		return nil, errors.New("dialer is nil")
	}
	cli.SetDialer(c.dialer)
	err := cli.Connect(service.DialURL())
	if err != nil {
		return nil, err
	}

	//读取消息
	go func(cli cim.Client) {
		err := readLoop(cli)
		if err != nil {
			log.Errorln(err)
		}
		clients.Remove(id)
		cli.Close()
	}(cli)
	clients.Add(cli)
	return cli, nil
}

func readLoop(cli cim.Client) error {
	log := logger.WithFields(logger.Fields{
		"module": "container",
		"func":   "readLoop",
	})
	log.Infof("readLoop started of %s %s", cli.ServiceID(), cli.ServiceName())
	for {
		frame, err := cli.Read()
		if err != nil {
			return err
		}
		if frame.GetOpCode() != cim.OpBinary {
			continue
		}
		buf := bytes.NewBuffer(frame.GetPayload())
		packet, err := pkt.MustReadLogicPkt(buf)
		if err != nil {
			log.Errorln(err)
			continue
		}
		err = pushMessage(packet)
		if err != nil {
			log.Errorln(err)
		}
	}
}

func pushMessage(packet *pkt.LogicPkt) error {
	server, _ := packet.GetMeta(wire.MetaDestServer)
	if server != c.Srv.ServiceID() {
		return fmt.Errorf("dest_server is not incorrect, %s != %s", server, c.Srv.ServiceID())
	}
	channels, ok := packet.GetMeta(wire.MetaDestChannels)
	if !ok {
		return fmt.Errorf("dest_channels is nil, %s", wire.MetaDestChannels)
	}

	channelIDs := strings.Split(channels.(string), ",")
	packet.DelMeta(wire.MetaDestServer)
	packet.DelMeta(wire.MetaDestChannels)
	payload := pkt.Marshal(packet)
	log.Debugf("Push to %v %v", channelIDs, packet)
	for _, channelID := range channelIDs {
		err := c.Srv.Push(channelID, payload)
		if err != nil {
			log.Errorln(err)
		}
	}
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
	clients, ok := c.srvClients[serviceName]
	if !ok {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	srvs := clients.Services(KeyServiceState, StateAdult)
	if len(srvs) == 0 {
		return nil, fmt.Errorf("no service found")
	}
	id := selector.Lookup(header, srvs)
	if cli, ok := clients.Get(id); ok {
		return cli, nil
	}
	return nil, errors.New("no client found")
}
