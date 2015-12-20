package agent

import (
	"net/http"

	"github.com/bryanl/dolb/service"
)

// RootResponse is a root response.
type RootResponse struct {
	ClusterStatus ClusterStatus `json:"cluster_status"`
}

// RootHandler is a handler for /.
func RootHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	rr := &RootResponse{
		ClusterStatus: config.ClusterStatus,
	}

	return service.Response{Body: rr, Status: http.StatusOK}
}
