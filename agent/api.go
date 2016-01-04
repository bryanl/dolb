package agent

import (
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dolbutil"
	"github.com/bryanl/dolb/firewall"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the agent api.
type Config struct {
	sync.Mutex
	ClusterStatus ClusterStatus

	AgentID               string
	Context               context.Context
	ClusterName           string
	ClusterID             string
	DigitalOceanToken     string
	DropletID             string
	Firewall              firewall.Firewall
	KVS                   kvs.KVS
	Name                  string
	Region                string
	ServerURL             string
	ServiceManagerFactory ServiceManagerFactory

	logger *logrus.Entry
}

// SetLogger sets the logger for Config.
func (c *Config) SetLogger(l *logrus.Entry) {
	c.logger = l
}

func (c *Config) GetLogger() *logrus.Entry {
	return c.logger
}

func (c *Config) IDGen() string {
	id := dolbutil.GenerateRandomID()
	return dolbutil.TruncateID(id)
}

// API is the http api for the agent.
type API struct {
	Mux *mux.Router
}

// NewAPI builds an instance of API.
func NewAPI(config *Config) *API {
	a := &API{
		Mux: mux.NewRouter(),
	}

	a.Mux.Handle("/", service.Handler{Config: config, F: RootHandler}).Methods("GET")
	a.Mux.Handle("/services", service.Handler{Config: config, F: ServiceCreateHandler}).Methods("POST")
	a.Mux.Handle("/services", service.Handler{Config: config, F: ServiceListHandler}).Methods("GET")
	a.Mux.Handle("/services/{service}", service.Handler{Config: config, F: ServiceRetrieveHandler}).Methods("GET")
	a.Mux.Handle("/services/{service}", service.Handler{Config: config, F: ServiceDeleteHandler}).Methods("DELETE")
	a.Mux.Handle("/services/{service}/upstreams", service.Handler{Config: config, F: UpstreamCreateHandler}).Methods("PUT")
	a.Mux.Handle("/services/{service}/upstreams/{upstream}", service.Handler{Config: config, F: UpstreamDeleteHandler}).Methods("DELETE")

	return a
}

func convertServiceToResponse(s kvs.Service) service.ServiceResponse {
	sr := service.ServiceResponse{
		Name:      s.Name(),
		Port:      s.Port(),
		Type:      s.Type(),
		Config:    map[string]interface{}{},
		Upstreams: []service.UpstreamResponse{},
	}

	for k, v := range s.ServiceConfig() {
		sr.Config[k] = v
	}

	for _, u := range s.Upstreams() {
		ur := service.UpstreamResponse{
			ID:   u.ID,
			Host: u.Host,
			Port: u.Port,
		}
		sr.Upstreams = append(sr.Upstreams, ur)
	}

	return sr

}
