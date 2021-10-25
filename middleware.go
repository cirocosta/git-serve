package main

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Middleware func(http.Handler) http.Handler

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()

		log.WithFields(log.Fields{
			"method":   r.Method,
			"url":      r.URL.String(),
			"duration": t2.Sub(t1),
		}).Debug("req")
	})
}

func NewChain() *chain {
	return &chain{[]Middleware{}}
}

type chain struct {
	middlewares []Middleware
}

func (c *chain) Use(m Middleware) {
	c.middlewares = append(c.middlewares, m)
}

func (c *chain) Build(h http.Handler) http.Handler {
	for i := range c.middlewares {
		h = c.middlewares[len(c.middlewares)-1-i](h)
	}
	return h
}
