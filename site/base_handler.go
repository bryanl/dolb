package site

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
)

type baseHandler struct {
	DBSession dao.Session
}

func (bh *baseHandler) currentUser(r *http.Request) *dao.User {
	session, err := sessionStore.Get(r, "_dolb_session")
	if err != nil {
		logrus.WithError(err).Error("unable to load session")
		return nil
	}

	if userID, ok := session.Values["user_id"]; ok {
		u, err := bh.DBSession.FindUser(userID.(string))
		if err != nil {
			return nil
		}
		return u
	}
	return nil

}
