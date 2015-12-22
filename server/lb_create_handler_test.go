package server

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_LBCreateHandler(t *testing.T) {
	clusterOpsMock := &ClusterOpsMock{}
	clusterOpsMock.On("Bootstrap", mock.Anything, mock.Anything).Return("http://example.com/action/12345", nil)

	config := &Config{
		ClusterOpsFactory: func() ClusterOps {
			return clusterOpsMock
		},
	}

	body := []byte(`{"name": "lb-1", "region": "dev0", "ssh_keys": ["12345"], "digitalocean_token": "do-token"}`)
	r, err := http.NewRequest("POST", "http://example.com/lb", bytes.NewBuffer(body))
	assert.NoError(t, err)

	resp := LBCreateHandler(config, r)

	assert.Equal(t, http.StatusCreated, resp.Status)

	bcr := resp.Body.(BootstrapClusterResponse)
	assert.Equal(t, "lb-1", bcr.ID)
	assert.Equal(t, "http://example.com/action/12345", bcr.MonitorURI)
}

func Test_LBCreateHandler_no_token(t *testing.T) {
	clusterOpsMock := &ClusterOpsMock{}
	clusterOpsMock.On("Bootstrap", mock.Anything, mock.Anything).Return("http://example.com/action/12345", nil)

	config := &Config{
		ClusterOpsFactory: func() ClusterOps {
			return clusterOpsMock
		},
	}

	body := []byte(`{"name": "lb-1", "region": "dev0", "ssh_keys": ["12345"]}`)
	r, err := http.NewRequest("POST", "http://example.com/lb", bytes.NewBuffer(body))
	assert.NoError(t, err)

	resp := LBCreateHandler(config, r)

	assert.Equal(t, 400, resp.Status)
}
