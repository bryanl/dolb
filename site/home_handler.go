package site

import (
	"fmt"
	"html/template"
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

	userID := session.Values["user_id"].(string)

	tmpl, err := Asset("templates/home.html")
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
	}

	t, err := template.New("home").Parse(string(tmpl))
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
	}

	d := &homeData{
		UnknownUser: userID == "",
	}

	err = t.Execute(w, d)
	if err != nil {
		logrus.WithError(err).Error("could not render home template")
	}
}
