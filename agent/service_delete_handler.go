package agent

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

func ServiceDeleteHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	vars := mux.Vars(r)
	svcName := vars["service"]

	sm := config.ServiceManagerFactory(config)
	err := sm.DeleteService(svcName)

	if err != nil {
		config.GetLogger().WithError(err).WithFields(logrus.Fields{
			"service-name": svcName,
		}).Error("could not delete service")
		return service.Response{Body: err, Status: 404}
	}

	return service.Response{Status: 204}

}
