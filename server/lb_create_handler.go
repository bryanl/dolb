package server

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// BootstrapClusterResponse is a bootstrap cluster response.
type BootstrapClusterResponse struct {
	ID         string
	MonitorURI string
}

// LBCreateHandler is a http handler for creating a load balancer.
func LBCreateHandler(config *Config, r *http.Request) Response {
	defer r.Body.Close()

	var bc BootstrapConfig
	err := json.NewDecoder(r.Body).Decode(&bc)
	if err != nil {
		return Response{body: err, status: 422}
	}

	co := config.ClusterOpsFactory()
	u, err := co.Bootstrap(&bc)
	if err != nil {
		log.WithError(err).Error("could not bootstrap cluster")
		return Response{body: err, status: 400}
	}

	bcResp := BootstrapClusterResponse{
		ID:         bc.Name,
		MonitorURI: u,
	}

	return Response{body: bcResp, status: http.StatusCreated}
}
