package server

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_LBCreateHandler(t *testing.T) {
	clusterOpsMock := &MockClusterOps{}
	clusterOpsMock.On("Bootstrap", mock.AnythingOfTypeArgument("*server.BootstrapOptions")).Return(nil)

	sess := &dao.MockSession{}
	lb := &dao.LoadBalancer{ID: "12345", Name: "lb-1"}

	sess.On("CreateLoadBalancer", "lb-1", "dev0", "do-token", mock.AnythingOfTypeArgument("*logrus.Entry")).Return(lb, nil)

	config := &Config{
		ClusterOpsFactory: func() ClusterOps {
			return clusterOpsMock
		},
		DBSession: sess,
		logger:    logrus.WithField("test", "test"),
	}

	body := []byte(`{"name": "lb-1", "region": "dev0", "ssh_keys": ["12345"], "digitalocean_token": "do-token"}`)
	r, err := http.NewRequest("POST", "http://example.com/lb", bytes.NewBuffer(body))
	assert.NoError(t, err)

	resp := LBCreateHandler(config, r)

	assert.Equal(t, http.StatusCreated, resp.Status)

	bcr := resp.Body.(BootstrapClusterResponse)
	assert.Equal(t, "12345", bcr.LoadBalancer.ID)
}

func Test_LBCreateHandler_no_token(t *testing.T) {
	clusterOpsMock := &MockClusterOps{}
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
