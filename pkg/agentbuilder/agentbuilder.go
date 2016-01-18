package agentbulder

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/bryanl/dolb/entity"
	"github.com/bryanl/dolb/pkg/agentuserdata"
	"github.com/bryanl/dolb/pkg/app"
)

var (
	discoveryGeneratorURI = "http://discovery.etcd.io/new"
	dropletSize           = "512mb"
)

// AgentBuilder creates and configure agent droplets.
type AgentBuilder struct {
	DOClient         app.DOClient
	EntityManager    entity.Manager
	GenerateRandomID func() string
	GenerateUserData func(*agentuserdata.Config) (string, error)
	DiscoveryURL     func() string

	bootstrapConfig *app.BootstrapConfig
	lb              *entity.LoadBalancer
}

var _ app.AgentBuilder = &AgentBuilder{}

// New builds an AgentBuilder.
func New(lb *entity.LoadBalancer, bootstrapConfig *app.BootstrapConfig, options ...func(*AgentBuilder)) *AgentBuilder {
	agentBuilder := AgentBuilder{
		bootstrapConfig: bootstrapConfig,
		lb:              lb,
	}

	for _, option := range options {
		option(&agentBuilder)
	}

	if agentBuilder.DiscoveryURL == nil {
		agentBuilder.DiscoveryURL = defaultDiscoveryURL
	}

	if agentBuilder.GenerateUserData == nil {
		agentBuilder.GenerateUserData = func(c *agentuserdata.Config) (string, error) {
			return agentuserdata.Generate(c)
		}
	}

	return &agentBuilder
}

// DOClient sets the DOClient for an AgentBuilder.
func DOClient(doClient app.DOClient) func(*AgentBuilder) {
	return func(ab *AgentBuilder) {
		ab.DOClient = doClient
	}
}

// EntityManager sets the EntityManager for an AgentBuilder.
func EntityManager(em entity.Manager) func(*AgentBuilder) {
	return func(ab *AgentBuilder) {
		ab.EntityManager = em
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

// DiscoveryURL sets DiscoveryURL for an AgentBuilder.
func DiscoveryURL(fn func() string) func(*AgentBuilder) {
	return func(ab *AgentBuilder) {
		ab.DiscoveryURL = fn
	}
}

// Create Creates a droplet in the database.
func (ab *AgentBuilder) Create(id int) (*entity.Agent, error) {
	name := fmt.Sprintf("agent-%s-%d", ab.lb.ID, id)

	agent := &entity.Agent{
		ID:          ab.GenerateRandomID(),
		ClusterID:   ab.lb.ID,
		DropletName: name,
		Region:      ab.lb.Region,
	}

	if err := ab.EntityManager.Create(agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// Configure builds and configures an agent droplet.
func (ab *AgentBuilder) Configure(agent *entity.Agent) error {
	userDataConfig := &agentuserdata.Config{
		AgentVersion:    "0.0.2", // TODO where is this coming from?
		AgentID:         agent.ID,
		BootstrapConfig: ab.bootstrapConfig,
		ClusterID:       agent.ClusterID,
		CoreosToken:     ab.DiscoveryURL(),
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

	agentCreateResponse, err := ab.DOClient.CreateAgent(agentCreateRequest)
	if err != nil {
		return err
	}
	agent.DropletID = agentCreateResponse.DropletID

	dnsName := ab.agentDNSName(agent)
	dnsEntry, err := ab.DOClient.CreateDNS(dnsName, agentCreateResponse.PublicIPAddress)
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
