package agent

import (
	"net/http"
	"time"

	"github.com/bryanl/dolb/service"
)

var (
	fleetAPI = "http://127.0.0.1:49153"
)

func AgentReloadHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	// FIXME this is not the best way
	go func() {
		time.Sleep(5)
		config.GetLogger().Fatal("restart requested")
	}()

	return service.Response{Status: http.StatusNoContent}
}
