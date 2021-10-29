package pkg

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/nulab/go-git-http-xfer/githttpxfer"

	"github.com/cirocosta/git-serve/pkg/log"
)

const (
	HTTPDefaultBindAddr              = ":8080"
	HTTPDefaultDataDirectory         = "/tmp/git-serve"
	HTTPDefaultGitExecutableFilepath = "/usr/bin/git"
)

type HTTPServer struct {
	BindAddress           string
	DataDirectory         string
	GitExecutableFilepath string
	NoAuth                bool
	Password              string
	Username              string

	logger *log.Logger
}

func (s *HTTPServer) Run(ctx context.Context) error {
	s.logger = log.From(ctx)

	s.logger.WithFields(log.Fields{
		"bind-addr": s.BindAddress,
		"data-dir":  s.DataDirectory,
		"git":       s.GitExecutableFilepath,
		"no-auth":   s.NoAuth,
	}).Info("starting")
	defer s.logger.Info("finished")

	server, err := s.server()
	if err != nil {
		return fmt.Errorf("server: %w", err)
	}

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return server.Shutdown(ctx)
	case err := <-doneCh:
		return err
	}
}

func (s *HTTPServer) onRouteMatch(xferCtxt githttpxfer.Context) {
	repositoryDirectory := filepath.Join(
		s.DataDirectory,
		xferCtxt.RepoPath(),
	)

	err := initDirAsBareRepository(repositoryDirectory)
	if err != nil {
		panic(err)
	}
}

func (s *HTTPServer) server() (*http.Server, error) {
	ghx, err := githttpxfer.New(s.DataDirectory, s.GitExecutableFilepath)
	if err != nil {
		return nil, fmt.Errorf("ghx new: %w", err)
	}

	ghx.Event.On(githttpxfer.AfterMatchRouting, s.onRouteMatch)

	middlewares := []middleware{
		s.loggingMiddleware,
	}

	if !s.NoAuth {
		s.logger.Info("auth enabled")
		middlewares = append(middlewares, s.authzMiddleware)
	}

	return &http.Server{
		Addr: s.BindAddress,
		Handler: newMiddlewareChain(
			ghx,
			middlewares...,
		),
	}, nil
}

func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()

		s.logger.WithFields(log.Fields{
			"method":   r.Method,
			"url":      r.URL.String(),
			"duration": t2.Sub(t1),
		}).Debug("req")
	})
}

func (s *HTTPServer) authzMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != s.Username || password != s.Password {
			w.Header().Set(
				"WWW-Authenticate",
				`Basic realm="Please enter your username and password."`,
			)

			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			w.Header().Set("Content-Type", "text/plain")

			return
		}
		next.ServeHTTP(w, r)
	})
}

type middleware func(http.Handler) http.Handler

func newMiddlewareChain(handler http.Handler, fns ...middleware) http.Handler {
	for _, f := range fns {
		handler = f(handler)
	}

	return handler
}

type chain struct {
	middlewares []middleware
}
