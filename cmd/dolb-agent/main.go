package main

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/agent"
	"github.com/bryanl/dolb/firewall"
	"github.com/bryanl/dolb/kvs"
	"github.com/ianschenck/envflag"
	"github.com/tylerb/graceful"
)

var (
	addr          = envflag.String("ADDR", ":8889", "listen address")
	agentID       = envflag.String("AGENT_ID", "", "agent id")
	agentName     = envflag.String("AGENT_NAME", "", "agent name")
	agentRegion   = envflag.String("AGENT_REGION", "", "agent DigitalOcean region")
	clusterID     = envflag.String("CLUSTER_ID", "", "cluster id")
	clusterName   = envflag.String("CLUSTER_NAME", "", "cluster name")
	etcdEndpoints = envflag.String("ETCDENDPOINTS", "", "comma separted list of ectd endpoints")
	dropletID     = envflag.String("DROPLET_ID", "", "current droplet id")
	doToken       = envflag.String("DIGITALOCEAN_ACCESS_TOKEN", "", "DigitalOcean access token")
	serverURL     = envflag.String("SERVER_URL", "", "DOLB Server URL")
)

func main() {
	envflag.Parse()

	if *agentID == "" {
		log.Fatal("invalid AGENT_ID environment variable")
	}

	if *agentName == "" {
		*agentName = generateInstanceID()
	}

	if *agentRegion == "" {
		log.Fatal("invalid AGENT_REGION environment variable")
	}

	if *clusterID == "" {
		log.Fatal("invalid CLUSTER_ID environment variable")
	}

	if *clusterName == "" {
		log.Fatal("invalid CLUSTER_NAME environment variable")
	}

	if *dropletID == "" {
		log.Fatal("invalid DROPLET_ID environment variable")
	}

	if *serverURL == "" {
		log.Fatal("invalid SERVER_URL environment variable")
	}

	// FIXME is this too much config?
	config := &agent.Config{
		AgentID:           *agentID,
		DigitalOceanToken: *doToken,
		ClusterID:         *clusterID,
		ClusterName:       *clusterName,
		Context:           context.Background(),
		DropletID:         *dropletID,
		Region:            *agentRegion,
		Name:              *agentName,
		ServerURL:         *serverURL,
		ServiceManagerFactory: func(c *agent.Config) agent.ServiceManager {
			return agent.NewEtcdServiceManager(c)
		},
	}

	logger := log.WithField("agent-name", *agentName)
	config.SetLogger(logger)

	ic := firewall.NewIptablesCommand()
	config.Firewall = firewall.NewIptablesFirewall(ic, logger)

	kapi, err := kvs.NewKeysAPI(*etcdEndpoints, nil)
	if err != nil {
		log.WithError(err).Fatal("could not create keys api client")
	}

	config.KVS = kvs.NewEtcd(config.Context, kapi)

	cm := agent.NewClusterMember(*agentName, config)
	err = cm.Start()
	if err != nil {
		log.WithError(err).Fatal("could not start cluster membership")
	}

	a, err := agent.New(cm, config)
	if err != nil {
		log.WithError(err).Fatal("could not create agent")
	}

	go a.PollClusterStatus()
	go a.PollFirewall()

	api := agent.NewAPI(config)

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
