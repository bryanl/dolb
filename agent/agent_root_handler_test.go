package agent

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AgentRootHandler(t *testing.T) {
	config := &Config{Leader: "theboss"}

	r, err := http.NewRequest("GET", "http://example.com/", nil)
	assert.NoError(t, err)

	resp := AgentRootHandler(config, r)

	assert.Equal(t, http.StatusOK, resp.Status)

	rr := resp.Body.(*RootResponse)
	assert.Equal(t, "theboss", rr.Leader)
}
