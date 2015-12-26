package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterHandler(t *testing.T) {
	mockSession := &dao.MockSession{}
	mockSession.On("UpdateAgent", mock.AnythingOfTypeArgument("*dao.UpdateAgentRequest")).Return(nil)

	lb := &dao.LoadBalancer{ID: "cluster-1"}
	mockSession.On("RetrieveLoadBalancer", "cluster-1").Return(lb, nil)

	c := &Config{
		logger:    logrus.WithField("test", "test"),
		DBSession: mockSession,
	}

	rReq := PingRequest{
		ClusterID:   "cluster-1",
		ClusterName: "cluster",
		Host:        "host-1",
	}

	u := "http://example.com/ping"

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(rReq)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", u, &b)
	resp := PingHandler(c, req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}
