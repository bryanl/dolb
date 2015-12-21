package agent

import (
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
	etcdclient "github.com/coreos/etcd/client"
	"github.com/gorilla/mux"
)

// Config is configuration for the agent api.
type Config struct {
	sync.Mutex
	ClusterStatus ClusterStatus

	Context           context.Context
	DigitalOceanToken string
	DropletID         string
	KeysAPI           etcdclient.KeysAPI
	Name              string
	Region            string

	logger *logrus.Entry
}

func (c *Config) SetLogger(l *logrus.Entry) {
	c.logger = l
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

	return a
}
