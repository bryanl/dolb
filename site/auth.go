package site

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bryanl/dolb/dao"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

var (
	sessionName  = "_gothic_session"
	sessionStore = sessions.NewCookieStore([]byte("secret"))
)

type userInfo struct {
	UserID string
	Email  string
}

func beginGoth(w http.ResponseWriter, r *http.Request) {
	url, err := getAuthURL(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type OauthCallback struct {
	DBSession dao.Session
}

func (oc *OauthCallback) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("State: ", gothic.GetState(r))

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	u, err := oc.DBSession.FindUser(user.UserID)
	if err != nil {
		// NOTE user didn't exist in DB most likely
		u = oc.DBSession.NewUser()
		u.ID = user.UserID
		u.Email = user.Email
	}

	u.AccessToken = user.AccessToken

	ui := userInfo{
		UserID: u.ID,
		Email:  u.Email,
	}

	j, err := json.Marshal(ui)
	if err != nil {
		logrus.WithError(err).Error("unable to create session")
		w.WriteHeader(500)
		fmt.Fprint(w, "session is unavailable")
		return
	}

	cookie := &http.Cookie{
		Name:    "dolb_user_info",
		Value:   string(j),
		Expires: time.Now().Add(time.Hour * 24 * 30),
		Path:    "/",
	}

	http.SetCookie(w, cookie)

	session, err := sessionStore.Get(r, "_dolb_session")
	if err != nil {
		logrus.WithError(err).Error("unable to load session")
		w.WriteHeader(500)
		fmt.Fprint(w, "session is unavailable")
		return
	}

	err = u.Save()
	if err != nil {
		logrus.WithError(err).Error("unable to save user")
		session.AddFlash("Unable to save user")
	}

	session.Values["user_id"] = u.ID

	err = session.Save(r, w)
	if err != nil {
		logrus.WithError(err).Error("unable to save session")
	}

	http.Redirect(w, r, "/", 302)
}

func getAuthURL(w http.ResponseWriter, r *http.Request) (string, error) {
	provider, err := goth.GetProvider("digitalocean")
	if err != nil {
		return "", err
	}
	sess, err := provider.BeginAuth(gothic.SetState(r))
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return "", err
	}

	session.Values[sessionName] = sess.Marshal()
	err = session.Save(r, w)
	if err != nil {
		return "", err
	}

	return url, err

}

func completeUserAuth(res http.ResponseWriter, req *http.Request) (goth.User, error) {
	providerName := "digitalocean"

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return goth.User{}, err
	}

	session, err := sessionStore.Get(req, sessionName)
	if err != nil {
		return goth.User{}, err
	}

	if session.Values[sessionName] == nil {
		return goth.User{}, errors.New("could not find a matching session for this request")
	}

	sess, err := provider.UnmarshalSession(session.Values[sessionName].(string))
	if err != nil {
		return goth.User{}, err
	}

	_, err = sess.Authorize(provider, req.URL.Query())

	if err != nil {
		return goth.User{}, err
	}
	return provider.FetchUser(sess)
}
