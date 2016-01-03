package site

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
)

type HomeHandler struct {
	bh *baseHandler
}

type homeData struct {
	UnknownUser bool
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := h.bh.currentUser(r)

	d := &homeData{
		UnknownUser: u == nil,
	}

	err := renderTemplate(w, "home.tmpl", d)
	if err != nil {
		logrus.WithError(err).Error("could not render template")
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
	}
}
