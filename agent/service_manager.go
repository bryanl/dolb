package agent

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/kvs"
	"github.com/bryanl/dolb/service"
)

type ServiceManager interface {
	AddUpstream(svc string, ucr UpstreamCreateRequest) error
	DeleteService(svcName string) error
	DeleteUpstream(svc, upstreamID string) error
	Create(service.ServiceCreateRequest) error
	Services() ([]kvs.Service, error)
	Service(name string) (kvs.Service, error)
}

type ServiceManagerFactory func(c *Config) ServiceManager

type EtcdServiceManager struct {
	Haproxy  kvs.Haproxy
	Firewall kvs.Firewall
	Log      *logrus.Entry
}

var _ ServiceManager = &EtcdServiceManager{}

func NewEtcdServiceManager(c *Config) ServiceManager {
	return &EtcdServiceManager{
		Haproxy:  kvs.NewLiveHaproxy(c.KVS, c.IDGen, c.GetLogger()),
		Firewall: kvs.NewLiveFirewall(c.KVS),
		Log:      c.GetLogger(),
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

	if er.Domain == "" && er.Regex == "" {
		return errors.New("supply a domain or a regex to create a service")
	}

	if er.Domain != "" {
		log.WithFields(logrus.Fields{
			"domain":       er.Domain,
			"service-name": er.Name,
			"port":         er.Port,
		}).Info("creating domain service")
		err := esm.Haproxy.Domain(er.Name, er.Domain, er.Port)
		if err != nil {
			log.WithFields(logrus.Fields{
				"domain":       er.Domain,
				"service-name": er.Name,
				"port":         er.Port,
			}).WithError(err).Error("could not create domain service")
			return err
		}
	} else if er.Regex != "" {
		log.WithFields(logrus.Fields{
			"regex":        er.Regex,
			"service-name": er.Name,
			"port":         er.Port,
		}).Info("creating regex service")
		err := esm.Haproxy.URLReg(er.Name, er.Regex, er.Port)
		if err != nil {
			log.WithFields(logrus.Fields{
				"regex":        er.Regex,
				"service-name": er.Name,
				"port":         er.Port,
			}).Info("creating regex service")
		}
		return err
	} else {
		return errors.New("not sure what type of service to create")
	}

	// TODO clean me up
	log.WithField("port", er.Port).Info("opening firewall port")
	err := esm.Firewall.EnablePort(er.Port)
	if err != nil {
		log.WithError(err).WithField("port", er.Port).Error("unable to open firewall port")
		return err
	}
	log.WithField("port", er.Port).Info("opened firewall port")

	return nil
}

func (esm *EtcdServiceManager) Services() ([]kvs.Service, error) {
	esm.Log.Info("retrieving services")
	return esm.Haproxy.Services()
}

func (esm *EtcdServiceManager) Service(name string) (kvs.Service, error) {
	esm.Log.WithField("service", name).Info("retrieving service")
	return esm.Haproxy.Service(name)
}

func (esm *EtcdServiceManager) AddUpstream(svc string, ucr UpstreamCreateRequest) error {
	addr := fmt.Sprintf("%s:%d", ucr.Host, ucr.Port)

	esm.Log.WithFields(logrus.Fields{
		"server": svc,
		"host":   ucr.Host,
		"port":   ucr.Port,
	}).Info("adding upstream to server")
	return esm.Haproxy.Upstream(svc, addr)
}

func (esm *EtcdServiceManager) DeleteUpstream(svc, id string) error {
	esm.Log.WithFields(logrus.Fields{
		"upstream-id":  id,
		"service-name": svc,
	}).Info("removing upstream from service")
	return esm.Haproxy.DeleteUpstream(svc, id)
}

func (esm *EtcdServiceManager) DeleteService(svc string) error {
	esm.Log.WithFields(logrus.Fields{
		"service-name": svc,
	}).Info("removing service")
	return esm.Haproxy.DeleteService(svc)
}
