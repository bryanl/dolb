package dolb

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type BootstrapClusterRequest struct {
	Region  string   `json:"region"`
	SSHKeys []string `json:"ssh_keys"`
	Token   string   `json:"token"`
}

type BootstrapClusterResponse struct {
	ID         string
	MonitorURI string
}

// LBCreateHandler is a http handler for creating a load balancer.
func LBCreateHandler(config *Config, r *http.Request) Response {
	defer r.Body.Close()

	var bcr BootstrapClusterRequest
	err := json.NewDecoder(r.Body).Decode(&bcr)
	if err != nil {
		return Response{body: err, status: 422}
	}

	bc := &BootConfig{
		Region:  bcr.Region,
		SSHKeys: bcr.SSHKeys,
		Token:   bcr.Token,
	}

	co := config.ClusterOpsFactory()
	u, err := co.Boot(bc)
	if err != nil {
		log.WithError(err).Error("could not bootstrap cluster")
		return Response{body: err, status: 500}
	}

	bcResp := BootstrapClusterResponse{
		ID:         "lb-1",
		MonitorURI: u,
	}

	return Response{body: bcResp, status: http.StatusCreated}
}
