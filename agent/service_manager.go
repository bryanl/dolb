package agent

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/service"
)

type ServiceManager interface {
	AddUpstream(service string, ucr UpstreamCreateRequest) error
	DeleteUpstream(service, upstreamID string) error
	Create(service.ServiceCreateRequest) error
	Services() ([]kvs.Service, error)
	Service(name string) (kvs.Service, error)
}

type ServiceManagerFactory func(c *Config) ServiceManager

type EtcdServiceManager struct {
	Haproxy kvs.Haproxy
	Log     *logrus.Entry
}

var _ ServiceManager = &EtcdServiceManager{}

func NewEtcdServiceManager(c *Config) ServiceManager {
	return &EtcdServiceManager{
		Haproxy: kvs.NewLiveHaproxy(c.KVS, c.IDGen, c.GetLogger()),
		Log:     c.GetLogger(),
	}
}

func (esm *EtcdServiceManager) Create(er service.ServiceCreateRequest) error {
	log := esm.Log

	if er.Name == "" {
		return errors.New("invalid service name")
	}

	if er.Domain != "" && er.Regex != "" {
		return errors.New("only supply a domain or a URL regex, not both")
	}

	if er.Domain != "" {
		log.WithFields(logrus.Fields{
			"domain":       er.Domain,
			"service-name": er.Name,
		}).Info("createing domain service")
		return esm.Haproxy.Domain(er.Name, er.Domain)
	}

	log.WithFields(logrus.Fields{
		"regex":        er.Regex,
		"service-name": er.Name,
	}).Info("creating regex service")
	return esm.Haproxy.URLReg(er.Name, er.Regex)
}

func (esm *EtcdServiceManager) Services() ([]kvs.Service, error) {
	esm.Log.Info("retrieving services")
	return esm.Haproxy.Services()
}

func (esm *EtcdServiceManager) Service(name string) (kvs.Service, error) {
	esm.Log.WithField("service", name).Info("retrieving service")
	return esm.Haproxy.Service(name)
}

func (esm *EtcdServiceManager) AddUpstream(service string, ucr UpstreamCreateRequest) error {
	addr := fmt.Sprintf("%s:%d", ucr.Host, ucr.Port)

	esm.Log.WithFields(logrus.Fields{
		"server": service,
		"host":   ucr.Host,
		"port":   ucr.Port,
	}).Info("adding upstream to server")
	return esm.Haproxy.Upstream(service, addr)
}

func (esm *EtcdServiceManager) DeleteUpstream(service, id string) error {
	esm.Log.WithFields(logrus.Fields{
		"upstream-id":  id,
		"service-name": service,
	}).Info("removing upstream from service")
	return esm.Haproxy.DeleteUpstream(service, id)
}
