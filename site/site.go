package site

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/gorilla/mux"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/digitalocean"
)

type Config struct {
	DBSession dao.Session
	IDGen     func() string
}

type Site struct {
	Mux   http.Handler
	IDGen func() string
}

func New(config *Config) *Site {

	clientID := "ed07f403db0397d43d4d026275203d03a4a2de2b24bf8ca2d7b6fff8987ddd5e"
	clientSecret := "ad20064ec058d771cd61b5c15f58b8a06cf6af072a9b5d6f63e812c82b7c6518"

	gothic.CompleteUserAuth = completeUserAuth
	goth.UseProviders(
		digitalocean.New(clientID, clientSecret, "https://dolb.ngrok.io/auth/digitalocean/callback", "read write"),
	)

	router := mux.NewRouter()
	loggingRouter := loggingMiddleware(idGen, router)
	s := &Site{Mux: loggingRouter}

	// auth
	router.HandleFunc("/auth/digitalocean", beginGoth).Methods("GET")
	oauthCallBack := &OauthCallback{DBSession: config.DBSession}
	router.Handle("/auth/{provider}/callback", oauthCallBack).Methods("GET")

	homeHandler := &HomeHandler{DBSession: config.DBSession}
	router.Handle("/", loggingMiddleware(idGen, homeHandler)).Methods("GET")

	// define this last
	assetDir := "/Users/bryan/Development/go/src/github.com/bryanl/dolb/site/assets/"
	fs := loggingMiddleware(idGen, http.StripPrefix("/", http.FileServer(http.Dir(assetDir))))
	router.PathPrefix("/{_dummy:.*}").Handler(fs)

	return s
}

func idGen() string {
	s, _ := dao.DefaultSnowflake()
	ui, _ := s.Next()

	return strconv.FormatUint(ui, 16)
}

func loggingMiddleware(idGen func() string, h http.Handler) http.Handler {
	return loggingHandler{handler: h, idGen: idGen}
}

type loggingHandler struct {
	handler http.Handler
	idGen   func() string
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	u := h.idGen()

	logger := logrus.WithFields(logrus.Fields{
		"request-id": u,
	})

	loggedWriter := &loggedResponse{w: w}

	h.handler.ServeHTTP(loggedWriter, req)

	totalTime := time.Now().Sub(now)
	logger.WithFields(logrus.Fields{
		"action":            "site request",
		"method":            req.Method,
		"url":               req.URL.String(),
		"remote_addr":       req.RemoteAddr,
		"header-user_agent": req.Header.Get("User-Agent"),
		"status":            loggedWriter.status,
		"request-elapased":  totalTime / 1000000,
	}).Info("site request")

}

type loggedResponse struct {
	w      http.ResponseWriter
	status int
}

func (w *loggedResponse) Flush() {
	if wf, ok := w.w.(http.Flusher); ok {
		wf.Flush()
	}
}

func (w *loggedResponse) Header() http.Header         { return w.w.Header() }
func (w *loggedResponse) Write(d []byte) (int, error) { return w.w.Write(d) }

func (w *loggedResponse) WriteHeader(status int) {
	w.status = status
	w.w.WriteHeader(status)
}
