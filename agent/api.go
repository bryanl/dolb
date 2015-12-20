package agent

import (
	"sync"

	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

// Config is configuration for the agent api.
type Config struct {
	sync.Mutex

	ClusterStatus ClusterStatus
}

// API is the http api for the agent.
type API struct {
	Mux *mux.Router
}

// New builds an instance of API.
func New(config *Config) *API {
	a := &API{
		Mux: mux.NewRouter(),
	}

	a.Mux.Handle("/", service.Handler{Config: config, F: RootHandler}).Methods("GET")

	return a
}
