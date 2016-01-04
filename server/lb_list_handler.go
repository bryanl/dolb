package server

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/service"
)

// LoadBalancersResponse is a response with load balancers.
type LoadBalancersResponse struct {
	LoadBalancers []LoadBalancerResponse `json:"load_balancers"`
}

// LoadBalancerResponse is a response with a load balancer.
type LoadBalancerResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	State      string     `json:"state"`
	Leader     string     `json:"leader"`
	Region     string     `json:"region"`
	FloatingIP FloatingIP `json:"floating_ip"`
	Host       string     `json:"host"`
}

// NewLoadBalancerFromDAO converts a dao LoadBalancer to an API LoadBalancer.
func NewLoadBalancerFromDAO(lb dao.LoadBalancer, baseDomain string) LoadBalancerResponse {
	return LoadBalancerResponse{
		ID:     lb.ID,
		Name:   lb.Name,
		State:  lb.State,
		Leader: lb.Leader,
		Region: lb.Region,
		FloatingIP: FloatingIP{
			ID:        lb.FloatingIpID,
			IPAddress: lb.FloatingIp,
		},
		Host: fmt.Sprintf("c-%s.%s", lb.Name, baseDomain),
	}
}

// FloatingIP is a floating ip.
type FloatingIP struct {
	ID        int    `json:"id,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
}

// LBListHandler is a handler for listing load balancers.
func LBListHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	lbs, err := config.DBSession.LoadLoadBalancers()
	if err != nil {
		logrus.WithError(err).Error("could not retrieve load balancers")
		return service.Response{Body: err, Status: 500}
	}

	lbr := LoadBalancersResponse{}
	lbr.LoadBalancers = make([]LoadBalancerResponse, len(lbs))
	for i, lb := range lbs {
		lbr.LoadBalancers[i] = NewLoadBalancerFromDAO(lb, config.BaseDomain)
	}

	return service.Response{Body: lbr, Status: http.StatusOK}
}
