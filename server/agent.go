package server

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/do"
)

var (
	discoveryGeneratorURI = "http://discovery.etcd.io/new"
	dropletSize           = "512mb"
)

type AgentBooter interface {
	Create(id int) (*dao.Agent, error)
	Configure(agent *dao.Agent) error
}

type agentBooter struct {
	bo             *BootstrapOptions
	discoveryToken string
}

var _ AgentBooter = &agentBooter{}

func (ab *agentBooter) Create(agentID int) (*dao.Agent, error) {
	bo := ab.bo

	name := fmt.Sprintf("lb-%s-%d", bo.LoadBalancer.ID, agentID)

	a := bo.Config.DBSession.NewAgent()
	a.ClusterID = bo.LoadBalancer.ID
	a.Name = name
	err := bo.Config.DBSession.SaveAgent(a)
	if err != nil {
		return nil, err
	}

	return a, nil

}

func (ab *agentBooter) Configure(dbAgent *dao.Agent) error {
	bo := ab.bo
	bc := bo.BootstrapConfig
	doc := bo.Config.DigitalOcean(bc.DigitalOceanToken)

	ud, err := userData(ab.discoveryToken, dbAgent.ID, bo)
	if err != nil {
		return err
	}

	dcr := do.DropletCreateRequest{
		Name:     dbAgent.Name,
		Region:   bc.Region,
		Size:     dropletSize,
		SSHKeys:  bc.SSHKeys,
		UserData: ud,
	}
	agent, err := doc.CreateAgent(&dcr)
	if err != nil {
		bo.Config.GetLogger().WithError(err).WithField("droplet-name", dbAgent.Name).Error("could not create agent")
		return err
	}

	dnsName := fmt.Sprintf("%s.%s", dbAgent.Name, bo.BootstrapConfig.Region)
	de, err := doc.CreateDNS(dnsName, agent.IPAddresses["public"])
	if err != nil {
		bo.Config.GetLogger().WithError(err).Error("could not assign dns for agent")
		return err
	}

	dbAgent.DropletID = agent.DropletID
	dbAgent.IpID = de.RecordID
	err = bo.Config.DBSession.SaveAgent(dbAgent)
	if err != nil {
		bo.Config.GetLogger().WithError(err).Error("unable to save agent in db")
		return err
	}

	bo.Config.logger.WithFields(logrus.Fields{
		"agent-id":   dbAgent.ID,
		"cluster-id": dbAgent.ClusterID,
		"ip-id":      dbAgent.IpID,
	}).Info("agent configured")

	return nil
}

func discoveryGenerator() (string, error) {
	resp, err := http.Get(discoveryGeneratorURI)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
