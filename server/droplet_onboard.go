package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/digitalocean/godo"
)

type DropletOnboard interface {
	setup()
}

// DropletOnboard onboards a droplet.
type LiveDropletOnboard struct {
	Droplet godo.Droplet

	agentID    string
	domain     string
	godoClient *godo.Client
	logger     *logrus.Entry
	session    dao.Session

	assignDNS               func(ldo *LiveDropletOnboard) (*godo.DomainRecord, error)
	publicIPV4Address       func(ldo *LiveDropletOnboard) (string, error)
	waitUntilDropletCreated func(ldo *LiveDropletOnboard) error
	createAction            func(ldo *LiveDropletOnboard) (*godo.Action, error)
}

func NewDropletOnboard(d godo.Droplet, agentID string, client *godo.Client, config *Config) *LiveDropletOnboard {
	return &LiveDropletOnboard{
		Droplet:    d,
		agentID:    agentID,
		domain:     config.BaseDomain,
		godoClient: client,
		logger:     config.logger,
		session:    config.DBSession,

		assignDNS:               assignDNS,
		publicIPV4Address:       publicIPV4Address,
		waitUntilDropletCreated: waitUntilDropletCreated,
		createAction:            createAction,
	}
}

func (dro *LiveDropletOnboard) setup() {
	logger := dro.logger.WithFields(logrus.Fields{
		"droplet-id": dro.Droplet.ID,
		"agent-id":   dro.agentID,
	})
	r, err := dro.assignDNS(dro)
	if err != nil {
		logger.WithError(err).Error("unable to set up droplet in dns")
		return
	}

	adc := &dao.AgentDOConfig{
		ID:        dro.agentID,
		IPID:      r.ID,
		DropletID: dro.Droplet.ID,
	}
	a, err := dro.session.UpdateAgentDOConfig(adc)
	if err != nil {
		logger.WithError(err).Error("unable to save agent info")
	}

	logger.WithFields(logrus.Fields{
		"ip-id": a.IPID,
	}).Info("droplet setup")
}

// assignDNS assigns a hostname in DNS for the droplet.
func assignDNS(dro *LiveDropletOnboard) (*godo.DomainRecord, error) {
	ip, err := dro.publicIPV4Address(dro)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%s.%s", dro.Droplet.Name, dro.Droplet.Region.Slug)

	drer := &godo.DomainRecordEditRequest{
		Type: "A",
		Name: name,
		Data: ip,
	}

	r, _, err := dro.godoClient.Domains.CreateRecord(dro.domain, drer)
	return r, err
}

// publicIPV4Address retrieves the public IPV4 address for a Droplet.
func publicIPV4Address(dro *LiveDropletOnboard) (string, error) {
	err := dro.waitUntilDropletCreated(dro)
	if err != nil {
		return "", err
	}

	d, _, err := dro.godoClient.Droplets.Get(dro.Droplet.ID)
	if err != nil {
		return "", err
	}

	var publicIP string
	for _, n := range d.Networks.V4 {
		if n.Type == "public" {
			publicIP = n.IPAddress
			break
		}
	}

	if publicIP == "" {
		return "", errors.New("unable to find public ipv4 address for droplet")
	}

	return publicIP, nil
}

// waitUntilDropletCreated blocks until the droplet is created. If there is an
// error, it is returned.
func waitUntilDropletCreated(dro *LiveDropletOnboard) error {
	a, err := dro.createAction(dro)
	if err != nil {
		return err
	}

	for {
		ca, _, err := dro.godoClient.Actions.Get(a.ID)
		if err != nil {
			return err
		}

		switch ca.Status {
		case "completed":
			return nil
		case "errored":
			return errors.New("action errored")
		case "in-progress":
			continue
		default:
			return fmt.Errorf("unknown action status: %v")
		}

		time.Sleep(10 * time.Second)
	}

	return nil
}

// createAction finds the create action for a droplet.
func createAction(dro *LiveDropletOnboard) (*godo.Action, error) {
	actions, _, err := dro.godoClient.Droplets.Actions(dro.Droplet.ID, nil)
	if err != nil {
		return nil, err
	}

	for _, a := range actions {
		if a.Type == "create" {
			return &a, nil
		}
	}

	return nil, errors.New("could not find create action for droplet")
}
