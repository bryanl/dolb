package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/pkg/app"
	"github.com/bryanl/dolb/pkg/lbfactory"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"

	"golang.org/x/net/context"
)

// LoadBalancerService is a service for managing LoadBalancers.
type LoadBalancerService struct {
	Context       context.Context
	KVS           kvs.KVS
	EntityManager entity.Manager
	LBFactoryFn   func(kvs.KVS, entity.Manager) app.LoadBalancerFactory
	Mux           *mux.Router
}

func defaultLBFactoryFn(kv kvs.KVS, em entity.Manager) app.LoadBalancerFactory {
	return lbfactory.New(kv, em)
}

// NewLoadBalancerService builds a LoadBalancerService.
func NewLoadBalancerService(kv kvs.KVS, em entity.Manager, options ...func(*LoadBalancerService)) *LoadBalancerService {
	lbs := LoadBalancerService{
		Mux:           mux.NewRouter(),
		KVS:           kv,
		EntityManager: em,
	}

	for _, option := range options {
		option(&lbs)
	}

	if lbs.LBFactoryFn == nil {
		lbs.LBFactoryFn = defaultLBFactoryFn
	}

	if lbs.Context == nil {
		lbs.Context = context.Background()
	}

	lbs.handle("/api2/lb", lbs.Create, "POST")

	return &lbs
}

// LBFactoryFn sets LBFactoryFn on a LoadBalancerService.
func LBFactoryFn(fn func(kvs.KVS, entity.Manager) app.LoadBalancerFactory) func(*LoadBalancerService) {
	return func(lbs *LoadBalancerService) {
		lbs.LBFactoryFn = fn
	}
}

// Context configures LoadBalancerService Context.
func Context(ctx context.Context) func(*LoadBalancerService) {
	return func(lbs *LoadBalancerService) {
		lbs.Context = ctx
	}
}

func (s *LoadBalancerService) handle(path string, fn service.HandlerFunc2, methods ...string) {
	s.Mux.Handle(path, service.Handler2{F: fn, Ctx: s.Context}).Methods(methods...)
}

// Create creates a LoadBalancer
func (s *LoadBalancerService) Create(ctx context.Context, r *http.Request) service.Response {
	defer r.Body.Close()

	var bc app.BootstrapConfig
	err := json.NewDecoder(r.Body).Decode(&bc)
	if err != nil {
		return service.Response{Body: fmt.Errorf("could not decode json: %v", err), Status: 422}
	}

	factory := s.LBFactoryFn(s.KVS, s.EntityManager)
	lb, err := factory.Build(&bc)

	if err != nil {
		return service.Response{Body: fmt.Errorf("unable to build load balancer: %v", err), Status: 400}
	}

	return service.Response{Body: lb, Status: 201}
}
