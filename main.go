package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/nulab/go-git-http-xfer/githttpxfer"
	"github.com/peterbourgon/ff/v3/ffcli"
	log "github.com/sirupsen/logrus"
)

var (
	version = "dev"

	cmdFlagSet = flag.NewFlagSet("git-serve serve", flag.ExitOnError)

	addr = cmdFlagSet.String(
		"addr", ":8080",
		"address to bind the server to",
	)

	directory = cmdFlagSet.String(
		"directory", "/tmp/git",
		"where git repositories should be stored",
	)

	git = cmdFlagSet.String(
		"git", "/usr/bin/git",
		"absolute path to git executable",
	)

	verbose = cmdFlagSet.Bool(
		"verbose", false,
		"turn verbose logs on/off",
	)
)

func main() {
	cmd := &ffcli.Command{
		Name:       "git-serve",
		ShortUsage: "git-serve [<arg> ...]",
		ShortHelp:  "start the git server",
		FlagSet:    cmdFlagSet,
		Exec: func(_ context.Context, _ []string) error {
			if *verbose {
				log.SetLevel(log.DebugLevel)
			}

			return serve(*addr, *directory, *git)
		},
	}

	err := cmd.ParseAndRun(context.Background(), os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func serve(bindAddr, dataDirectory, gitFpath string) error {
	ghx, err := githttpxfer.New(dataDirectory, gitFpath)
	if err != nil {
		return fmt.Errorf("githttpxfer new: %w", err)
	}

	log.WithFields(log.Fields{
		"bind-addr":      bindAddr,
		"data-directory": dataDirectory,
		"git-fpath":      gitFpath,
	}).Info("started")

	ghx.Event.On(githttpxfer.BeforeUploadPack, func(ctx githttpxfer.Context) {
		log.WithField("ctx", ctx).Info("before-upload-pack")
	})

	ghx.Event.On(githttpxfer.BeforeReceivePack, func(ctx githttpxfer.Context) {
		log.WithField("ctx", ctx).Info("before-receive-pack")
	})

	ghx.Event.On(githttpxfer.AfterMatchRouting, func(ctx githttpxfer.Context) {
		repositoryDirectory := filepath.Join(dataDirectory, ctx.RepoPath())

		err := initDirAsBareRepository(repositoryDirectory)
		if err != nil {
			panic(err)
		}
	})

	chain := NewChain()
	chain.Use(Logging)

	err = http.ListenAndServe(*addr, chain.Build(ghx))
	if err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}
