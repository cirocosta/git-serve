package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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

	startSSHD = cmdFlagSet.Bool(
		"start-sshd", false,
		"start sshd",
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

			log.WithFields(log.Fields{
				"bind-addr":      *addr,
				"data-directory": *directory,
				"git-fpath":      *git,
				"start-sshd":     *startSSHD,
			}).Info("starting")

			g, ctx := errgroup.WithContext(signalHandlingContext())

			if *startSSHD {
				g.Go(func() error {
					log.Info("starting sshd")
					return runSSHDaemon(ctx)
				})
			}

			g.Go(func() error {
				log.Info("starting http")
				return listenAndServeHTTP(ctx, *directory, *git)
			})

			return g.Wait()
		},
	}

	err := cmd.ParseAndRun(context.Background(), os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
