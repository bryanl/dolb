package main

import (
	"flag"
	"math/rand"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/agent"
	etcdclient "github.com/coreos/etcd/client"
)

var (
	etcdEndpoints = flag.String("etcdEndpoints", "", "comma separted list of ectd endpoints")
)

func main() {
	flag.Parse()

	if *etcdEndpoints == "" {
		flag.Usage()
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
	el := agent.NewClusterMember(ctx, generateInstanceID(), c)
	err = el.Start()
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case leader := <-el.Change():
			log.WithField("new-leader", leader).Info("leader changed")
		}
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
