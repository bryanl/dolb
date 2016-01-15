package cluster

import "github.com/bryanl/dolb/entity"

type Cluster interface {
	Bootstrap(lb *entity.LoadBalancer) error
}

type cluster struct{}

var _ Cluster = &cluster{}

func NewCluster() Cluster {
	return &cluster{}
}

func (c cluster) Bootstrap(lb *entity.LoadBalancer) error {
	return nil
}
