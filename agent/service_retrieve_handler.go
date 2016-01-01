package agent

import (
	"net/http"

	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

func ServiceRetrieveHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	vars := mux.Vars(r)
	svcName := vars["service"]

	sm := config.ServiceManagerFactory(config)
	s, err := sm.Service(svcName)
	if err != nil {
		config.GetLogger().WithError(err).WithField("service-name", svcName).Error("could not retrieve service")
		return service.Response{Body: err, Status: 404}
	}
	sr := convertServiceToResponse(s)

	return service.Response{Body: sr, Status: http.StatusOK}
}
