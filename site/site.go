package site

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/server"
	"github.com/gorilla/mux"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/digitalocean"
)

type Site struct {
	Mux   http.Handler
	IDGen func() string
}

func New(config *server.Config) *Site {

	clientID := config.OauthClientID
	clientSecret := config.OauthClientSecret

	gothic.CompleteUserAuth = completeUserAuth
	goth.UseProviders(
		digitalocean.New(clientID, clientSecret, config.OauthCallback, "read write"),
	)

	router := mux.NewRouter()
	loggingRouter := loggingMiddleware(config.IDGen, router)
	s := &Site{Mux: loggingRouter}

	// auth
	router.HandleFunc("/auth/digitalocean", beginGoth).Methods("GET")
	oauthCallBack := &OauthCallback{DBSession: config.DBSession}
	router.Handle("/auth/{provider}/callback", oauthCallBack).Methods("GET")

	// TODO clean this up for prod mode
	//homeHandler := &HomeHandler{bh: bh}
	//router.Handle("/", homeHandler).Methods("GET")

	// define this last
	//assetDir := "/Users/bryan/Development/go/src/github.com/bryanl/dolb/site/assets/"
	//baseAssetDir := "/Users/bryan/Development/go/src/github.com/bryanl/dolb/site"

	//bowerCompDir := baseAssetDir + "/bower_components"
	//bowerFs := http.StripPrefix("/bower_components", http.FileServer(http.Dir(bowerCompDir)))
	//router.PathPrefix("/bower_components/{_dummy:.*}").Handler(bowerFs)

	//appDir := baseAssetDir + "/app"
	//fs := http.StripPrefix("/", http.FileServer(http.Dir(appDir)))
	//router.PathPrefix("/{_dummy:.*}").Handler(fs)

	u, _ := url.Parse("http://localhost:9000")
	rp := httputil.NewSingleHostReverseProxy(u)
	router.Handle("/{_dummy:.*}", rp)

	return s
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
