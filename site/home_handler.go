package site

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
)

type HomeHandler struct {
	DBSession dao.Session
}

type homeData struct {
	UnknownUser bool
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "_dolb_session")
	if err != nil {
		logrus.WithError(err).Error("unable to load session")
		w.WriteHeader(500)
		fmt.Fprint(w, "session is unavailable")
		return
	}

	var userID string
	if id, ok := session.Values["user_id"]; ok {
		userID = id.(string)
	}

	d := &homeData{
		UnknownUser: userID == "",
	}

	err = renderTemplate(w, "home.tmpl", d)
	if err != nil {
		logrus.WithError(err).Error("could not render home template")
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
	}
}
