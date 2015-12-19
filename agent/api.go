package agent

import (
	"sync"

	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

type Config struct {
	sync.Mutex

	Leader string
}

type API struct {
	Mux *mux.Router
}

func New(config *Config) *API {
	a := &API{
		Mux: mux.NewRouter(),
	}

	a.Mux.Handle("/", service.Handler{Config: config, F: AgentRootHandler}).Methods("GET")

	return a
}
