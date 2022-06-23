package server

import "net/http"

type router struct{}

func NewRouter() *router {
	return &router{}
}

func (r *router) Serve(w http.ResponseWriter, req *http.Request) {
}
