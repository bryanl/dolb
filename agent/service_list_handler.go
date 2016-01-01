package agent

import (
	"net/http"

	"github.com/bryanl/dolb/service"
)

func ServiceListHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	sm := config.ServiceManagerFactory(config)
	services, err := sm.Services()
	if err != nil {
		config.GetLogger().WithError(err).Error("could not retrieve services")
		return service.Response{Body: err, Status: 400}
	}

	srs := &service.ServicesResponse{
		Services: []service.ServiceResponse{},
	}
	for _, s := range services {
		sr := convertServiceToResponse(s)
		srs.Services = append(srs.Services, sr)
	}

	return service.Response{Body: srs, Status: http.StatusOK}
}
