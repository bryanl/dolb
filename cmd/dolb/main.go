package main

import (
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb"
	"github.com/bryanl/dolb/vendor/github.com/tylerb/graceful"
	"github.com/ianschenck/envflag"
)

var (
	addr = envflag.String("ADDR", ":8888", "listen address")
)

func main() {
	envflag.Parse()

	api := dolb.New()

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
