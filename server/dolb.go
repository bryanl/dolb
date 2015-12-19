package server

import (
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the load balancer service.
type Config struct {
	ClusterOpsFactory func() ClusterOps
}

// API is a the load balancer API.
type API struct {
	Mux *mux.Router
}

// New creates an instance of API.
func New() *API {
	a := &API{
		Mux: mux.NewRouter(),
	}

	config := &Config{
		ClusterOpsFactory: NewClusterOps,
	}

	a.Mux.Handle("/lb", service.Handler{Config: config, F: LBCreateHandler}).Methods("POST")

	return a
}
