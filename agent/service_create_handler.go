package agent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bryanl/dolb/service"
)

func ServiceCreateHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)
	defer r.Body.Close()

	var ereq service.ServiceCreateRequest
	err := json.NewDecoder(r.Body).Decode(&ereq)
	if err != nil {
		return service.Response{Body: fmt.Errorf("could not decode json: %v", err), Status: 422}
	}

	config.GetLogger().WithField("ereq", fmt.Sprintf("%#v", ereq)).Info("service create request")

	sm := config.ServiceManagerFactory(config)
	err = sm.Create(ereq)
	if err != nil {
		config.GetLogger().WithError(err).Error("could not create service")
		return service.Response{Body: err, Status: 400}
	}

	resp := service.ServiceCreateResponse{
		Name:   ereq.Name,
		Port:   ereq.Port,
		Domain: ereq.Domain,
		Regex:  ereq.Regex,
	}

	return service.Response{Body: resp, Status: http.StatusCreated}
}
