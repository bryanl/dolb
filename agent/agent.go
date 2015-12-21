package agent

import log "github.com/Sirupsen/logrus"

// Agent is the load balancer agent. It controlls all things load balancer.
type Agent struct {
	ClusterMember     *ClusterMember
	Config            *Config
	FloatingIPManager *FloatingIPManager
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
				// handle floating ip
				ip, err := a.FloatingIPManager.Reserve()
				if err != nil {
					log.WithError(err).Error("could not retrieve floating ip for agent")
				}

				a.Config.ClusterStatus.FloatingIP = ip

				log.WithField("cluster-ip", ip).Info("retrieved cluster ip")
			}
		}
	}
}
