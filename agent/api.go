package agent

import (
	"strconv"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/firewall"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the agent api.
type Config struct {
	sync.Mutex
	ClusterStatus ClusterStatus

	AgentID           string
	Context           context.Context
	ClusterName       string
	ClusterID         string
	DigitalOceanToken string
	DropletID         string
	Firewall          firewall.Firewall
	KVS               KVS
	Name              string
	Region            string
	ServerURL         string

	logger *logrus.Entry
}

// SetLogger sets the logger for Config.
func (c *Config) SetLogger(l *logrus.Entry) {
	c.logger = l
}

func (c *Config) IDGen() string {
	s, _ := dao.DefaultSnowflake()
	ui, _ := s.Next()

	return strconv.FormatUint(ui, 16)
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
