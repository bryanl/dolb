package main

import (
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/doa"
	"github.com/bryanl/dolb/server"
	"github.com/ianschenck/envflag"
	"github.com/tylerb/graceful"
)

const (
	baseDomain = "lb.doitapp.io"
)

var (
	addr      = envflag.String("ADDR", ":8888", "listen address")
	serverURL = envflag.String("SERVER_URL", "", "URL for this service")
	dbAddr    = envflag.String("DB_ADDR", "", "database address")
	dbName    = envflag.String("DB_NAME", "dolb", "database name")
)

func main() {
	envflag.Parse()

	if *serverURL == "" {
		log.Fatal("SERVER_URL environment variable is required")
	}

	if *dbAddr == "" {
		log.Fatal("DB_ADDR environment variable is required")
	}

	sess, err := doa.NewSession(*dbAddr, *dbName)
	if err != nil {
		log.WithError(err).Fatal("could not create database connection")
	}

	c := server.NewConfig(baseDomain, *serverURL, sess)
	api, err := server.New(c)
	if err != nil {
		log.WithError(err).Fatal("could not create Api")
	}

	errChan := make(chan error)
	go func() {
		httpServer := graceful.Server{
			Timeout: 10 * time.Second,
			Server: &http.Server{
				Addr:    *addr,
				Handler: api.Mux,
			},
		}

		log.WithFields(log.Fields{
			"addr":       *addr,
			"db-addr":    *dbAddr,
			"db-name":    *dbName,
			"server-url": *serverURL,
		}).Info("starting http listener")
		errChan <- httpServer.ListenAndServe()
	}()

	err = <-errChan
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.WithError(err).Panic("unexpected error")
	}

}
