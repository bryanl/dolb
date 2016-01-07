package server

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/service"
	"github.com/gorilla/sessions"
)

var (
	sessionStore = sessions.NewCookieStore([]byte("secret"))
)

func UserRetrieveHandler(c interface{}, r *http.Request) service.Response {
	config := c.(*Config)

	session, err := sessionStore.Get(r, "_dolb_session")
	if err != nil {
		logrus.WithError(err).Error("unable to load session")
		return service.Response{Body: "unknown user", Status: 401}
	}

	userID, ok := session.Values["user_id"]
	if !ok {
		return service.Response{Body: "unknown user", Status: 401}
	}

	u, err := config.DBSession.FindUser(userID.(string))
	if err != nil {
		return service.Response{Body: "unknown user", Status: 401}
	}

	uir := service.UserInfoResponse{
		UserID:      u.ID,
		Email:       u.Email,
		AccessToken: u.AccessToken,
	}

	return service.Response{Body: uir, Status: 200}
}
