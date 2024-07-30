package naming

import "cirno-im"

type Naming interface {
	Find(serviceName string) ([]cim.ServiceRegistration, error)
	Remove(serviceName, serviceID string) error
	Subscribe(serviceName string, callback func(services []cim.ServiceRegistration)) error
	Unsubscribe(serviceName string) error
	Register(serviceRegistration cim.ServiceRegistration) error
	Deregister(serviceID string) error
}
