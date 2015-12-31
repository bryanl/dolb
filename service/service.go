package service

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	PingPath = "/ping"
)

// HandlerFunc is a handler function that returns a Response.
type HandlerFunc func(config interface{}, r *http.Request) Response

// Response is a status and the body of the response.
type Response struct {
	Status int
	Body   interface{}
}

// MarshalJSON marshals a response as JSON. If the response is an error,
// it marshals an error message.
func (r *Response) MarshalJSON() ([]byte, error) {
	if r.Status >= 400 {
		return json.Marshal(map[string]interface{}{"error": r.Body})
	}

	return json.Marshal(r.Body)
}

// Handler is a handler for a http request.
type Handler struct {
	F      HandlerFunc
	Config HandlerConfig
}

type HandlerConfig interface {
	SetLogger(*log.Entry)
	GetLogger() *log.Entry
	IDGen() string
}

// ServeHTTP services a http request. It calls the appropriate handler,
// logs the request, and encodes the response.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	u := h.Config.IDGen()

	logger := log.WithFields(log.Fields{
		"request-id": u,
	})
	h.Config.SetLogger(logger)

	w.Header().Set("Content-Type", "application/json")
	resp := h.F(h.Config, r)
	w.WriteHeader(resp.Status)

	_ = json.NewEncoder(w).Encode(&resp)

	totalTime := time.Now().Sub(now)
	logger.WithFields(log.Fields{
		"action":            "web-service",
		"method":            r.Method,
		"url":               r.URL.String(),
		"remote_addr":       r.RemoteAddr,
		"header-user_agent": r.Header.Get("User-Agent"),
		"status":            resp.Status,
		"request-elapased":  totalTime / 1000000,
	}).Info("api request")
}
