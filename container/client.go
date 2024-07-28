package container

import (
	"cirno-im"
	"cirno-im/logger"
	"sync"
)

type ClientMap interface {
	Add(client cim.Client)
	Remove(id string)
	Get(id string) (client cim.Client, ok bool)
	Services(kvs ...string) []cim.Service
}

type ClientsImpl struct {
	clients *sync.Map
}

func NewClients(nums int) ClientMap {
	return &ClientsImpl{
		clients: new(sync.Map),
	}
}

func (c *ClientsImpl) Add(client cim.Client) {
	if client.ServiceID() == "" {
		logger.WithFields(logger.Fields{
			"module": "ClientsImpl",
		}).Error("client id is required")
	}
	c.clients.Store(client.ServiceID(), client)
}

func (c *ClientsImpl) Remove(id string) {
	c.clients.Delete(id)
}

func (c *ClientsImpl) Get(id string) (client cim.Client, ok bool) {
	if id == "" {
		logger.WithFields(logger.Fields{
			"module": "ClientsImpl",
		}).Error("client id is required")
	}
	val, ok := c.clients.Load(id)
	if !ok {
		return nil, false
	}
	return val.(cim.Client), true
}

// 返回服务列表，传一对
func (c *ClientsImpl) Services(kvs ...string) []cim.Service {
	kvLen := len(kvs)
	if kvLen != 0 && kvLen != 2 {
		return nil
	}
	arr := make([]cim.Service, 0)
	c.clients.Range(func(key, value any) bool {
		service := value.(cim.Service)
		if kvLen > 0 && service.GetMetadata()[kvs[0]] != kvs[1] {
			return true
		}
		arr = append(arr, service)
		return true
	})
	return arr
}
