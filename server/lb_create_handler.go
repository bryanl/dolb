package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/service"
)

// BootstrapClusterResponse is a bootstrap cluster response.
type BootstrapClusterResponse struct {
	LoadBalancer LoadBalancerResponse
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

	lb, err := CreateLoadBalancer(bc, config)
	if err != nil {
		return service.Response{Body: err, Status: 400}
	}

	bcResp := BootstrapClusterResponse{
		LoadBalancer: NewLoadBalancerFromDAO(*lb, config.BaseDomain),
	}

	return service.Response{Body: bcResp, Status: http.StatusCreated}
}

func CreateLoadBalancer(bc BootstrapConfig, config *Config) (*dao.LoadBalancer, error) {
	if bc.DigitalOceanToken == "" {
		return nil, fmt.Errorf("DigitalOcean token is required")
	}

	lb := config.DBSession.NewLoadBalancer()
	lb.Name = bc.Name
	lb.Region = bc.Region
	lb.DigitaloceanAccessToken = bc.DigitalOceanToken
	err := lb.Save()
	if err != nil {
		return nil, err
	}

	co := config.ClusterOpsFactory()
	bo := &BootstrapOptions{
		Config:          config,
		LoadBalancer:    lb,
		BootstrapConfig: &bc,
	}

	err = co.Bootstrap(bo)
	if err != nil {
		config.GetLogger().WithError(err).Error("could not bootstrap cluster")
		return nil, err
	}

	config.GetLogger().WithFields(logrus.Fields{
		"cluster-name":   bc.Name,
		"cluster-region": bc.Region,
	}).Info("created load balancer")

	return lb, nil
}
