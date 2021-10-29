package main

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/nulab/go-git-http-xfer/githttpxfer"
)

func listenAndServeHTTP(ctx context.Context, dataDir, gitFpath string) error {
	ghx, err := githttpxfer.New(dataDir, gitFpath)
	if err != nil {
		return fmt.Errorf("githttpxfer new: %w", err)
	}

	ghx.Event.On(githttpxfer.AfterMatchRouting, func(ctx githttpxfer.Context) {
		repositoryDirectory := filepath.Join(dataDir, ctx.RepoPath())

		err := initDirAsBareRepository(repositoryDirectory)
		if err != nil {
			panic(err)
		}
	})

	chain := NewChain()
	chain.Use(Logging)

	server := &http.Server{Addr: *addr, Handler: chain.Build(ghx)}

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
