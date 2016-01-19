package service

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dolbutil"
)

var (
	PingPath = "/api/ping"
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

		body := r.Body
		if err, ok := r.Body.(error); ok {
			body = err.Error()
		}

		return json.Marshal(map[string]interface{}{"error": body})
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

	if os.Getenv("QUIET_LOG") == "1" && resp.Status <= 400 {
		return
	}

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

type Handler2 struct {
	F   HandlerFunc2
	Ctx context.Context
}

type HandlerFunc2 func(ctx context.Context, r *http.Request) Response

// ServeHTTP services a http request. It calls the appropriate handler,
// logs the request, and encodes the response.
func (h Handler2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	u := dolbutil.GenerateRandomID()

	logger := log.WithFields(log.Fields{
		"request-id": u,
	})

	w.Header().Set("Content-Type", "application/json")
	resp := h.F(h.Ctx, r)
	w.WriteHeader(resp.Status)

	_ = json.NewEncoder(w).Encode(&resp)

	totalTime := time.Now().Sub(now)

	if os.Getenv("QUIET_LOG") == "1" && resp.Status <= 400 {
		return
	}

	logger.WithFields(log.Fields{
		"action":            "web-service",
		"method":            r.Method,
		"url":               r.URL.String(),
		"remote_addr":       r.RemoteAddr,
		"header-user_agent": r.Header.Get("User-Agent"),
		"status":            resp.Status,
		"request-elapased":  totalTime / 1000000,
	}).Info("api2 request")
}
