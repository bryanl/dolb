package server

import (
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the load balancer service.
type Config struct {
	ClusterOpsFactory func() ClusterOps
	ServerURL         string

	logger *logrus.Entry
}

// NewConfig creates a Config.
func NewConfig(su string) *Config {
	return &Config{
		ClusterOpsFactory: NewClusterOps,
		ServerURL:         su,
	}
}

func (c *Config) SetLogger(l *logrus.Entry) {
	c.logger = l
}

// API is a the load balancer API.
type API struct {
	Mux *mux.Router
}

// New creates an instance of API.
func New(config *Config) (*API, error) {
	a := &API{
		Mux: mux.NewRouter(),
	}

	if config.ServerURL == "" {
		return nil, errors.New("missing ServerURL")
	}

	a.Mux.Handle("/lb", service.Handler{Config: config, F: LBCreateHandler}).Methods("POST")
	a.Mux.Handle("/register", service.Handler{Config: config, F: RegisterHandler}).Methods("POST")

	return a, nil
}
