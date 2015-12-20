package agent

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RootHandler(t *testing.T) {
	config := &Config{
		ClusterStatus: ClusterStatus{
			Leader:    "theboss",
			NodeCount: 2,
		},
	}

	r, err := http.NewRequest("GET", "http://example.com/", nil)
	assert.NoError(t, err)

	resp := RootHandler(config, r)

	assert.Equal(t, http.StatusOK, resp.Status)

	rr := resp.Body.(*RootResponse)
	assert.Equal(t, "theboss", rr.ClusterStatus.Leader)
	assert.Equal(t, 2, rr.ClusterStatus.NodeCount)
}
