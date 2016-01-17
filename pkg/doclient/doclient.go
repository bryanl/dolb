package doclient

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bryanl/dolb/pkg/app"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

var (
	// coreosImage is the agent image.
	coreosImage = "coreos-alpha"

	// actionTimeout is how long to wait before checking an action's status.
	actionTimeout = 5 * time.Second
)

// DOClient is a DigitalOcean client for DOLB. It wraps godo to provide higher level
// convience functions.
type DOClient struct {
	GodoClient *godo.Client
}

// New build a DOClient.
func New(token string, options ...func(*DOClient)) *DOClient {
	doClient := DOClient{}

	for _, option := range options {
		option(&doClient)
	}

	if doClient.GodoClient == nil {
		ts := &tokenSource{AccessToken: token}
		oc := oauth2.NewClient(oauth2.NoContext, ts)
		doClient.GodoClient = godo.NewClient(oc)
	}

	return &doClient
}

// GodoClient sets a DOClient instance's GodoClient.
func GodoClient(gc *godo.Client) func(*DOClient) {
	return func(dc *DOClient) {
		dc.GodoClient = gc
	}
}

var _ app.DOClient = &DOClient{}

// CreateAgent creates an agent. Returns the droplet's public IP address or an error
// if something unexpected occurred.
func (dc *DOClient) CreateAgent(acr *app.AgentCreateRequest) (*app.AgentCreateResponse, error) {
	keys := []godo.DropletCreateSSHKey{}
	for _, k := range acr.SSHKeys {
		i, err := strconv.Atoi(k)
		if err != nil {
			return nil, err
		}
		keys = append(keys, godo.DropletCreateSSHKey{ID: i})
	}

	agent := acr.Agent

	gdcr := godo.DropletCreateRequest{
		Name:              agent.DropletName,
		Region:            agent.Region,
		Image:             godo.DropletCreateImage{Slug: coreosImage},
		Size:              acr.Size,
		PrivateNetworking: true,
		SSHKeys:           keys,
		UserData:          acr.UserData,
	}

	droplet, _, err := dc.GodoClient.Droplets.Create(&gdcr)
	if err != nil {
		return nil, err
	}

	a, err := dc.findAction(droplet, "create")
	if err != nil {
		return nil, err
	}

	err = dc.waitForAction(a)
	if err != nil {
		return nil, err
	}

	droplet, _, err = dc.GodoClient.Droplets.Get(droplet.ID)
	if err != nil {
		return nil, err
	}

	resp := &app.AgentCreateResponse{
		DropletID: droplet.ID,
	}

	for _, n := range droplet.Networks.V4 {
		if n.Type == "public" {
			resp.PublicIPAddress = n.IPAddress
			return resp, nil
		}
	}

	return nil, fmt.Errorf("unable to find public ip for droplet")
}

// DeleteAgent deletes an agent.
func (dc *DOClient) DeleteAgent(id int) error {
	return nil
}

// CreateDNS assigns an a name to an ip in DNS.
func (dc *DOClient) CreateDNS(name, ipAddress string) (*app.DNSEntry, error) {
	return nil, nil
}

// DeleteDNS deletes a DNS entry.
func (dc *DOClient) DeleteDNS(id int) error {
	return nil
}

// CreateFloatingIP creates a floating ip in a region.
func (dc *DOClient) CreateFloatingIP(region string) (*app.FloatingIP, error) {
	return nil, nil
}

// DeleteFloatingIP deletes a floating ip.
func (dc *DOClient) DeleteFloatingIP(ipAddress string) error {
	return nil
}

// AssignFloatingIP assigns a floating ip to an agent.
func (dc *DOClient) AssignFloatingIP(agentID, floatingIP string) error {
	return nil
}

// UnassignFloatingIP removing a floating ip assign from an agent.
func (dc *DOClient) UnassignFloatingIP(floatingIP string) error {
	return nil
}

func (dc *DOClient) findAction(droplet *godo.Droplet, actionType string) (*godo.Action, error) {
	actions, _, err := dc.GodoClient.Droplets.Actions(droplet.ID, nil)
	if err != nil {
		return nil, err
	}

	for _, a := range actions {
		if a.Type == actionType {
			return &a, nil
		}
	}

	return nil, fmt.Errorf("could not find %q action for droplet", actionType)
}

func (dc *DOClient) waitForAction(action *godo.Action) error {
	for {
		a, _, err := dc.GodoClient.Actions.Get(action.ID)
		if err != nil {
			return err
		}

		switch a.Status {
		case "completed":
			return nil
		case "errored":
			return errors.New("action errored")
		case "in-progress":
			time.Sleep(actionTimeout)
			continue
		default:
			return fmt.Errorf("unknown action status: %v")
		}

	}
}

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}
