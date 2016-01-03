package site

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
)

type LBNewHandler struct {
	bh *baseHandler
}

func (h *LBNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var u *dao.User
	if u = h.bh.currentUser(r); u == nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	err := renderTemplate(w, "lb_new.tmpl", nil)
	if err != nil {
		logrus.WithError(err).Error("could not render template")
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
	}

}
