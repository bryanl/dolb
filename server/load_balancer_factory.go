package server

import (
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/pkg/cluster"
	"golang.org/x/net/context"
)

type LoadBalancerFactory interface {
	Build(bootstrapConfig *BootstrapConfig) (*dao.LoadBalancer, error)
}

type LoadBalancerFactoryConfig struct {
	Context context.Context
	Session dao.Session
	Cluster cluster.Cluster
	KVS     kvs.KVS
}

type loadBalancerFactory struct {
}

var _ LoadBalancerFactory = &loadBalancerFactory{}

func NewLoadBalancerFactory(config *LoadBalancerFactoryConfig) LoadBalancerFactory {
	return &loadBalancerFactory{}
}

func (lbf *loadBalancerFactory) Build(bootstrapConfig *BootstrapConfig) (*dao.LoadBalancer, error) {
	return nil, nil
}
