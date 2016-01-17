package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/bryanl/dolb/pkg/app"
	"github.com/bryanl/dolb/service"

	"golang.org/x/net/context"
)

// LoadBalancerService is a service for managing LoadBalancers.
type LoadBalancerService struct {
	LBFactoryFn func() app.LoadBalancerFactory
}

// NewLoadBalancerService builds a LoadBalancerService.
func NewLoadBalancerService(options ...func(*LoadBalancerService)) (*LoadBalancerService, error) {
	lbs := LoadBalancerService{}

	for _, option := range options {
		option(&lbs)
	}

	if lbs.LBFactoryFn == nil {
		return nil, errors.New("missing LBFactoryFn")
	}

	return &lbs, nil
}

// LBFactoryFn sets LBFactoryFn on a LoadBalancerService.
func LBFactoryFn(fn func() app.LoadBalancerFactory) func(*LoadBalancerService) {
	return func(lbs *LoadBalancerService) {
		lbs.LBFactoryFn = fn
	}
}

// Create creates a LoadBalancer
func (s *LoadBalancerService) Create(ctx context.Context, r *http.Request) service.Response {
	defer r.Body.Close()

	var bc app.BootstrapConfig
	err := json.NewDecoder(r.Body).Decode(&bc)
	if err != nil {
		return service.Response{Body: fmt.Errorf("could not decode json: %v", err), Status: 422}
	}

	factory := s.LBFactoryFn()
	lb, err := factory.Build(&bc)

	if err != nil {
		return service.Response{Body: fmt.Errorf("unable to build load balancer: %v", err), Status: 400}
	}

	return service.Response{Body: lb, Status: 201}
}
