package agent

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/do"
	etcdclient "github.com/coreos/etcd/client"
	"github.com/digitalocean/godo"
	"golang.org/x/net/context"
)

var (
	fipKey            = "/agent/floating_ip"
	fipDropletKey     = "/agent/floating_ip_droplet"
	actionPollTimeout = time.Second
)

// FloatingIPManager manages DigitalOcean floating ips for the agent.
type FloatingIPManager interface {
	Reserve() (string, error)
}

// EtcdFloatingIPManager manages DigitalOcean floating ips for the agent.
type EtcdFloatingIPManager struct {
	context    context.Context
	dropletID  string
	godoClient *godo.Client
	fipKVS     *FipKVS
	locker     Locker
	name       string
	logger     *logrus.Entry

	assignNewIP func(*EtcdFloatingIPManager) (string, error)
	existingIP  func(*EtcdFloatingIPManager) (string, error)
}

var _ FloatingIPManager = &EtcdFloatingIPManager{}

// NewFloatingIPManager creates a FloatingIPManager.
func NewFloatingIPManager(config *Config) (*EtcdFloatingIPManager, error) {
	if config.DigitalOceanToken == "" {
		return nil, errors.New("requires DigitalOceanToken")
	}

	locker := &etcdLocker{
		context: config.Context,
		key:     "/agent/floating_ip",
		who:     config.Name,
		kvs:     config.KVS,
	}

	return &EtcdFloatingIPManager{
		context:    config.Context,
		dropletID:  config.DropletID,
		godoClient: do.GodoClientFactory(config.DigitalOceanToken),
		fipKVS:     NewFipKVS(config.KVS),
		locker:     locker,
		logger:     config.logger,

		assignNewIP: assignNewIP,
		existingIP:  existingIP,
	}, nil
}

// Reserve reserves a floating ip.
func (fim *EtcdFloatingIPManager) Reserve() (string, error) {
	fim.logger.Info("reserving floating ip")

	ip, err := fim.existingIP(fim)
	if err != nil {
		if kverr, ok := err.(*KVError); ok {
			if eerr, ok := kverr.err.(etcdclient.Error); ok && eerr.Code == etcdclient.ErrorCodeKeyNotFound {
				ip, err = fim.assignNewIP(fim)
				if err != nil {
					logrus.WithError(err).Error("could not assign ip")
					return "", err
				}
			} else {
				// who knows how we got to this state?
				logrus.WithError(err).Error("unknown error when checking for existing ip")
				fim.logger.WithField("raw-error", fmt.Sprintf("%#v", err)).Info("extra debug info")
				return "", err
			}
		} else {
			// who knows how we got to this state?
			logrus.WithError(err).Error("unknown error when checking for existing ip")
			fim.logger.WithField("raw-error", fmt.Sprintf("%#v", err)).Info("extra debug info")
			return "", err
		}
	}

	fim.logger.WithField("current-fip", ip).Info("existing ip")

	fip, _, err := fim.godoClient.FloatingIPs.Get(ip)
	if err != nil {
		fim.logger.WithField("fip", ip).WithError(err).Error("could not retrieve floating ip from DigitalOcean")
		return "", err
	}

	id, err := fim.dropletIDInt()
	if err != nil {
		return "", err
	}

	fim.logger.WithFields(logrus.Fields{
		"current-fip-droplet-id": fip.Droplet.ID,
		"my-droplet-id":          id,
	}).Info("floating ip check")

	if fip.Droplet.ID != id {
		fim.logger.WithFields(logrus.Fields{
			"current-id": fip.Droplet.ID,
			"wanted-id":  id,
		}).Info("moving floating ip")
		fim.locker.Lock()
		defer fim.locker.Unlock()

		action, _, err := fim.godoClient.FloatingIPActions.Assign(ip, id)
		if err != nil {
			logrus.WithError(err).Error("could not retrieve DigitalOcean floating ip to current droplet")
			return "", err
		}

		actionID := action.ID

		for {
			action, _, err := fim.godoClient.FloatingIPActions.Get(ip, actionID)
			if err != nil {
				return "", fmt.Errorf("could not poll action progress: %v", err)
			}

			switch action.Status {
			case "completed":
				return ip, nil
			case "errored":
				return "", errors.New("assign IP action failed")
			case "in-progress":
				continue
			default:
				return "", fmt.Errorf("unknown action status when assigning ip: %v", action.Status)
			}

			time.Sleep(time.Second)
		}
	} else {
		fim.logger.Info("leader has fip")
	}

	return ip, nil
}

func existingIP(fim *EtcdFloatingIPManager) (string, error) {
	fim.logger.Info("checking for existing floating ip")
	node, err := fim.fipKVS.Get(fipKey, nil)
	if err != nil {
		return "", err
	}

	return node.Value, nil
}

func assignNewIP(fim *EtcdFloatingIPManager) (string, error) {
	fim.logger.Info("assigning floating ip")
	id, err := fim.dropletIDInt()
	if err != nil {
		return "", err
	}

	ficr := &godo.FloatingIPCreateRequest{
		DropletID: id,
	}
	fip, _, err := fim.godoClient.FloatingIPs.Create(ficr)
	if err != nil {
		return "", err
	}

	_, err = fim.fipKVS.Set(fipKey, fip.IP, nil)
	if err != nil {
		return "", err
	}

	return fip.IP, nil
}

func (fim *EtcdFloatingIPManager) dropletIDInt() (int, error) {
	return strconv.Atoi(fim.dropletID)
}
