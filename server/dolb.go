package server

import (
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/doa"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the load balancer service.
type Config struct {
	BaseDomain        string
	ClusterOpsFactory func() ClusterOps
	DBSession         doa.Session
	ServerURL         string

	logger *logrus.Entry
}

// NewConfig creates a Config.
func NewConfig(bd, su string, sess doa.Session) *Config {
	return &Config{
		BaseDomain:        bd,
		ClusterOpsFactory: NewClusterOps,
		DBSession:         sess,
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
