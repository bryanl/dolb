package site

import (
	"fmt"
	"html/template"
	"net/http"
)

type HomeHandler struct {
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

	args := map[string]string{}
	t.Execute(w, args)
}
