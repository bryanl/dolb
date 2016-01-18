package app

import "github.com/bryanl/dolb/entity"

// Cluster manages load balancer agent clusters.
type Cluster interface {
	Bootstrap(lb *entity.LoadBalancer, bootstrapConfig *BootstrapConfig) (chan int, error)
}
