package server

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/do"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// LBDeleteHandler deletes a load balancer.
func LBDeleteHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	vars := mux.Vars(r)
	lbID := vars["lb_id"]

	lb, err := config.DBSession.LoadLoadBalancer(lbID)
	if err != nil {
		return service.Response{Body: "not found", Status: 404}
	}

	godoc := config.DigitalOcean(lb.DigitaloceanAccessToken)
	agents, err := config.DBSession.LoadBalancerAgents(lbID)
	if err != nil {
		config.logger.WithError(err).Error("could not load agents")
	}

	for _, a := range agents {
		err = deleteAgent(&a, godoc)
		if err != nil {
			config.logger.WithFields(logrus.Fields{
				"err":      err.Error(),
				"agent-id": a.ID,
			}).Error("could not delete agent")
			return service.Response{Body: err, Status: 500}
		}

	}

	err = deleteLB(lb, godoc, config.logger)
	if err != nil {
		config.logger.WithFields(logrus.Fields{
			"err":              err.Error(),
			"load-balancer-id": lb.ID,
		}).Error("could not delete load balancer")
		return service.Response{Body: err, Status: 500}
	}

	err = config.KVS.Delete("/dolb/clusters/" + lb.ID)
	if err != nil {
		config.logger.
			WithError(err).
			WithField("cluster-id", lb.ID).
			Error("could not delete kvs entry for cluster")
	}

	return service.Response{Body: nil, Status: 204}
}

func deleteAgent(a *dao.Agent, godoc do.DigitalOcean) error {
	err := godoc.DeleteAgent(a.DropletID)
	if err != nil {
		return err
	}

	err = godoc.DeleteDNS(a.IpID)
	if err != nil {
		return err
	}

	err = a.Delete()
	if err != nil {
		return err
	}

	return nil
}

func deleteLB(lb *dao.LoadBalancer, godoc do.DigitalOcean, logger *logrus.Entry) error {
	var err error

	if lb.FloatingIpID > 0 {
		lb.FloatingIpID = 0
		err = godoc.DeleteDNS(lb.FloatingIpID)
		if err != nil {
			logger.WithError(err).Error("could not delete load balancer ip")
			return err
		}
	}

	lb.FloatingIp = ""
	lb.Leader = ""

	err = lb.Delete()
	if err != nil {
		logger.WithError(err).Error("could not delete load balancer")
		return err
	}

	return nil
}
