package agent

import (
	"strconv"

	"github.com/bryanl/dolb/do"
	etcdclient "github.com/coreos/etcd/client"
	"github.com/digitalocean/godo"
	"golang.org/x/net/context"
)

var (
	fipKey        = "/agent/floating_ip"
	fipDropletKey = "/agent/floating_ip_droplet"
)

// FloatingIPManager manages DigitalOcean floating ips for the agent.
type FloatingIPManager struct {
	context    context.Context
	dropletID  string
	godoClient *godo.Client
	kapi       etcdclient.KeysAPI
	name       string
	region     string

	assignNewIP func(*FloatingIPManager) (string, error)
	existingIP  func(*FloatingIPManager) (string, error)
}

// NewFloatingIPManager creates a FloatingIPManager.
func NewFloatingIPManager(config *Config) *FloatingIPManager {
	return &FloatingIPManager{
		context:    config.Context,
		dropletID:  config.DropletID,
		godoClient: do.GodoClientFactory(config.DigitalOceanToken),
		kapi:       config.KeysAPI,
		region:     config.Region,

		assignNewIP: assignNewIP,
		existingIP:  existingIP,
	}
}

// Reserve reserves a floating ip.
func (fim *FloatingIPManager) Reserve(cs ClusterStatus) (string, error) {
	ip, err := fim.existingIP(fim)
	if err != nil {
		if cerr, ok := err.(*etcdclient.Error); ok {
			switch cerr.Code {
			case etcdclient.ErrorCodeKeyNotFound:
				ip, err = fim.assignNewIP(fim)
			}
		} else {
			// who knows how we got to this state?
			return "", err
		}
	}

	fip, _, err := fim.godoClient.FloatingIPs.Get(ip)
	if err != nil {
		return "", err
	}

	id, err := fim.dropletIDInt()
	if err != nil {
		return "", err
	}

	if fip.Droplet.ID != id {
		_, _, err = fim.godoClient.FloatingIPActions.Assign(ip, id)
		if err != nil {
			return "", err
		}
	}

	return ip, nil
}

func existingIP(fim *FloatingIPManager) (string, error) {
	resp, err := fim.kapi.Get(fim.context, fipKey, nil)
	if err != nil {
		return "", err
	}

	return resp.Node.Value, nil
}

func assignNewIP(fim *FloatingIPManager) (string, error) {
	id, err := fim.dropletIDInt()
	if err != nil {
		return "", err
	}

	ficr := &godo.FloatingIPCreateRequest{
		Region:    fim.region,
		DropletID: id,
	}
	fip, _, err := fim.godoClient.FloatingIPs.Create(ficr)
	if err != nil {
		return "", err
	}

	_, err = fim.kapi.Set(fim.context, fipKey, fip.IP, nil)
	if err != nil {
		return "", err
	}

	return fip.IP, nil
}

func (fim *FloatingIPManager) dropletIDInt() (int, error) {
	return strconv.Atoi(fim.dropletID)
}
