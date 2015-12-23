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

		assignNewIP: assignNewIP,
		existingIP:  existingIP,
	}, nil
}

// Reserve reserves a floating ip.
func (fim *EtcdFloatingIPManager) Reserve() (string, error) {
	ip, err := fim.existingIP(fim)
	if err != nil {
		if eerr, ok := err.(etcdclient.Error); ok && eerr.Code == etcdclient.ErrorCodeKeyNotFound {
			ip, err = fim.assignNewIP(fim)
			if err != nil {
				logrus.WithError(err).Error("could not assign ip")
				return "", err
			}
		} else {
			// who knows how we got to this state?
			logrus.WithError(err).WithField("raw-error", fmt.Sprintf("%#v", err)).Error("unknown error when checking for existing ip")
			return "", err
		}
	}

	fip, _, err := fim.godoClient.FloatingIPs.Get(ip)
	if err != nil {
		logrus.WithField("fip", ip).WithError(err).Error("could not retrieve floating ip from DigitalOcean")
		return "", err
	}

	id, err := fim.dropletIDInt()
	if err != nil {
		return "", err
	}

	if fip.Droplet.ID != id {
		logrus.WithFields(logrus.Fields{
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

	ACTION_CHECK:
		for {
			action, _, err := fim.godoClient.FloatingIPActions.Get(ip, actionID)
			if err != nil {
				return "", fmt.Errorf("could not poll action progress: %v", err)
			}

			switch action.Status {
			case "completed":
				break ACTION_CHECK
			case "errored":
				return "", errors.New("assign IP action failed")
			case "in-progress":
				continue
			default:
				return "", fmt.Errorf("unknown action status when assigning ip: %v", action.Status)
			}

			time.Sleep(time.Second)
		}
	}

	return ip, nil
}

func existingIP(fim *EtcdFloatingIPManager) (string, error) {
	node, err := fim.fipKVS.Get(fipKey, nil)
	if err != nil {
		return "", err
	}

	return node.Value, nil
}

func assignNewIP(fim *EtcdFloatingIPManager) (string, error) {
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
