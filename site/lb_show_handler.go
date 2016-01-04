package site

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/gorilla/mux"
)

type LBShowHandler struct {
	bh *baseHandler
}

func (h *LBShowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var u *dao.User
	if u = h.bh.currentUser(r); u == nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	vars := mux.Vars(r)
	lbID := vars["lb_id"]

	data, err := h.buildView(lbID)
	if err != nil {
		logrus.WithError(err).Error("could not load load balancer")
		w.WriteHeader(404)
		fmt.Fprintln(w, "404 page not found")
		return
	}

	err = renderTemplate(w, "lb_show.tmpl", data)
	if err != nil {
		logrus.WithError(err).Error("could not render template")
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
	}
}

func (h *LBShowHandler) buildView(id string) (map[string]interface{}, error) {
	lb, err := h.bh.config.DBSession.LoadLoadBalancer(id)
	if err != nil {
		return nil, err
	}

	agents, err := h.bh.config.DBSession.LoadBalancerAgents(id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"LoadBalancer": lb,
		"Agents":       agents,
	}, nil
}
