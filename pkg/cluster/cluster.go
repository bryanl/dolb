package cluster

import "github.com/bryanl/dolb/dao"

type Cluster interface {
	Bootstrap(lb *dao.LoadBalancer) error
}

type cluster struct{}

var _ Cluster = &cluster{}

func NewCluster() Cluster {
	return &cluster{}
}

func (c cluster) Bootstrap(lb *dao.LoadBalancer) error {
	return nil
}
