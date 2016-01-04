package site

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/server"
	"github.com/gorilla/mux"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/digitalocean"
)

var templates map[string]*template.Template

func init() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	includes := []string{}
	layouts := []string{}
	for _, name := range AssetNames() {
		if strings.HasPrefix(name, "templates/includes") {
			includes = append(includes, name)
		} else if strings.HasPrefix(name, "templates/layouts") {
			layouts = append(layouts, name)
		}
	}

	for _, layout := range layouts {
		key := filepath.Base(layout)
		templates[key] = template.New(key)

		files := append(includes, layout)
		logrus.WithFields(logrus.Fields{
			"template-name":  key,
			"template-files": strings.Join(files, ","),
		}).Info("preloading templates")
		for _, file := range files {
			in, err := Asset(file)
			if err != nil {
				logrus.WithField("asset-name", file).
					WithError(err).
					Fatal("could not load asset")
			}

			template.Must(templates[key].Parse(string(in)))
		}
	}
}

type Site struct {
	Mux   http.Handler
	IDGen func() string
}

func New(config *server.Config) *Site {

	clientID := "ed07f403db0397d43d4d026275203d03a4a2de2b24bf8ca2d7b6fff8987ddd5e"
	clientSecret := "ad20064ec058d771cd61b5c15f58b8a06cf6af072a9b5d6f63e812c82b7c6518"

	gothic.CompleteUserAuth = completeUserAuth
	goth.UseProviders(
		digitalocean.New(clientID, clientSecret, "https://dolb.ngrok.io/auth/digitalocean/callback", "read write"),
	)

	router := mux.NewRouter()
	loggingRouter := loggingMiddleware(config.IDGen, router)
	s := &Site{Mux: loggingRouter}

	bh := &baseHandler{config: config}

	// auth
	router.HandleFunc("/auth/digitalocean", beginGoth).Methods("GET")
	oauthCallBack := &OauthCallback{DBSession: config.DBSession}
	router.Handle("/auth/{provider}/callback", oauthCallBack).Methods("GET")

	lbNewHandler := &LBNewHandler{bh: bh}
	router.Handle("/lb/new", lbNewHandler).Methods("GET")

	lbShowHandler := &LBShowHandler{bh: bh}
	router.Handle("/lb/{lb_id}", lbShowHandler).Methods("GET")

	lbCreateHandler := &LBCreateHandler{bh: bh}
	router.Handle("/lb", lbCreateHandler).Methods("POST")

	//homeHandler := &HomeHandler{bh: bh}
	//router.Handle("/", homeHandler).Methods("GET")

	// define this last
	//assetDir := "/Users/bryan/Development/go/src/github.com/bryanl/dolb/site/assets/"
	baseAssetDir := "/Users/bryan/Development/go/src/github.com/bryanl/dolb/site"

	bowerCompDir := baseAssetDir + "/bower_components"
	bowerFs := http.StripPrefix("/bower_components", http.FileServer(http.Dir(bowerCompDir)))
	router.PathPrefix("/bower_components/{_dummy:.*}").Handler(bowerFs)

	appDir := baseAssetDir + "/app"
	fs := http.StripPrefix("/", http.FileServer(http.Dir(appDir)))
	router.PathPrefix("/{_dummy:.*}").Handler(fs)

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

func renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("The template %s does not exist.", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", data)
}
