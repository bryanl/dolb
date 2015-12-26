package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/bryanl/dolb/doa"
	"github.com/bryanl/dolb/mocks"
	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_isValidClusterName(t *testing.T) {
	cases := []struct {
		name    string
		isValid bool
	}{
		{name: "my-cluster", isValid: true},
		{name: "12345", isValid: true},
		{name: "!", isValid: false},
		{name: "", isValid: false},
	}

	for _, c := range cases {
		got := isValidClusterName(c.name)
		assert.Equal(t, c.isValid, got)
	}
}

func TestBoostrapConfig_HasSyslog(t *testing.T) {
	bc := &BootstrapConfig{
		Region:            "dev0",
		SSHKeys:           []string{"123456"},
		DigitalOceanToken: "token",
	}

	assert.False(t, bc.HasSyslog())

	bc.RemoteSyslog = &RemoteSyslog{
		Host: "example.com",
		Port: 515,
	}

	assert.True(t, bc.HasSyslog())
}

func TestNewClusterOps(t *testing.T) {
	co := NewClusterOps()
	assert.NotNil(t, co)
}

type withMockGodoClusterOpts func(*clusterOps, *godoMocks)
type godoMocks struct {
	Droplets *mocks.DropletsService
}

type dropletOnboardMock struct{}

func (dom *dropletOnboardMock) setup() {}

func withMockGodo(fn withMockGodoClusterOpts) {
	co := &clusterOps{}
	co.DiscoveryGenerator = func() (string, error) {
		return "http://example.com/token", nil
	}

	gc := &godo.Client{}
	ds := &mocks.DropletsService{}
	gc.Droplets = ds
	co.GodoClientFactory = func(string) *godo.Client {
		return gc
	}

	co.DropletOnboardFactory = func(godo.Droplet, string, *godo.Client, *Config) DropletOnboard {
		return &dropletOnboardMock{}
	}

	gm := &godoMocks{
		Droplets: ds,
	}

	fn(co, gm)
}

func TestBootstrap(t *testing.T) {
	withMockGodo(func(co *clusterOps, gm *godoMocks) {
		droplet := godo.Droplet{}
		gm.Droplets.On("Create", mock.Anything).Return(&droplet, nil, nil)

		bc := &BootstrapConfig{
			Name:              "test-cluster",
			Region:            "dev0",
			SSHKeys:           []string{"123456"},
			DigitalOceanToken: "token",
		}

		sessionMock := &doa.MockSession{}

		members := []doa.LoadBalancerMember{
			{ID: "1", ClusterID: "12345", Name: "lb-test-cluster-1"},
			{ID: "2", ClusterID: "12345", Name: "lb-test-cluster-2"},
			{ID: "3", ClusterID: "12345", Name: "lb-test-cluster-3"},
		}

		for _, m := range members {
			cmr := &doa.CreateMemberRequest{ClusterID: m.ClusterID, Name: m.Name}
			sessionMock.On("CreateLBMember", cmr).Return(&m, nil).Once()
		}

		lb := &doa.LoadBalancer{ID: "12345"}

		config := &Config{
			ServerURL: "http://example.com",
			DBSession: sessionMock,
		}

		bo := &BootstrapOptions{
			LoadBalancer:    lb,
			BootstrapConfig: bc,
			Config:          config,
		}
		err := co.Bootstrap(bo)
		assert.NoError(t, err)
	})
}

func TestBootstrap_MissingName(t *testing.T) {
	withMockGodo(func(co *clusterOps, gm *godoMocks) {
		bc := &BootstrapConfig{
			Region:            "dev0",
			SSHKeys:           []string{"123456"},
			DigitalOceanToken: "token",
		}

		lb := &doa.LoadBalancer{}

		config := &Config{
			ServerURL: "http://example.com",
		}

		bo := &BootstrapOptions{
			LoadBalancer:    lb,
			BootstrapConfig: bc,
			Config:          config,
		}
		err := co.Bootstrap(bo)
		assert.Error(t, err)
	})
}

func Test_discoveryGenerator(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "http://example.com/token")
	}))
	defer ts.Close()

	defer func(u string) { discoveryGeneratorURI = u }(discoveryGeneratorURI)
	discoveryGeneratorURI = ts.URL

	uri, err := discoveryGenerator()
	assert.NoError(t, err)

	assert.Equal(t, "http://example.com/token", uri)
}

func TestUserData(t *testing.T) {
	defer func(udt string) { userDataTemplate = udt }(userDataTemplate)

	mT := map[string]string{
		"token":          "{{.CoreosToken}}",
		"region":         "{{.BootstrapConfig.Region}}",
		"do_token":       "{{.BootstrapConfig.DigitalOceanToken}}",
		"cluster_config": "{{.BootstrapConfig.Name}}",
		"server_url":     "{{.ServerURL}}",
		"log.host":       "{{.BootstrapConfig.RemoteSyslog.Host}}",
		"log.port":       "{{.BootstrapConfig.RemoteSyslog.Port}}",
	}

	b, err := json.Marshal(&mT)
	assert.NoError(t, err)
	userDataTemplate = string(b)

	token := "12345"
	agentID := "agent-2"

	bc := &BootstrapConfig{
		Region: "dev0",
		RemoteSyslog: &RemoteSyslog{
			EnableSSL: true,
			Host:      "host",
			Port:      515,
		},
	}

	config := &Config{
		ServerURL: "http://example.com",
	}

	lb := &doa.LoadBalancer{ID: "lb-1"}
	bo := &BootstrapOptions{
		LoadBalancer:    lb,
		Config:          config,
		BootstrapConfig: bc,
	}

	co := &clusterOps{}

	userData, err := co.userData(token, agentID, bo)
	assert.NoError(t, err)

	fmt.Println(userData)

	var m map[string]interface{}
	err = json.Unmarshal([]byte(userData), &m)
	assert.NoError(t, err)
	assert.Equal(t, token, m["token"])
	assert.Equal(t, bc.Region, m["region"])
	assert.Equal(t, bc.RemoteSyslog.Host, m["log.host"])
	assert.Equal(t, strconv.Itoa(bc.RemoteSyslog.Port), m["log.port"])
}

func Test_generateInstanceID(t *testing.T) {
	id := generateInstanceID()
	assert.Equal(t, 10, len(id))
}
