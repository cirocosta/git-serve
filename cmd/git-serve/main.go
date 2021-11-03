package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3"
	"golang.org/x/sync/errgroup"

	"github.com/cirocosta/git-serve/pkg"
	"github.com/cirocosta/git-serve/pkg/log"
	"github.com/cirocosta/git-serve/pkg/server"
)

var (
	version = "dev"

	cmdFlagSet = flag.NewFlagSet("git-serve", flag.ExitOnError)

	httpBindAddr = cmdFlagSet.String(
		"http-bind-addr", server.HTTPDefaultBindAddr,
		"address to bind the http server to",
	)

	httpUsername = cmdFlagSet.String(
		"http-username", "admin",
		"username",
	)

	httpPassword = cmdFlagSet.String(
		"http-password", "admin",
		"password",
	)

	httpNoAuth = cmdFlagSet.Bool(
		"http-no-auth", false,
		"disable default use of basic auth for http",
	)

	dataDirectory = cmdFlagSet.String(
		"data-dir", server.HTTPDefaultDataDirectory,
		"directory where repositories will be stored",
	)

	git = cmdFlagSet.String(
		"git", server.HTTPDefaultGitExecutableFilepath,
		"absolute path to git executable",
	)

	sshBindAddr = cmdFlagSet.String(
		"ssh-bind-addr", server.SSHDefaultBindAddress,
		"address to bind the ssh server to",
	)

	sshHostKey = cmdFlagSet.String(
		"ssh-host-key", "",
		"path to private key to use for the ssh server",
	)

	sshAuthorizedKeys = cmdFlagSet.String(
		"ssh-authorized-keys", "",
		"path to public keys to authorized (ssh format)",
	)

	sshNoAuth = cmdFlagSet.Bool(
		"ssh-no-auth", false,
		"disable default use of public key auth for ssh",
	)

	verbose = cmdFlagSet.Bool(
		"v", false,
		"turn verbose logs on/off",
	)
)

func main() {
	if err := ff.Parse(
		cmdFlagSet, os.Args[1:],
		ff.WithEnvVarPrefix("GIT_SERVE_"),
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := pkg.SignalHandlingContext(context.Background())
	if err := exec(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exec(ctx context.Context) error {
	if *verbose {
		log.Verbose()
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		ctx := log.WithLogger(ctx, log.From(ctx).
			WithField("component", "http"),
		)

		return (&server.HTTPServer{
			BindAddress:           *httpBindAddr,
			DataDirectory:         *dataDirectory,
			GitExecutableFilepath: *git,
			NoAuth:                *httpNoAuth,
			Password:              *httpPassword,
			Username:              *httpUsername,
		}).Run(ctx)
	})

	g.Go(func() error {
		ctx := log.WithLogger(ctx, log.From(ctx).
			WithField("component", "ssh"),
		)

		return (&server.SSHServer{
			AuthorizedKeysFilepath: *sshAuthorizedKeys,
			BindAddress:            *sshBindAddr,
			DataDirectory:          *dataDirectory,
			GitExecutableFilepath:  *git,
			HostKeyFilepath:        *sshHostKey,
			NoAuth:                 *sshNoAuth,
		}).Run(ctx)
	})

	if err := g.Wait(); err != nil {
		log.From(ctx).Error(err)
	}

	return nil
}
