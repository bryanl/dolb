package do

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

var (
	coreosImage   = "coreos-beta"
	ActionTimeout = 5 * time.Second
)

// TokenSource holds an oauth token.
type TokenSource struct {
	AccessToken string
}

// Token returns an oauth token.
func (t *TokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}

type GodoClientFactoryFn func(string) *godo.Client

func GodoClientFactory(token string) *godo.Client {
	ts := &TokenSource{AccessToken: token}
	oc := oauth2.NewClient(oauth2.NoContext, ts)
	return godo.NewClient(oc)
}

type DigitalOcean interface {
	CreateAgent(*DropletCreateRequest) (*Agent, error)
	DeleteAgent(id int) error

	CreateDNS(name, ipAddress string) (*DNSEntry, error)
	DeleteDNS(id int) error

	CreateFloatingIP(region string) (*FloatingIP, error)
	DeleteFloatingIP() error
	AssignFloatingIP(agentID, floatingIP string) error
	UnassignFloatingIP(floatingIP string) error
}

type DropletCreateRequest struct {
	Name     string
	Region   string
	Size     string
	SSHKeys  []string
	UserData string
}

type Agent struct {
	DropletID   int
	IPAddresses IPAddresses
}

type IPAddresses map[string]string

type DNSEntry struct {
	RecordID int
	Domain   string
	Name     string
	Type     string
	IP       string
}

type FloatingIP struct {
}

type LiveDigitalOcean struct {
	Client     *godo.Client
	BaseDomain string
}

var _ DigitalOcean = &LiveDigitalOcean{}

func NewLiveDigitalOcean(client *godo.Client, baseDomain string) *LiveDigitalOcean {
	return &LiveDigitalOcean{
		Client:     client,
		BaseDomain: baseDomain,
	}
}

func (ldo *LiveDigitalOcean) CreateAgent(dcr *DropletCreateRequest) (*Agent, error) {
	keys := []godo.DropletCreateSSHKey{}
	for _, k := range dcr.SSHKeys {
		i, err := strconv.Atoi(k)
		if err != nil {
			return nil, err
		}
		keys = append(keys, godo.DropletCreateSSHKey{ID: i})
	}

	gdcr := godo.DropletCreateRequest{
		Name:              dcr.Name,
		Region:            dcr.Region,
		Image:             godo.DropletCreateImage{Slug: coreosImage},
		Size:              dcr.Size,
		PrivateNetworking: true,
		SSHKeys:           keys,
		UserData:          dcr.UserData,
	}

	droplet, _, err := ldo.Client.Droplets.Create(&gdcr)
	if err != nil {
		return nil, err
	}

	a, err := ldo.findAction(droplet, "create")
	if err != nil {
		return nil, err
	}

	err = ldo.waitForAction(a)
	if err != nil {
		return nil, err
	}

	droplet, _, err = ldo.Client.Droplets.Get(droplet.ID)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		DropletID:   droplet.ID,
		IPAddresses: IPAddresses{},
	}

	for _, n := range droplet.Networks.V4 {
		agent.IPAddresses[n.Type] = n.IPAddress
	}

	return agent, nil
}

func (ldo *LiveDigitalOcean) findAction(droplet *godo.Droplet, actionType string) (*godo.Action, error) {
	actions, _, err := ldo.Client.Droplets.Actions(droplet.ID, nil)
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

func (ldo *LiveDigitalOcean) waitForAction(action *godo.Action) error {
	for {
		a, _, err := ldo.Client.Actions.Get(action.ID)
		if err != nil {
			return err
		}

		switch a.Status {
		case "completed":
			return nil
		case "errored":
			return errors.New("action errored")
		case "in-progress":
			time.Sleep(ActionTimeout)
			continue
		default:
			return fmt.Errorf("unknown action status: %v")
		}

	}
}

func (ldo *LiveDigitalOcean) DeleteAgent(id int) error {
	return nil
}

func (ldo *LiveDigitalOcean) CreateDNS(name, ipAddress string) (*DNSEntry, error) {

	drer := &godo.DomainRecordEditRequest{
		Type: "A",
		Name: name,
		Data: ipAddress,
	}

	r, _, err := ldo.Client.Domains.CreateRecord(ldo.BaseDomain, drer)
	if err != nil {
		return nil, err
	}

	return &DNSEntry{
		RecordID: r.ID,
		Domain:   ldo.BaseDomain,
		Name:     name,
		Type:     "A",
		IP:       r.Data,
	}, nil
}

func (ldo *LiveDigitalOcean) DeleteDNS(id int) error {
	return nil
}

func (ldo *LiveDigitalOcean) CreateFloatingIP(region string) (*FloatingIP, error) {
	return nil, nil
}

func (ldo *LiveDigitalOcean) DeleteFloatingIP() error {
	return nil
}

func (ldo *LiveDigitalOcean) AssignFloatingIP(agentID, floatingIP string) error {
	return nil
}

func (ldo *LiveDigitalOcean) UnassignFloatingIP(floatingIP string) error {
	return nil
}
