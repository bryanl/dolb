package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/server"
	"github.com/bryanl/dolb/service"
)

var (
	haproxyDiscoverKey = "/haproxy-discover/services"
)

// Agent is the load balancer agent. It controlls all things load balancer.
type Agent struct {
	ClusterMember     *ClusterMember
	Config            *Config
	FloatingIPManager FloatingIPManager
}

// New creates a load balancer agent.
func New(cm *ClusterMember, config *Config) (*Agent, error) {
	fim, err := NewFloatingIPManager(config)
	if err != nil {
		return nil, err
	}

	return &Agent{
		ClusterMember:     cm,
		Config:            config,
		FloatingIPManager: fim,
	}, nil
}

func pingServer(config *Config) error {
	u, err := url.Parse(config.ServerURL)
	if err != nil {
		return err
	}

	u.Path = service.PingPath

	rr := server.PingRequest{
		AgentID:     config.AgentID,
		ClusterID:   config.ClusterID,
		ClusterName: config.ClusterName,
		FloatingIP:  config.ClusterStatus.FloatingIP,
		Host:        config.Name,
		IsLeader:    config.ClusterStatus.IsLeader,
	}

	if config.ClusterStatus.IsLeader {
		rr.FloatingIP = config.ClusterStatus.FloatingIP
	}

	b, err := json.Marshal(&rr)
	buf := bytes.NewReader(b)

	_, err = http.Post(u.String(), "application/json", buf)
	if err != nil {
		return err
	}

	config.logger.WithFields(log.Fields{
		"server-url":   config.ServerURL,
		"cluster-name": config.ClusterName,
	}).Info("register agent")

	return nil
}

// PollClusterStatus polls cluster status to see if this node is the leader.
func (a *Agent) PollClusterStatus() {
	for {
		select {
		case cs := <-a.ClusterMember.Change():
			a.Config.logger.WithFields(log.Fields{
				"leader":     cs.Leader,
				"node-count": cs.NodeCount,
			}).Info("cluster changed")
			a.Config.Lock()
			a.Config.ClusterStatus = cs

			node, err := a.Config.KVS.Get(fipKey, nil)
			if err == nil {
				a.Config.ClusterStatus.FloatingIP = node.Value
			}

			a.Config.Unlock()

			a.Config.logger.WithFields(log.Fields{
				"is-leader": cs.IsLeader,
			}).Info("leader check")

			if cs.IsLeader {
				hkvs := NewHaproxyKVS(a.Config.KVS)

				err := hkvs.Init()
				if err != nil {
					a.Config.logger.WithError(err).Error("could not create haproxy keys")
				}

				handleLeaderElection(a)
			}

			if err := pingServer(a.Config); err != nil {
				a.Config.logger.WithError(err).Error("could not register agent")
			}
		}
	}
}

func handleLeaderElection(a *Agent) {
	ip, err := a.FloatingIPManager.Reserve()
	if err != nil {
		a.Config.logger.WithError(err).Error("could not retrieve floating ip for agent")
	}

	a.Config.ClusterStatus.FloatingIP = ip

	a.Config.logger.WithField("cluster-ip", ip).Info("retrieved cluster ip")
}
