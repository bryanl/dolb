package main

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/server"
	"github.com/bryanl/dolb/site"
	"github.com/gorilla/mux"
	"github.com/ianschenck/envflag"
	"github.com/tylerb/graceful"
)

const (
	baseDomain = "lb.doitapp.io"
)

var (
	addr              = envflag.String("ADDR", ":8888", "listen address")
	dsn               = envflag.String("DB_URL", "", "URL for database")
	serverURL         = envflag.String("SERVER_URL", "", "URL for this service")
	oauthClientID     = envflag.String("OAUTH_CLIENT_ID", "", "oauth client id")
	oauthClientSecret = envflag.String("OAUTH_CLIENT_SECRET", "", "oauth client secret")
	oauthCallback     = envflag.String("OAUTH_CALLBACK_URL", "", "oauth callback URL")
	etcdEndpoints     = envflag.String("ETCDENDPOINTS", "", "comma separted list of ectd endpoints")
	etcdCAPemFile     = envflag.String("ETCD_CA_PEM", "", "etcd ca pem")
	etcdClientKeyFile = envflag.String("ETCD_CLIENT_KEY", "", "etcd ca key")
	etcdClientPemFile = envflag.String("ETCD_CLIENT_PEM", "", "etcd ca pem")
)

func main() {
	envflag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	if *serverURL == "" {
		log.Fatal("SERVER_URL environment variable is required")
	}

	if *dsn == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	if *oauthClientID == "" || *oauthClientSecret == "" || *oauthCallback == "" {
		log.Fatal("OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, and OAUTH_CALLBACK_URL environment variables are required")
	}

	sess, err := dao.NewSession(*dsn)
	if err != nil {
		log.WithError(err).Fatal("could not create database connection")
	}

	c := server.NewConfig(baseDomain, *serverURL, sess)
	serverAPI, err := server.New(c)
	if err != nil {
		log.WithError(err).Fatal("could not create Api")
	}

	c.OauthClientID = *oauthClientID
	c.OauthClientSecret = *oauthClientSecret
	c.OauthCallback = *oauthCallback

	kv, err := initKVS(c)
	if err != nil {
		log.WithError(err).Fatal("could not initialize kvs")
	}

	c.KVS = kv

	lbStatus := server.NewLBStatus(c)
	go lbStatus.Track()

	dolbSite := site.New(c)

	// FIXME everything above this might be a huge screwup

	em, err := initEntityManager()
	if err != nil {
		log.WithError(err).Fatal("could not initialize entity manager")
	}

	newLBS := server.NewLoadBalancerService(kv, em)

	rootMux := mux.NewRouter()
	rootMux.Handle("/api/", serverAPI.Mux)
	rootMux.Handle("/api/{_dummy:.*}", serverAPI.Mux)
	rootMux.Handle("/api2/", newLBS.Mux)
	rootMux.Handle("/api2/{_dummy:.*}", newLBS.Mux)
	rootMux.Handle("/", dolbSite.Mux)
	rootMux.Handle("/{_dummy:.*}", dolbSite.Mux)

	errChan := make(chan error)
	go func() {
		httpServer := graceful.Server{
			Timeout: 10 * time.Second,
			Server: &http.Server{
				Addr:    *addr,
				Handler: rootMux,
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

func initKVS(c *server.Config) (kvs.KVS, error) {
	tlsConfig := &kvs.TLSConfig{
		RootPEM:     *etcdCAPemFile,
		Certificate: *etcdClientPemFile,
		Key:         *etcdClientKeyFile,
	}
	kapi, err := kvs.NewKeysAPI(*etcdEndpoints, tlsConfig)
	if err != nil {
		return nil, err
	}

	return kvs.NewEtcd(c.Context, kapi), nil
}

func initEntityManager() (entity.Manager, error) {
	connection, err := entity.NewConnection(*dsn)
	if err != nil {
		return nil, err
	}

	return entity.NewManager(connection), nil
}
