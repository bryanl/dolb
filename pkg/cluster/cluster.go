package cluster

import (
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
func New(ab app.AgentBuilder, options ...func(*Cluster)) app.Cluster {
	c := &Cluster{
		AgentBuilder: ab,
	}

	for _, option := range options {
		option(c)
	}

	if c.Logger == nil {
		c.Logger = app.DefaultLogger()
	}

	return c
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

				logger := c.Logger.WithFields(logrus.Fields{
					"agent-iteration": id,
					"loadbalancer-id": lb.ID,
				})

				logger.Info("creating agent")

				agent, err := c.AgentBuilder.Create(id)
				if err != nil {
					logger.
						WithError(err).
						WithFields(logrus.Fields{
						"agent-name": agent.DropletName,
						"agent-id":   agent.ID,
					}).Error("could not create agent")
					errors[i] = err
					return
				}

				logger.Info("configuring agent")

				err = c.AgentBuilder.Configure(agent)
				if err != nil {
					logger.
						WithError(err).
						WithFields(logrus.Fields{
						"agent-name": agent.DropletName,
						"agent-id":   agent.ID,
					}).Error("could not configure agent")
					errors[i] = err
					return
				}

				logger.Info("created agent")

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
