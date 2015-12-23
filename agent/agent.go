package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/server"
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

	err = register(config)
	if err != nil {
		return nil, err
	}

	return &Agent{
		ClusterMember:     cm,
		Config:            config,
		FloatingIPManager: fim,
	}, nil
}

func register(config *Config) error {
	u, err := url.Parse(config.ServerURL)
	if err != nil {
		return err
	}

	u.Path = "/register"

	rr := server.RegisterRequest{
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
			log.WithFields(log.Fields{
				"leader":     cs.Leader,
				"node-count": cs.NodeCount,
			}).Info("cluster changed")
			a.Config.Lock()
			a.Config.ClusterStatus = cs

			resp, err := a.Config.KeysAPI.Get(a.Config.Context, fipKey, nil)
			if err == nil {
				a.Config.ClusterStatus.FloatingIP = resp.Node.Value
			}

			a.Config.Unlock()

			if cs.IsLeader {
				handleLeaderElection(a)
			}
		}
	}
}

func handleLeaderElection(a *Agent) {
	ip, err := a.FloatingIPManager.Reserve()
	if err != nil {
		log.WithError(err).Error("could not retrieve floating ip for agent")
	}

	a.Config.ClusterStatus.FloatingIP = ip

	log.WithField("cluster-ip", ip).Info("retrieved cluster ip")
}
