package cluster

import (
	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/pkg/app"
)

// Cluster managements load balancer agent clusters.
type Cluster struct{}

var _ app.Cluster = &Cluster{}

// New builds a cluster.
func New() app.Cluster {
	return &Cluster{}
}

// Bootstrap bootstraps an agent cluster.
func (c Cluster) Bootstrap(lb *entity.LoadBalancer) error {
	return nil
}
