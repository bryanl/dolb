package server

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/do"
	"github.com/bryanl/dolb/dolbutil"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/pkg/app"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the load balancer service.
type Config struct {
	BaseDomain          string
	ClusterOpsFactory   func() ClusterOps
	Context             context.Context
	DBSession           dao.Session
	KVS                 kvs.KVS
	ServerURL           string
	DigitalOceanFactory func(token string, config *Config) do.DigitalOcean
	OauthClientID       string
	OauthClientSecret   string
	OauthCallback       string
	LBUpdateChan        chan *dao.LoadBalancer

	logger *logrus.Entry
}

// NewConfig creates a Config.
func NewConfig(bd, su string, sess dao.Session) *Config {
	return &Config{
		BaseDomain:        bd,
		ClusterOpsFactory: NewClusterOps,
		Context:           context.Background(),
		DBSession:         sess,
		ServerURL:         su,
		DigitalOceanFactory: func(token string, config *Config) do.DigitalOcean {
			client := do.GodoClientFactory(token)
			return do.NewLiveDigitalOcean(client, config.BaseDomain)
		},
		LBUpdateChan: make(chan *dao.LoadBalancer, 10),
		logger:       app.DefaultLogger(),
	}
}

// SetLogger sets a logger for config.
func (c *Config) SetLogger(l *logrus.Entry) {
	c.logger = l
}

// GetLogger returns the config's logger.
func (c *Config) GetLogger() *logrus.Entry {
	return c.logger
}

// IDGen returns a new random id.
func (c *Config) IDGen() string {
	id := dolbutil.GenerateRandomID()
	return dolbutil.TruncateID(id)
}

// DigitalOcean returns a new instance of do.DigitalOcean.
func (c *Config) DigitalOcean(token string) do.DigitalOcean {
	return c.DigitalOceanFactory(token, c)
}

// API is a the load balancer API.
type API struct {
	Mux http.Handler
}

// New creates an instance of API.
func New(config *Config) (*API, error) {
	mux := mux.NewRouter()

	a := &API{
		Mux: mux,
	}

	if config.ServerURL == "" {
		return nil, errors.New("missing ServerURL")
	}

	mux.Handle("/api/lb", service.Handler{Config: config, F: LBListHandler}).Methods("GET")
	mux.Handle("/api/lb", service.Handler{Config: config, F: LBCreateHandler}).Methods("POST")
	mux.Handle("/api/lb/{lb_id}", service.Handler{Config: config, F: LBRetrieveHandler}).Methods("GET")
	mux.Handle("/api/lb/{lb_id}", service.Handler{Config: config, F: LBDeleteHandler}).Methods("DELETE")
	mux.Handle("/api/user", service.Handler{Config: config, F: UserRetrieveHandler}).Methods("GET")
	mux.Handle(service.PingPath, service.Handler{Config: config, F: PingHandler}).Methods("POST")
	mux.Handle("/api/lb/{lb_id}/services", service.Handler{Config: config, F: ServiceCreateHandler}).Methods("POST")
	mux.Handle("/api/lb/{lb_id}/services", service.Handler{Config: config, F: ServiceListHandler}).Methods("GET")

	return a, nil
}
