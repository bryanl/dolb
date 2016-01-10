package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

func ServiceListHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	vars := mux.Vars(r)
	lbID := vars["lb_id"]

	lb, err := config.DBSession.LoadLoadBalancer(lbID)
	if err != nil {
		return service.Response{Body: "not found", Status: 404}
	}

	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", lb.FloatingIp, 8889),
		Path:   "/services",
	}

	defer r.Body.Close()

	resp, err := http.Get(u.String())
	if err != nil {
		return service.Response{Body: "cannot contact agent", Status: 500}
	}

	var sr service.ServicesResponse
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		return service.Response{Body: "cannot read agent response", Status: 500}
	}

	return service.Response{Body: sr, Status: resp.StatusCode}
}
