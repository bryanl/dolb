package site

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bryanl/dolb/vendor/github.com/markbates/goth/gothic"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
)

var (
	sessionName = "_gothic_session"
	store       = sessions.NewCookieStore([]byte("secret"))
)

func beginGoth(w http.ResponseWriter, r *http.Request) {
	url, err := getAuthURL(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func gothCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("State: ", gothic.GetState(r))

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	// FIXME save or update user here

	fmt.Printf("found user: %#v\n", user)
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

	session, err := store.Get(r, sessionName)
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

	session, err := store.Get(req, sessionName)
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
