package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bryanl/dolb/service"

	"golang.org/x/net/context"
)

type LoadBalancerService struct {
	loadBalancerFactory LoadBalancerFactory
}

func (s *LoadBalancerService) Create(ctx context.Context, r *http.Request) service.Response {
	defer r.Body.Close()

	var bc BootstrapConfig
	err := json.NewDecoder(r.Body).Decode(&bc)
	if err != nil {
		return service.Response{Body: fmt.Errorf("could not decode json: %v", err), Status: 422}
	}

	factory := s.loadBalancerFactory
	lb, err := factory.Build(&bc)

	if err != nil {
		return service.Response{Body: fmt.Errorf("unable to build load balancer: %v", err), Status: 400}
	}

	return service.Response{Body: lb, Status: 201}
}
