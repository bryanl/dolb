package do

import (
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

// TokenSource holds an oauth token.
type TokenSource struct {
	AccessToken string
}

// Token returns an oauth token.
func (t *TokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}

type GodoClientFactoryFn func(string) *godo.Client

func GodoClientFactory(token string) *godo.Client {
	ts := &TokenSource{AccessToken: token}
	oc := oauth2.NewClient(oauth2.NoContext, ts)
	return godo.NewClient(oc)
}
