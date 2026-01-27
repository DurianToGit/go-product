package registry

import "context"

type ServiceInstance struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

type ServiceRegistry interface {
	Register(ctx context.Context, serviceName string, instance *ServiceInstance, ttlSeconds int64) error
	Close() error
}

type ServiceDiscovery interface {
	ListInstances(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Close() error
}
