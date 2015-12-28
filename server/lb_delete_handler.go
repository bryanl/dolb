package server

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

func LBDeleteHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	vars := mux.Vars(r)
	lbID := vars["lb_id"]

	lb, err := config.DBSession.RetrieveLoadBalancer(lbID)
	if err != nil {
		return service.Response{Body: "not found", Status: 404}
	}

	err = config.DBSession.DeleteLoadBalancer(lbID)
	if err != nil {
		config.logger.WithError(err).Error("could not delete load balancer")
		return service.Response{Body: err, Status: 400}
	}

	godoc := config.DigitalOcean(lb.DigitalOceanToken)
	for _, a := range lb.Members {
		errList := []error{}

		err = godoc.DeleteAgent(a.DropletID)
		if err != nil {
			errList = append(errList, err)
		}

		err = godoc.DeleteDNS(a.IPID)

		if len(errList) > 0 {
			config.logger.WithFields(logrus.Fields{
				"err":      err.Error(),
				"agent-id": a.ID,
			}).Error("could not delete agent")
			return service.Response{Body: err, Status: 400}
		}
	}

	if lb.FloatingIPID > 0 {
		err = godoc.DeleteDNS(lb.FloatingIPID)
		if err != nil {
			config.logger.WithError(err).Error("could not delete load balancer ip")
			return service.Response{Body: err, Status: 400}
		}
	}

	return service.Response{Body: nil, Status: 204}
}
