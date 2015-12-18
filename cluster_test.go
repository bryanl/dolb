package dolb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateDiscoveryURI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "http://example.com/token")
	}))
	defer ts.Close()

	cm := NewClusterOps()
	cm.DiscoveryGeneratorURL = ts.URL

	uri, err := cm.DiscoveryURI()
	assert.NoError(t, err)

	assert.Equal(t, "http://example.com/token", uri)
}

func TestGenerateDiscoveryURI_ExternalError(t *testing.T) {
	cm := NewClusterOps()
	cm.DiscoveryGeneratorURL = ""

	_, err := cm.DiscoveryURI()
	assert.Error(t, err)
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
