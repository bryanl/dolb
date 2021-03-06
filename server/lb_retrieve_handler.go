package server

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// LBRetrieveHandler retrieves a load balancer.
func LBRetrieveHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	vars := mux.Vars(r)
	lbID := vars["lb_id"]
	daolb, err := config.DBSession.LoadLoadBalancer(lbID)
	if err != nil {
		logrus.WithError(err).Error("could not retrieve load balancer")
		return service.Response{Body: err, Status: 404}
	}

	lb := NewLoadBalancerFromDAO(*daolb, config.BaseDomain)

	agents, err := config.DBSession.LoadBalancerAgents(lbID)
	if err != nil {
		logrus.WithError(err).Error("could not retrieve load balancer agents")
		return service.Response{Body: err, Status: 404}
	}

	for _, a := range agents {
		lb.Agents = append(lb.Agents, NewAgentFromDAO(a))
	}

	return service.Response{Body: lb, Status: http.StatusOK}
}
