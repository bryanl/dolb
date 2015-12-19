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

func TestBoostrapConfig_HasSyslog(t *testing.T) {
	bc := &BootstrapConfig{
		Region:  "dev0",
		SSHKeys: []string{"123456"},
		Token:   "token",
	}

	assert.False(t, bc.HasSyslog())

	bc.RemoteSyslog = &RemoteSyslog{
		Host: "example.com",
		Port: 515,
	}

	assert.True(t, bc.HasSyslog())
}

func TestTokenSource(t *testing.T) {
	ts := TokenSource{AccessToken: "token"}
	_, err := ts.Token()
	assert.NoError(t, err)
}

func TestNewClusterOps(t *testing.T) {
	co := NewClusterOps()
	assert.NotNil(t, co)
}

func TestBoot(t *testing.T) {
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

	droplet := godo.Droplet{}
	droplets := []godo.Droplet{droplet}
	resp := &godo.Response{
		Links: &godo.Links{
			Actions: []godo.LinkAction{
				godo.LinkAction{HREF: "http://example.com/actions/1234", Rel: "multiple_create"},
			},
		},
	}

	ds.On("CreateMultiple", mock.Anything).Return(droplets, resp, nil)

	bc := &BootstrapConfig{
		Region:  "dev0",
		SSHKeys: []string{"123456"},
		Token:   "token",
	}

	uri, err := co.Bootstrap(bc)
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/actions/1234", uri)
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
		"token":    "{{.Token}}",
		"region":   "{{.BootstrapConfig.Region}}",
		"log.host": "{{.BootstrapConfig.RemoteSyslog.Host}}",
		"log.port": "{{.BootstrapConfig.RemoteSyslog.Port}}",
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
	userData, err := co.userData(token, bc)
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

func Test_godoClientFactory(t *testing.T) {
	gc := godoClientFactory("test-token")
	assert.NotNil(t, gc)
}

func Test_generateInstanceID(t *testing.T) {
	id := generateInstanceID()
	assert.Equal(t, 10, len(id))
}
