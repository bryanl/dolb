package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

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

	gm := &godoMocks{
		Droplets: ds,
	}

	fn(co, gm)
}

func TestBootstrap(t *testing.T) {
	withMockGodo(func(co *clusterOps, gm *godoMocks) {
		droplet := godo.Droplet{}
		droplets := []godo.Droplet{droplet}
		resp := &godo.Response{
			Links: &godo.Links{
				Actions: []godo.LinkAction{
					godo.LinkAction{HREF: "http://example.com/actions/1234", Rel: "multiple_create"},
				},
			},
		}

		gm.Droplets.On("CreateMultiple", mock.Anything).Return(droplets, resp, nil)

		bc := &BootstrapConfig{
			Name:              "test-cluster",
			Region:            "dev0",
			SSHKeys:           []string{"123456"},
			DigitalOceanToken: "token",
		}

		su := "http://example.com"

		uri, err := co.Bootstrap(bc, su)
		assert.NoError(t, err)
		assert.Equal(t, "http://example.com/actions/1234", uri)
	})
}

func TestBootstrap_MissingName(t *testing.T) {
	withMockGodo(func(co *clusterOps, gm *godoMocks) {
		bc := &BootstrapConfig{
			Region:            "dev0",
			SSHKeys:           []string{"123456"},
			DigitalOceanToken: "token",
		}

		su := "http://example.com"
		_, err := co.Bootstrap(bc, su)
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
		"token":      "{{.CoreosToken}}",
		"region":     "{{.BootstrapConfig.Region}}",
		"do_token":   "{{.BootstrapConfig.DigitalOceanToken}}",
		"server_url": "{{.ServerURL}}",
		"log.host":   "{{.BootstrapConfig.RemoteSyslog.Host}}",
		"log.port":   "{{.BootstrapConfig.RemoteSyslog.Port}}",
	}

	b, err := json.Marshal(&mT)
	assert.NoError(t, err)
	userDataTemplate = string(b)

	token := "12345"

	bc := &BootstrapConfig{
		Region: "dev0",
		RemoteSyslog: &RemoteSyslog{
			EnableSSL: true,
			Host:      "host",
			Port:      515,
		},
	}

	co := &clusterOps{}
	su := "http://example.com"
	userData, err := co.userData(token, su, bc)
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
