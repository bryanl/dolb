package site

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/bryanl/dolb/dao"
	"github.com/bryanl/dolb/server"
)

type LBCreateHandler struct {
	bh *baseHandler
}

func (h *LBCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var u *dao.User
	if u = h.bh.currentUser(r); u == nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	name := r.FormValue("name")
	region := r.FormValue("region")
	sshKeys := r.FormValue("ssh_keys")

	sshKeys = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, sshKeys)
	keys := strings.Split(sshKeys, ",")

	bc := server.BootstrapConfig{
		DigitalOceanToken: u.AccessToken,
		Name:              name,
		Region:            region,
		SSHKeys:           keys,
	}

	lb, err := server.CreateLoadBalancer(bc, h.bh.config)
	if err != nil {
		http.Redirect(w, r, "/lb/new", 302)
		return
	}

	http.Redirect(w, r, "/lb/"+lb.ID, 302)
}
