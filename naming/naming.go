package naming

import "cirno-im"

type Naming interface {
	Find(serviceName string) ([]cim.ServiceRegistration, error)
	Remove(serviceName, serviceID string) error
	Unsubscribe(serviceName string) error
	Register(serviceRegistration cim.ServiceRegistration) error
	Deregister(serviceID string) error
}
