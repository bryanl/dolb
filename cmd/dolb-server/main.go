package main

import (
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/server"
	"github.com/ianschenck/envflag"
	"github.com/tylerb/graceful"
)

const (
	baseDomain = "lb.doitapp.io"
)

var (
	addr      = envflag.String("ADDR", ":8888", "listen address")
	dbURL     = envflag.String("DB_URL", "", "URL for database")
	serverURL = envflag.String("SERVER_URL", "", "URL for this service")
)

func main() {
	envflag.Parse()

	if *serverURL == "" {
		log.Fatal("SERVER_URL environment variable is required")
	}

	if *dbURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	sess, err := dao.NewSession(*dbURL)
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
			"server-url": *serverURL,
		}).Info("starting http listener")
		errChan <- httpServer.ListenAndServe()
	}()

	err = <-errChan
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.WithError(err).Panic("unexpected error")
	}

}
