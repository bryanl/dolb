package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type API2 struct {
	Mux http.Handler
}

func NewAPI2() *API2 {
	m := mux.NewRouter()

	a := &API2{
		Mux: m,
	}

	return a
}
