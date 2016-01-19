package agentbuilder

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dolbutil"
	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/pkg/agentuserdata"
	"github.com/bryanl/dolb/pkg/app"
	"github.com/bryanl/dolb/pkg/doclient"
)

var (
	discoveryGeneratorURI = "http://discovery.etcd.io/new"
	dropletSize           = "512mb"
)

// AgentBuilder creates and configure agent droplets.
type AgentBuilder struct {
	EntityManager        entity.Manager
	GenerateRandomID     func() string
	GenerateUserData     func(*agentuserdata.Config) (string, error)
	GenerateDiscoveryURL func() string
	Logger               *logrus.Entry

	bootstrapConfig *app.BootstrapConfig
	lb              *entity.LoadBalancer
	discoveryURL    string

	DOClientFactory func(token string) app.DOClient
}

var _ app.AgentBuilder = &AgentBuilder{}

func defaultDOClientFactory(token string) app.DOClient {
	return doclient.New(token)
}

// New builds an AgentBuilder.
func New(lb *entity.LoadBalancer, bootstrapConfig *app.BootstrapConfig, em entity.Manager, options ...func(*AgentBuilder)) *AgentBuilder {
	agentBuilder := AgentBuilder{
		EntityManager:   em,
		bootstrapConfig: bootstrapConfig,
		lb:              lb,
	}

	for _, option := range options {
		option(&agentBuilder)
	}

	if agentBuilder.GenerateDiscoveryURL == nil {
		agentBuilder.GenerateDiscoveryURL = defaultDiscoveryURL
	}

	if agentBuilder.GenerateUserData == nil {
		agentBuilder.GenerateUserData = func(c *agentuserdata.Config) (string, error) {
			return agentuserdata.Generate(c)
		}
	}

	if agentBuilder.GenerateRandomID == nil {
		agentBuilder.GenerateRandomID = dolbutil.GenerateRandomID
	}

	if agentBuilder.DOClientFactory == nil {
		agentBuilder.DOClientFactory = defaultDOClientFactory
	}

	if agentBuilder.Logger == nil {
		agentBuilder.Logger = app.DefaultLogger()
	}

	agentBuilder.discoveryURL = agentBuilder.GenerateDiscoveryURL()

	return &agentBuilder
}

// DOClientFactory sets the DOClientFactory for an AgentBuilder.
func DOClientFactory(fn func(string) app.DOClient) func(*AgentBuilder) {
	return func(ab *AgentBuilder) {
		ab.DOClientFactory = fn
	}
}

// GenerateRandomID sets GenerateRandomID for an AgentBuilder.
func GenerateRandomID(fn func() string) func(*AgentBuilder) {
	return func(ab *AgentBuilder) {
		ab.GenerateRandomID = fn
	}
}

// GenerateUserData sets GenerateUserData for an AgentBuilder.
func GenerateUserData(fn func(c *agentuserdata.Config) (string, error)) func(*AgentBuilder) {
	return func(ab *AgentBuilder) {
		ab.GenerateUserData = fn
	}
}

// GenerateDiscoveryURL sets DiscoveryURL for an AgentBuilder.
func GenerateDiscoveryURL(fn func() string) func(*AgentBuilder) {
	return func(ab *AgentBuilder) {
		ab.GenerateDiscoveryURL = fn
	}
}

// Logger sets Logger for an AgentBuilder.
func Logger(logger *logrus.Entry) func(*AgentBuilder) {
	return func(ab *AgentBuilder) {
		ab.Logger = logger
	}
}

// Create Creates a droplet in the database.
func (ab *AgentBuilder) Create(id int) (*entity.Agent, error) {
	var err error
	defer func() {
		if err != nil {
			ab.Logger.WithError(err).WithField("agent-instance", id).Error("unable to create agent")
		}
	}()
	name := fmt.Sprintf("agent-%s-%d", ab.lb.ID, id)

	agent := &entity.Agent{
		ID:          ab.GenerateRandomID(),
		ClusterID:   ab.lb.ID,
		DropletName: name,
		Region:      ab.lb.Region,
	}

	if err = ab.EntityManager.Create(agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// Configure builds and configures an agent droplet.
func (ab *AgentBuilder) Configure(agent *entity.Agent) error {
	var err error
	defer func() {
		if err != nil {
			ab.Logger.WithError(err).WithField("agent-id", agent.ID).Error("unable to configure agent")
		} else {
			ab.Logger.WithField("agent-id", agent.ID).Info("agent configured")
		}
	}()

	userDataConfig := &agentuserdata.Config{
		AgentVersion:    "0.0.2", // TODO where is this coming from?
		AgentID:         agent.ID,
		BootstrapConfig: ab.bootstrapConfig,
		ClusterID:       agent.ClusterID,
		CoreosToken:     ab.discoveryURL,
		ServerURL:       "https://dolb.ngrok.io", // TODO this needs to be injected
	}

	userData, err := ab.GenerateUserData(userDataConfig)
	if err != nil {
		return err
	}

	agentCreateRequest := &app.AgentCreateRequest{
		Agent:    agent,
		SSHKeys:  ab.bootstrapConfig.SSHKeys,
		Size:     dropletSize,
		UserData: userData,
	}

	doClient := ab.DOClientFactory(ab.bootstrapConfig.DigitalOceanToken)
	agentCreateResponse, err := doClient.CreateAgent(agentCreateRequest)
	if err != nil {
		return err
	}
	agent.DropletID = agentCreateResponse.DropletID

	dnsName := ab.agentDNSName(agent)
	dnsEntry, err := doClient.CreateDNS(dnsName, agentCreateResponse.PublicIPAddress)
	if err != nil {
		return err
	}

	agent.DNSID = dnsEntry.RecordID

	return ab.EntityManager.Save(agent)
}

func (ab *AgentBuilder) agentDNSName(agent *entity.Agent) string {
	return fmt.Sprintf("%s.%s", agent.DropletName, agent.Region)
}

func defaultDiscoveryURL() string {
	resp, err := http.Get(discoveryGeneratorURI)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return ""
	}

	return buf.String()
}
