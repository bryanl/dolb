package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/do"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the load balancer service.
type Config struct {
	BaseDomain          string
	ClusterOpsFactory   func() ClusterOps
	DBSession           dao.Session
	ServerURL           string
	DigitalOceanFactory func(token string, config *Config) do.DigitalOcean

	logger *logrus.Entry
}

// NewConfig creates a Config.
func NewConfig(bd, su string, sess dao.Session) *Config {
	return &Config{
		BaseDomain:        bd,
		ClusterOpsFactory: NewClusterOps,
		DBSession:         sess,
		ServerURL:         su,
		DigitalOceanFactory: func(token string, config *Config) do.DigitalOcean {
			client := do.GodoClientFactory(token)
			return do.NewLiveDigitalOcean(client, config.BaseDomain)
		},
		logger: logrus.WithFields(logrus.Fields{}),
	}
}

// SetLogger sets a logger for config.
func (c *Config) SetLogger(l *logrus.Entry) {
	c.logger = l
}

func (c *Config) GetLogger() *logrus.Entry {
	return c.logger
}

func (c *Config) IDGen() string {
	s, _ := dao.DefaultSnowflake()
	ui, _ := s.Next()

	return strconv.FormatUint(ui, 16)
}

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
	mux.Handle(service.PingPath, service.Handler{Config: config, F: PingHandler}).Methods("POST")

	return a, nil
}
