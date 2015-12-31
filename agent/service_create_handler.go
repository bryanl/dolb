package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bryanl/dolb/service"
)

type EndpointRequest struct {
	ServiceName string `json:"service_name"`
	Domain      string `json:"domain"`
	Regex       string `json:"url_regex"`
}

type EndpointResponse struct {
	Domain    string     `json:"domain"`
	Regex     string     `json:"url_regex"`
	Upstreams []Upstream `json:"upstreams"`
}

type Upstream struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func ServiceCreateHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)
	defer r.Body.Close()

	var ereq EndpointRequest
	err := json.NewDecoder(r.Body).Decode(&ereq)
	if err != nil {
		return service.Response{Body: fmt.Errorf("could not decode json: %v", err), Status: 422}
	}

	sm := config.ServiceManagerFactory(config)
	err = sm.Create(ereq)
	if err != nil {
		config.GetLogger().WithError(err).Error("could not create service")
		return service.Response{Body: err, Status: 400}
	}

	return service.Response{Body: config, Status: http.StatusCreated}
}
