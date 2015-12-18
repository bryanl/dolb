package main

import (
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/server"
	"github.com/ianschenck/envflag"
	"github.com/tylerb/graceful"
)

var (
	addr = envflag.String("ADDR", ":8888", "listen address")
)

func main() {
	envflag.Parse()

	api := server.New()

	errChan := make(chan error)
	go func() {
		httpServer := graceful.Server{
			Timeout: 10 * time.Second,
			Server: &http.Server{
				Addr:    *addr,
				Handler: api.Mux,
			},
		}

		log.WithField("addr", *addr).Info("starting http listener")
		errChan <- httpServer.ListenAndServe()
	}()

	err := <-errChan
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.WithError(err).Panic("unexpected error")
	}

}
