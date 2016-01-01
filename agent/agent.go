package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/firewall"
	"github.com/bryanl/dolb/kvs"
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

	resp, err := http.Post(u.String(), "application/json", buf)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		config.logger.WithField("status-code", resp.StatusCode).Warning("unable to ping server")
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
				hkvs := kvs.NewHaproxy(a.Config.KVS, a.Config.IDGen, a.Config.GetLogger())

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

func (a *Agent) PollFirewall() {
	log := a.Config.logger
	fw := a.Config.Firewall
	fkvs := kvs.NewFirewallKVS(a.Config.KVS)

	err := fkvs.Init()
	if err != nil {
		log.WithError(err).Error("unable to start firewall poller")
	}

	ticker := time.NewTicker(time.Second * 5)

	log.Info("starting firewall poller")

	for {
		select {
		case <-ticker.C:
			if a.Config.ClusterStatus.IsLeader {
				state, err := fw.State()
				if err != nil {
					log.WithError(err).Error("unable to load firewall state")
					continue
				}

				rules, err := state.Rules()
				if err != nil {
					log.WithError(err).Error("unable to load firewall rules")
					continue
				}

				m := map[int]*firewall.Rule{}
				for _, r := range rules {
					m[r.Destination] = &r
				}

				ports, err := fkvs.Ports()
				if err != nil {
					log.WithError(err).Error("unable to load ports from kvs")
					continue
				}

				for _, p := range ports {
					if m[p.Port] == nil {
						// port rule doesn't exist in iptables
						if p.Enabled {
							log.WithField("firewall-port", p.Port).Info("opening firewall port")
							err = fw.Open(p.Port)
							if err != nil {
								log.WithError(err).WithField("firewall-port", p.Port).Error("unable to open port")
							}
						}
					} else {
						// port rule exists in iptables
						if !p.Enabled {
							log.WithField("firewall-port", p.Port).Info("closing firewall port")
							err = fw.Close(p.Port)
							if err != nil {
								log.WithError(err).WithField("firewall-port", p.Port).Error("unable to close port")
							}
						}
					}
				}

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
