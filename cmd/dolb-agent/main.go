package main

import (
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/agent"
	etcdclient "github.com/coreos/etcd/client"
	"github.com/ianschenck/envflag"
	"github.com/tylerb/graceful"
)

var (
	addr          = envflag.String("ADDR", ":8889", "listen address")
	agentName     = envflag.String("AGENT_NAME", "", "name for agent")
	etcdEndpoints = envflag.String("ETCDENDPOINTS", "", "comma separted list of ectd endpoints")
	doToken       = envflag.String("DIGITALOCEAN_ACCESS_TOKEN", "", "DigitalOcean access token")
)

func main() {
	envflag.Parse()

	if *agentName == "" {
		*agentName = generateInstanceID()
	}

	kapi, err := genKeysAPI()
	if err != nil {
		log.WithError(err).Fatal("could not create keys api client")
	}

	ctx := context.Background()
	cm := agent.NewClusterMember(ctx, *agentName, kapi)
	err = cm.Start()
	if err != nil {
		log.WithError(err).Fatal("could not start cluster membership")
	}

	config := &agent.Config{
		DigitalOceanToken: *doToken,
	}
	go updateClusterStatus(cm, config)
	api := agent.New(config)

	errChan := make(chan error)
	go runServer(api, errChan)
	err = <-errChan
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.WithError(err).Panic("unexpected error")
	}
}

func runServer(api *agent.API, errChan chan error) {
	httpServer := graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:    *addr,
			Handler: api.Mux,
		},
	}

	log.WithField("addr", *addr).Info("starting http listener")
	errChan <- httpServer.ListenAndServe()
}

func generateInstanceID() string {
	strlen := 10
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func genKeysAPI() (etcdclient.KeysAPI, error) {
	if *etcdEndpoints == "" {
		return nil, errors.New("missing ETCDENDPOINTS environment variable")
	}

	etcdConfig := etcdclient.Config{
		Endpoints:               []string{},
		Transport:               etcdclient.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	endpoints := strings.Split(*etcdEndpoints, ",")
	for _, ep := range endpoints {
		etcdConfig.Endpoints = append(etcdConfig.Endpoints, ep)
	}

	c, err := etcdclient.New(etcdConfig)
	if err != nil {
		return nil, err
	}

	return etcdclient.NewKeysAPI(c), nil
}

func updateClusterStatus(cm *agent.ClusterMember, config *agent.Config) {
	for {
		select {
		case cs := <-cm.Change():
			log.WithFields(log.Fields{
				"leader":     cs.Leader,
				"node-count": cs.NodeCount,
			}).Info("cluster changed")
			config.Lock()
			config.ClusterStatus = cs
			config.Unlock()
		}
	}
}
