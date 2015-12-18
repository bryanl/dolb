package dolb

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type Config struct {
	ClusterOpsFactory func() ClusterOps
}

// API is a the load balancer API.
type API struct {
	Mux *mux.Router
}

// HandlerFunc is a handler function that returns a Response.
type HandlerFunc func(config *Config, r *http.Request) Response

// Response is a status and the body of the response.
type Response struct {
	status int
	body   interface{}
}

// MarshalJSON marshals a response as JSON. If the response is an error,
// it marshals an error message.
func (r *Response) MarshalJSON() ([]byte, error) {
	if r.status >= 400 {
		return json.Marshal(map[string]interface{}{"error": r.body})
	}

	return json.Marshal(r.body)
}

// Handler is a handler for a http request.
type Handler struct {
	f      HandlerFunc
	config *Config
}

// ServeHTTP services a http request. It calls the appropriate handler,
// logs the request, and encodes the response.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	w.Header().Set("Content-Type", "application/json")
	resp := h.f(h.config, r)
	w.WriteHeader(resp.status)

	_ = json.NewEncoder(w).Encode(&resp)

	totalTime := time.Now().Sub(now)
	log.WithFields(log.Fields{
		"method":            r.Method,
		"url":               r.URL.String(),
		"remote_addr":       r.RemoteAddr,
		"header-user_agent": r.Header.Get("User-Agent"),
		"status":            resp.status,
		"request-elapased":  totalTime / 1000000,
	}).Info("api request")
}

// New creates an instance of API.
func New() *API {
	a := &API{
		Mux: mux.NewRouter(),
	}

	config := &Config{
		ClusterOpsFactory: NewClusterOps,
	}

	a.Mux.Handle("/lb", Handler{config: config, f: LBCreateHandler}).Methods("POST")

	return a
}
