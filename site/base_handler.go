package site

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/server"
)

type baseHandler struct {
	config *server.Config
}

func (bh *baseHandler) currentUser(r *http.Request) *dao.User {
	session, err := sessionStore.Get(r, "_dolb_session")
	if err != nil {
		logrus.WithError(err).Error("unable to load session")
		return nil
	}

	if userID, ok := session.Values["user_id"]; ok {
		u, err := bh.config.DBSession.FindUser(userID.(string))
		if err != nil {
			return nil
		}
		return u
	}
	return nil
}
