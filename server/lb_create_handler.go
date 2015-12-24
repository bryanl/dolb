package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
)

// BootstrapClusterResponse is a bootstrap cluster response.
type BootstrapClusterResponse struct {
	ID         string
	MonitorURI string
}

// LBCreateHandler is a http handler for creating a load balancer.
func LBCreateHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)
	defer r.Body.Close()

	var bc BootstrapConfig
	err := json.NewDecoder(r.Body).Decode(&bc)
	if err != nil {
		return service.Response{Body: fmt.Errorf("could not decode json: %v", err), Status: 422}
	}

	if bc.DigitalOceanToken == "" {
		return service.Response{Body: "digitalocean_token is required", Status: 400}
	}

	co := config.ClusterOpsFactory()
	u, err := co.Bootstrap(&bc, config)
	if err != nil {
		config.logger.WithError(err).Error("could not bootstrap cluster")
		return service.Response{Body: err, Status: 400}
	}

	config.logger.WithFields(logrus.Fields{
		"cluster-name":   bc.Name,
		"cluster-region": bc.Region,
	}).Info("created load balancer")

	bcResp := BootstrapClusterResponse{
		ID:         bc.Name,
		MonitorURI: u,
	}

	return service.Response{Body: bcResp, Status: http.StatusCreated}
}
