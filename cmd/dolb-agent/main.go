package main

import (
	"flag"
	"math/rand"
	"os"
	"strings"
	"time"

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

	lc := agent.NewEtcdLeader(generateInstanceID(), c)
	err = lc.Start()
	if err != nil {
		log.Fatal(err)
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
