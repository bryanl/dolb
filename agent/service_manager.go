package agent

import (
	"errors"

	"github.com/Sirupsen/logrus"
)

type ServiceManager interface {
	Create(EndpointRequest) error
}

type ServiceManagerFactory func(c *Config) ServiceManager

type EtcdServiceManager struct {
	HKVS *HaproxyKVS
	Log  *logrus.Entry
}

var _ ServiceManager = &EtcdServiceManager{}

func NewEtcdServiceManager(c *Config) ServiceManager {
	return &EtcdServiceManager{
		HKVS: NewHaproxyKVS(c.KVS),
		Log:  c.GetLogger(),
	}
}

func (esm *EtcdServiceManager) Create(er EndpointRequest) error {
	log := esm.Log

	if er.ServiceName == "" {
		return errors.New("invalid service name")
	}

	if er.Domain != "" && er.Regex != "" {
		return errors.New("only supply a domain or a URL regex, not both")
	}

	if er.Domain != "" {
		log.WithFields(logrus.Fields{
			"domain": er.Domain,
		}).Info("createing domain service")
		return esm.HKVS.Domain(er.ServiceName, er.Domain)
	}

	log.WithFields(logrus.Fields{
		"regex": er.Regex,
	}).Info("createing regex service")
	return esm.HKVS.URLReg(er.ServiceName, er.Regex)
}
