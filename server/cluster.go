package server

import (
	"errors"
	"regexp"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
)

const (
	agentVersion = "0.0.1"
)

var (
	reClusterName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]*?$`)
)

type userDataConfig struct {
	AgentVersion    string
	AgentID         string
	BootstrapConfig *BootstrapConfig
	ClusterID       string
	CoreosToken     string
	ServerURL       string
}

// BootstrapConfig is configuration for Bootstrap.
type BootstrapConfig struct {
	DigitalOceanToken string   `json:"digitalocean_token"`
	Name              string   `json:"name"`
	Region            string   `json:"region"`
	SSHKeys           []string `json:"ssh_keys"`

	RemoteSyslog *RemoteSyslog `json:"remote_syslog"`
}

// HasSyslog returns if a BootstrapConfig has a syslog configuration.
func (bc *BootstrapConfig) HasSyslog() bool {
	return bc.RemoteSyslog != nil
}

// RemoteSyslog is a remote syslog server configuration.
type RemoteSyslog struct {
	EnableSSL bool   `json:"enable_ssl"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
}

// BootstrapOptions are options for the bootstrap process.
type BootstrapOptions struct {
	BootstrapConfig *BootstrapConfig
	LoadBalancer    *dao.LoadBalancer
	Config          *Config
}

// ClusterOps is an interface for cluster operations.
type ClusterOps interface {
	Bootstrap(bo *BootstrapOptions) error
}

// LiveClusterOps are operations for building clusters.
type LiveClusterOps struct {
	AgentBooter func(*BootstrapOptions) AgentBooter
}

var _ ClusterOps = &LiveClusterOps{}

// NewClusterOps creates an instance of clusterOps.
func NewClusterOps() ClusterOps {
	return &LiveClusterOps{
		AgentBooter: func(bo *BootstrapOptions) AgentBooter {

			t, _ := discoveryGenerator()

			return &agentBooter{
				bo:             bo,
				discoveryToken: t,
			}
		},
	}
}

// Bootstrap bootstraps the cluster by creating agents and tracking and their
// completion status.
func (co *LiveClusterOps) Bootstrap(bo *BootstrapOptions) error {
	if bo.Config == nil {
		return errors.New("missing config")
	}

	if bo.BootstrapConfig == nil {
		return errors.New("missing bootstrap config")
	}

	if bo.LoadBalancer == nil {
		return errors.New("missing load balancer")
	}

	if name := bo.BootstrapConfig.Name; !isValidClusterName(name) {
		return errors.New("invalid load balancer name")
	}

	go func() {

		var errors []error
		var wg sync.WaitGroup
		wg.Add(3)

		ab := co.AgentBooter(bo)
		for i := 1; i < 4; i++ {
			go func(id int) {
				defer wg.Done()
				agent, err := ab.Create(id)
				if err != nil {
					bo.Config.GetLogger().
						WithError(err).
						WithFields(logrus.Fields{
						"agent-name":      agent.Name,
						"agent-id":        agent.ID,
						"loadbalancer-id": agent.ClusterID,
					}).Error("could not create agent")
					errors[i] = err
					return
				}

				err = ab.Configure(agent)
				if err != nil {
					bo.Config.GetLogger().
						WithError(err).
						WithFields(logrus.Fields{
						"agent-name":      agent.Name,
						"agent-id":        agent.ID,
						"loadbalancer-id": agent.ClusterID,
					}).Error("could not configure agent")
					errors[i] = err
				}
			}(i)
		}

		wg.Wait()

		for _, err := range errors {
			if err != nil {
				return
			}
		}

		lbState := &LBState{
			lbID:      bo.LoadBalancer.ID,
			logger:    bo.Config.logger,
			dbSession: bo.Config.DBSession,
		}

		go lbState.Track()

	}()
	return nil
}

func isValidClusterName(name string) bool {
	return reClusterName.Match([]byte(name))
}
