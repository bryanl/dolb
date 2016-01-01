package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

type UpstreamCreateRequest struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func UpstreamCreateHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)
	defer r.Body.Close()

	vars := mux.Vars(r)
	svcName := vars["service"]

	var ucr UpstreamCreateRequest
	err := json.NewDecoder(r.Body).Decode(&ucr)
	if err != nil {
		return service.Response{Body: fmt.Errorf("could not decode json: %v", err), Status: 422}
	}

	sm := config.ServiceManagerFactory(config)
	err = sm.AddUpstream(svcName, ucr)
	if err != nil {
		return service.Response{Body: err, Status: 400}
	}

	svc, err := sm.Service(svcName)
	if err != nil {
		return service.Response{Body: err, Status: 400}
	}

	return service.Response{Body: svc, Status: http.StatusOK}
}
