package server

import (
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/do"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the load balancer service.
type Config struct {
	BaseDomain        string
	ClusterOpsFactory func() ClusterOps
	DBSession         dao.Session
	ServerURL         string
	GodoClientFactory do.GodoClientFactoryFn

	logger *logrus.Entry
}

// NewConfig creates a Config.
func NewConfig(bd, su string, sess dao.Session) *Config {
	return &Config{
		BaseDomain:        bd,
		ClusterOpsFactory: NewClusterOps,
		DBSession:         sess,
		ServerURL:         su,
		GodoClientFactory: do.GodoClientFactory,
	}
}

// SetLogger sets a logger for config.
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

	a.Mux.Handle("/lb", service.Handler{Config: config, F: LBListHandler}).Methods("GET")
	a.Mux.Handle("/lb", service.Handler{Config: config, F: LBCreateHandler}).Methods("POST")
	a.Mux.Handle("/lb/{lb_id}", service.Handler{Config: config, F: LBRetrieveHandler}).Methods("GET")
	a.Mux.Handle(service.PingPath, service.Handler{Config: config, F: PingHandler}).Methods("POST")

	return a, nil
}
