package dolb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanl/dolb/mocks"
	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTokenSource(t *testing.T) {
	ts := TokenSource{AccessToken: "token"}
	_, err := ts.Token()
	assert.NoError(t, err)
}

func TestBoot(t *testing.T) {
	co := NewClusterOps()
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

	bc := &BootConfig{
		Region:  "dev0",
		SSHKeys: []string{"123456"},
		Token:   "token",
	}

	uri, err := co.Boot(bc)
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

	userDataTemplate = `{"token":"{{.Token}}", "region":"{{.Region}}"}`

	token := "12345"
	region := "dev0"

	cm := NewClusterOps()
	userData, err := cm.UserData(token, region)
	assert.NoError(t, err)

	fmt.Println(userData)

	var m map[string]interface{}
	err = json.Unmarshal([]byte(userData), &m)
	assert.NoError(t, err)
	assert.Equal(t, token, m["token"])
	assert.Equal(t, region, m["region"])
}
