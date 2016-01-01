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
	Create(service.EndpointRequest) error
	Services() ([]kvs.Service, error)
	Service(name string) (kvs.Service, error)
}

type ServiceManagerFactory func(c *Config) ServiceManager

type EtcdServiceManager struct {
	HKVS *kvs.Haproxy
	Log  *logrus.Entry
}

var _ ServiceManager = &EtcdServiceManager{}

func NewEtcdServiceManager(c *Config) ServiceManager {
	return &EtcdServiceManager{
		HKVS: kvs.NewHaproxy(c.KVS, c.IDGen, c.GetLogger()),
		Log:  c.GetLogger(),
	}
}

func (esm *EtcdServiceManager) Create(er service.EndpointRequest) error {
	log := esm.Log

	if er.ServiceName == "" {
		return errors.New("invalid service name")
	}

	if er.Domain != "" && er.Regex != "" {
		return errors.New("only supply a domain or a URL regex, not both")
	}

	if er.Domain != "" {
		log.WithFields(logrus.Fields{
			"domain":       er.Domain,
			"service-name": er.ServiceName,
		}).Info("createing domain service")
		return esm.HKVS.Domain(er.ServiceName, er.Domain)
	}

	log.WithFields(logrus.Fields{
		"regex":        er.Regex,
		"service-name": er.ServiceName,
	}).Info("creating regex service")
	return esm.HKVS.URLReg(er.ServiceName, er.Regex)
}

func (esm *EtcdServiceManager) Services() ([]kvs.Service, error) {
	esm.Log.Info("retrieving services")
	return esm.HKVS.Services()
}

func (esm *EtcdServiceManager) Service(name string) (kvs.Service, error) {
	esm.Log.WithField("service", name).Info("retrieving service")
	return esm.HKVS.Service(name)
}

func (esm *EtcdServiceManager) AddUpstream(service string, ucr UpstreamCreateRequest) error {
	addr := fmt.Sprintf("%s:%d", ucr.Host, ucr.Port)

	esm.Log.WithFields(logrus.Fields{
		"server": service,
		"host":   ucr.Host,
		"port":   ucr.Port,
	}).Info("adding upstream to server")
	return esm.HKVS.Upstream(service, addr)
}
