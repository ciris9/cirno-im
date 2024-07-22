package naming

type Naming interface {
	Find(serviceName string) ([]ServiceRegistration, error)
	Remove(serviceName, serviceID string) error
	Register(serviceRegistration ServiceRegistration) error
	Deregister(serviceID string) error
}
