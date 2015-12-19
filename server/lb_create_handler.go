package server

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
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
		return service.Response{Body: err, Status: 422}
	}

	co := config.ClusterOpsFactory()
	u, err := co.Bootstrap(&bc)
	if err != nil {
		log.WithError(err).Error("could not bootstrap cluster")
		return service.Response{Body: err, Status: 400}
	}

	bcResp := BootstrapClusterResponse{
		ID:         bc.Name,
		MonitorURI: u,
	}

	return service.Response{Body: bcResp, Status: http.StatusCreated}
}
