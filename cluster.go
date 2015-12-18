package dolb

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
)

var (
	coreosImage           = "coreos-stable"
	discoveryGeneratorURI = "http://discovery.etcd.io/new?size=3"
	dropletSize           = "512mb"
)

type userDataConfig struct {
	Token  string
	Region string
}

// BootstrapConfig is configuration for Bootstrap.
type BootstrapConfig struct {
	Region  string
	SSHKeys []string
	Token   string
}

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

// ClusterOps is an interface for cluster operations.
type ClusterOps interface {
	Bootstrap(bc *BootstrapConfig) (string, error)
}

// clusterOps are operations for building clusters.
type clusterOps struct {
	DiscoveryGenerator func() (string, error)
	GodoClientFactory  func(string) *godo.Client
}

var _ ClusterOps = &clusterOps{}

// NewClusterOps creates an instance of clusterOps.
func NewClusterOps() ClusterOps {
	return &clusterOps{
		DiscoveryGenerator: discoveryGenerator,
		GodoClientFactory:  godoClientFactory,
	}
}

// Bootstrap bootstraps the cluster and returns a tracking URI or error.
func (co *clusterOps) Bootstrap(bc *BootstrapConfig) (string, error) {
	names := make([]string, 3)
	id := generateInstanceID()
	for i := 0; i < 3; i++ {
		names[i] = fmt.Sprintf("lb-node-%s", id)
	}

	keys := []godo.DropletCreateSSHKey{}
	for _, k := range bc.SSHKeys {
		i, err := strconv.Atoi(k)
		if err != nil {
			return "", err
		}
		keys = append(keys, godo.DropletCreateSSHKey{ID: i})
	}

	du, err := co.DiscoveryGenerator()
	if err != nil {
		return "", err
	}

	userData, err := co.userData(du, bc.Region)
	if err != nil {
		return "", err
	}

	dmcr := godo.DropletMultiCreateRequest{
		Names:             names,
		Region:            bc.Region,
		Image:             godo.DropletCreateImage{Slug: coreosImage},
		Size:              dropletSize,
		SSHKeys:           keys,
		PrivateNetworking: true,
		UserData:          userData,
	}

	godoc := co.GodoClientFactory(bc.Token)
	_, resp, err := godoc.Droplets.CreateMultiple(&dmcr)
	if err != nil {
		return "", err
	}

	action := findAction("multiple_create", resp.Links.Actions)
	if action == "" {
		return "", errors.New("no multiple_create action found")
	}

	return action, nil
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

// UserData creates a cloud config.
func (co *clusterOps) userData(token, region string) (string, error) {
	t, err := template.New("user-data").Parse(userDataTemplate)
	if err != nil {
		return "", err
	}

	udc := &userDataConfig{
		Token:  token,
		Region: region,
	}

	var b bytes.Buffer

	err = t.Execute(&b, udc)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func godoClientFactory(token string) *godo.Client {
	ts := &TokenSource{AccessToken: token}
	oc := oauth2.NewClient(oauth2.NoContext, ts)
	return godo.NewClient(oc)
}

func findAction(rel string, actions []godo.LinkAction) string {
	for _, a := range actions {
		if a.Rel == rel {
			return a.HREF
		}
	}

	return ""
}

func generateInstanceID() string {
	strlen := 10
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

//go:generate embed file -var userDataTemplate --source user_data_template.yml
var userDataTemplate = "#cloud-config\n\ncoreos:\n  etcd2:\n    discovery: {{.Token}}\n    advertise-client-urls: http://$private_ipv4:2379,http://$private_ipv4:4001\n    initial-advertise-peer-urls: http://$private_ipv4:2380\n    listen-client-urls: http://0.0.0.0:2379,http://0.0.0.0:4001\n    listen-peer-urls: http://$private_ipv4:2380\n  fleet:\n    public-ip: $private_ipv4\n    metadata: region={{.Region}},public_ip=$public_ipv4\n  units:\n    - name: etcd2.service\n      command: start\n    - name: fleet.service\n      command: start\n"
