package server

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
)

// LoadBalancersResponse is a response with load balancers.
type LoadBalancersResponse struct {
	LoadBalancers []LoadBalancer `json:"load_balancers"`
}

// LoadBalancer is a load balancer.
type LoadBalancer struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Leader     string     `json:"leader"`
	Region     string     `json:"region"`
	FloatingIP FloatingIP `json:"floating_ip"`
}

// FloatingIP is a floating ip.
type FloatingIP struct {
	ID        int    `json:"id"`
	IPAddress string `json:"ip_address"`
}

// LBListHandler is a handler for listing load balancers.
func LBListHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	dbLBS, err := config.DBSession.ListLoadBalancers()
	if err != nil {
		logrus.WithError(err).Error("could not retrieve load balancers")
		return service.Response{Body: err, Status: 500}
	}

	lbr := LoadBalancersResponse{}
	lbr.LoadBalancers = make([]LoadBalancer, len(dbLBS))
	for i, lb := range dbLBS {
		lbr.LoadBalancers[i] = LoadBalancer{
			ID:     lb.ID,
			Name:   lb.Name,
			Leader: lb.LeaderString(),
			Region: lb.Region,
			FloatingIP: FloatingIP{
				ID:        lb.FloatingIPID,
				IPAddress: lb.FloatingIP,
			},
		}
	}

	return service.Response{Body: lbr, Status: http.StatusOK}
}
