package main

import (
	"math/rand"
	"net/http"
	"os"
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
)

func main() {
	envflag.Parse()

	if *agentName == "" {
		*agentName = generateInstanceID()
	}

	if *etcdEndpoints == "" {
		log.Error("missing ETCDENDPOINTS environment variable")
		envflag.PrintDefaults()
		os.Exit(1)
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
		log.Fatal(err)
	}

	ctx := context.Background()
	el := agent.NewClusterMember(ctx, *agentName, c)
	err = el.Start()
	if err != nil {
		log.Fatal(err)
	}

	config := &agent.Config{}

	go func(c *agent.Config) {
		for {
			select {
			case cs := <-el.Change():
				log.WithFields(log.Fields{
					"leader":     cs.Leader,
					"node-count": cs.NodeCount,
				}).Info("cluster changed")
				config.Lock()
				config.ClusterStatus = cs
				config.Unlock()
			}
		}
	}(config)

	api := agent.New(config)

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

	err = <-errChan
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.WithError(err).Panic("unexpected error")
	}

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
