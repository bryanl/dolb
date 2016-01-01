package agent

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/mux"
)

func UpstreamDeleteHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	vars := mux.Vars(r)
	svcName := vars["service"]
	uName := vars["upstream"]

	sm := config.ServiceManagerFactory(config)
	err := sm.DeleteUpstream(svcName, uName)

	if err != nil {
		config.GetLogger().WithError(err).WithFields(logrus.Fields{
			"service-name": svcName,
			"upstream-id":  uName,
		}).Error("could not delete upstream")
		return service.Response{Body: err, Status: 404}
	}

	return service.Response{Status: 204}
}
