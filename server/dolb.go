package server

import (
	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the load balancer service.
type Config struct {
	ClusterOpsFactory func() ClusterOps

	logger *logrus.Entry
}

func (c *Config) SetLogger(l *logrus.Entry) {
	c.logger = l
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
