package agent

import (
	"net/http"

	"github.com/bryanl/dolb/service"
)

type RootResponse struct {
	Leader string `json:"leader"`
}

func AgentRootHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	rr := &RootResponse{
		Leader: config.Leader,
	}

	return service.Response{Body: rr, Status: http.StatusOK}
}
