package cluster

import (
	"errors"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/pkg/app"
)

const (
	agentCount = 3
)

// Cluster managements load balancer agent clusters.
type Cluster struct {
	AgentBuilder app.AgentBuilder
	Logger       *logrus.Entry
}

var _ app.Cluster = &Cluster{}

// New builds a cluster.
func New(options ...func(*Cluster)) (app.Cluster, error) {
	c := &Cluster{}

	for _, option := range options {
		option(c)
	}

	if c.Logger == nil {
		c.Logger = logrus.WithFields(logrus.Fields{})
	}

	if c.AgentBuilder == nil {
		return nil, errors.New("cluster AgentBuilder is nil")
	}

	return c, nil
}

// AgentBuilder sets Cluster AgentBuilder.
func AgentBuilder(ab app.AgentBuilder) func(*Cluster) {
	return func(c *Cluster) {
		c.AgentBuilder = ab
	}
}

// Logger sets Cluster Logger.
func Logger(l *logrus.Entry) func(*Cluster) {
	return func(c *Cluster) {
		c.Logger = l
	}
}

// Bootstrap bootstraps an agent cluster.
func (c Cluster) Bootstrap(lb *entity.LoadBalancer, bc *app.BootstrapConfig) (chan int, error) {
	// TODO validate we have enough to get started or return an error

	statusChan := make(chan int)

	go func() {
		var errors []error
		var wg sync.WaitGroup
		wg.Add(agentCount)

		for i := 0; i < agentCount; i++ {
			go func(id int) {
				defer func() {
					statusChan <- i
				}()
				agent, err := c.AgentBuilder.Create(id)
				if err != nil {
					c.Logger.
						WithError(err).
						WithFields(logrus.Fields{
						"agent-name":      agent.DropletName,
						"agent-id":        agent.ID,
						"loadbalancer-id": agent.ClusterID,
					}).Error("could not create agent")
					errors[i] = err
					return
				}

				err = c.AgentBuilder.Configure(agent)
				if err != nil {
					c.Logger.
						WithError(err).
						WithFields(logrus.Fields{
						"agent-name":      agent.DropletName,
						"agent-id":        agent.ID,
						"loadbalancer-id": agent.ClusterID,
					}).Error("could not configure agent")
					errors[i] = err
					return
				}

			}(i + 1)
		}

		wg.Wait()

		for _, err := range errors {
			if err != nil {
				return
			}
		}
	}()
	return statusChan, nil
}
