package server

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/pkg/cluster"
	"golang.org/x/net/context"
)

// LoadBalancerFactory is an interface that can build LoadBalancers.
type LoadBalancerFactory interface {
	Build(bootstrapConfig *BootstrapConfig) (*entity.LoadBalancer, error)
}

// LoadBalancerFactoryConfig is a configurationg for LoadBalancerFactory.
type LoadBalancerFactoryConfig struct {
	Context          context.Context
	Cluster          cluster.Cluster
	KVS              kvs.KVS
	EntityManager    entity.Manager
	GenerateRandomID func() string
}

type loadBalancerFactory struct {
	entityManager entity.Manager
	logger        *logrus.Entry
	config        *LoadBalancerFactoryConfig
}

var _ LoadBalancerFactory = &loadBalancerFactory{}

// NewLoadBalancerFactory creates an instance of LoadBalancerFactory.
func NewLoadBalancerFactory(config *LoadBalancerFactoryConfig) LoadBalancerFactory {
	logger, ok := config.Context.Value("logger").(*logrus.Entry)
	if !ok {
		logger = logrus.WithFields(logrus.Fields{})
	}

	return &loadBalancerFactory{
		config: config,
		logger: logger,
	}
}

func (lbf *loadBalancerFactory) Build(bootstrapConfig *BootstrapConfig) (*entity.LoadBalancer, error) {
	var err error

	em := lbf.config.EntityManager

	if bootstrapConfig.DigitalOceanToken == "" {
		return nil, fmt.Errorf("DigitalOcean token is required")
	}

	lb := &entity.LoadBalancer{
		ID:                      lbf.config.GenerateRandomID(),
		Name:                    bootstrapConfig.Name,
		Region:                  bootstrapConfig.Region,
		DigitaloceanAccessToken: bootstrapConfig.DigitalOceanToken,
		State: "initialized",
	}

	if err = em.Create(lb); err != nil {
		lbf.logger.WithError(err).Error("unable to create load balancer")
		return nil, err
	}

	defer func() {
		if err != nil {
			lb.State = "invalid"
			if serr := em.Save(lb); serr != nil {
				lbf.logger.WithError(err).Error("unable to change load balancer state to invalid")
			}
		}
	}()

	if err := lbf.config.Cluster.Bootstrap(lb); err != nil {
		lbf.logger.WithError(err).Error("unable to bootstrap load balancer cluster")
		return nil, err
	}

	if _, err := lbf.config.KVS.Set("/dolb/cluster/"+lb.ID, lb.ID, nil); err != nil {
		lbf.logger.WithError(err).Error("unable to create cluster in kvs")
		return nil, err
	}

	lbf.logger.WithFields(logrus.Fields{
		"cluster-name":   lb.Name,
		"cluster-region": lb.Region,
	}).Info("created load balancer")

	return lb, nil
}
