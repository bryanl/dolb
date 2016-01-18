package lbfactory

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/pkg/app"
	"github.com/docker/docker/pkg/stringid"
)

// LBFactory is the default implementation of github.com/bryanl/dolb/pkg/app/LBFactory.
type LBFactory struct {
	Context          context.Context
	Cluster          app.Cluster
	KVS              kvs.KVS
	EntityManager    entity.Manager
	GenerateRandomID func() string
	Logger           *logrus.Entry
}

var _ app.LoadBalancerFactory = &LBFactory{}

// New builds an instance of LBFactory.
func New(kv kvs.KVS, em entity.Manager, options ...func(*LBFactory)) (app.LoadBalancerFactory, error) {
	lbf := LBFactory{
		Context:          context.Background(),
		EntityManager:    em,
		KVS:              kv,
		Logger:           logrus.WithFields(logrus.Fields{}),
		GenerateRandomID: stringid.GenerateRandomID,
	}

	for _, option := range options {
		option(&lbf)
	}

	return &lbf, nil
}

// Context returns a function that sets LBFactory Context.
func Context(ctx context.Context) func(*LBFactory) {
	return func(lbf *LBFactory) {
		lbf.Context = ctx
	}
}

// Cluster returns a function that sets LBFactory Cluster.
func Cluster(c app.Cluster) func(*LBFactory) {
	return func(lbf *LBFactory) {
		lbf.Cluster = c
	}
}

// Logger returns a function that sets LBFactory Logger.
func Logger(logctx *logrus.Entry) func(*LBFactory) {
	return func(lbf *LBFactory) {
		lbf.Logger = logctx
	}
}

// GenerateRandomID returns a function that sets LBFactory GenerateRandomID.
func GenerateRandomID(fn func() string) func(*LBFactory) {
	return func(lbf *LBFactory) {
		lbf.GenerateRandomID = fn
	}
}

// Build builds a load balancer.
func (lbf *LBFactory) Build(bootstrapConfig *app.BootstrapConfig) (*entity.LoadBalancer, error) {
	var err error

	em := lbf.EntityManager

	if bootstrapConfig.DigitalOceanToken == "" {
		return nil, fmt.Errorf("DigitalOcean token is required")
	}

	lb := &entity.LoadBalancer{
		ID:                      lbf.GenerateRandomID(),
		Name:                    bootstrapConfig.Name,
		Region:                  bootstrapConfig.Region,
		DigitaloceanAccessToken: bootstrapConfig.DigitalOceanToken,
		State: "initialized",
	}

	if err = em.Create(lb); err != nil {
		lbf.Logger.WithError(err).Error("unable to create load balancer")
		return nil, err
	}

	defer func() {
		if err != nil {
			lb.State = "invalid"
			if serr := em.Save(lb); serr != nil {
				lbf.Logger.WithError(err).Error("unable to change load balancer state to invalid")
			}
		}
	}()

	if _, err := lbf.Cluster.Bootstrap(lb, bootstrapConfig); err != nil {
		lbf.Logger.WithError(err).Error("unable to bootstrap load balancer cluster")
		return nil, err
	}

	if _, err := lbf.KVS.Set("/dolb/cluster/"+lb.ID, lb.ID, nil); err != nil {
		lbf.Logger.WithError(err).Error("unable to create cluster in kvs")
		return nil, err
	}

	lbf.Logger.WithFields(logrus.Fields{
		"cluster-name":   lb.Name,
		"cluster-region": lb.Region,
	}).Info("created load balancer")

	return lb, nil
}
