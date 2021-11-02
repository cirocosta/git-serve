package server

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/gliderlabs/ssh"
	"github.com/google/shlex"
	"golang.org/x/sync/errgroup"

	"github.com/cirocosta/git-serve/pkg/log"
)

// SSHDefaultBindAddress is the default <ip>:<port> tuple that's used to
// determine to which ip and port the SSH server should bind its listening
// socket to.
//
const SSHDefaultBindAddress = ":2222"

//go:embed default_host_key.txt
var defaultHostKey []byte

type SSHServer struct {
	AuthorizedKeysFilepath string
	BindAddress            string
	DataDirectory          string
	GitExecutableFilepath  string
	HostKeyFilepath        string
	NoAuth                 bool

	logger         *log.Logger
	authorizedKeys []ssh.PublicKey
}

func (s *SSHServer) Run(ctx context.Context) error {
	s.logger = log.From(ctx)

	s.logger.WithFields(log.Fields{
		"authorized-keys": s.AuthorizedKeysFilepath,
		"bind-addr":       s.BindAddress,
		"host-key":        s.HostKeyFilepath,
		"no-auth":         s.NoAuth,
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

func (s *SSHServer) loadAuthorizedKeys() error {
	content, err := os.ReadFile(s.AuthorizedKeysFilepath)
	if err != nil {
		return fmt.Errorf("read file '%s': %w",
			s.AuthorizedKeysFilepath, err,
		)
	}

	pk, _, _, _, err := ssh.ParseAuthorizedKey(content)
	if err != nil {
		return fmt.Errorf("parse auth key: %w", err)
	}

	s.authorizedKeys = append(s.authorizedKeys, pk)

	return nil
}

func (s *SSHServer) server() (*ssh.Server, error) {
	var err error

	server := &ssh.Server{
		Addr:    s.BindAddress,
		Handler: s.handleSession,
	}

	if !s.NoAuth {
		s.logger.Info("auth enabled")

		err = s.loadAuthorizedKeys()
		if err != nil {
			return nil, fmt.Errorf("load authorized keys: %w", err)
		}

		err := server.SetOption(ssh.PublicKeyAuth(s.isAuthz))
		if err != nil {
			return nil, fmt.Errorf("opt publickeyauth: %w", err)
		}
	}

	hostKeyPEM := defaultHostKey

	if s.HostKeyFilepath != "" {
		s.logger.Info("using user-provided host-key")
		hostKeyPEM, err = os.ReadFile(s.HostKeyFilepath)
		if err != nil {
			return nil, fmt.Errorf("read file '%s': %w",
				s.HostKeyFilepath, err,
			)
		}
	}

	err = server.SetOption(ssh.HostKeyPEM(hostKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("opt custom hostkey: %w", err)
	}

	return server, nil
}

func (s *SSHServer) handleSession(session ssh.Session) {
	logger := s.logger.WithFields(log.Fields{
		"raw-cmd": session.RawCommand(),
	})

	logger.Debug("session start")
	defer logger.Debug("session finished")

	ctx := log.WithLogger(session.Context(), logger)

	if err := s.runSessionCmd(ctx, session); err != nil {
		s.logger.WithError(err).Error("run session cmd")
	}
}

func (s *SSHServer) runSessionCmd(ctx context.Context, session ssh.Session) error {
	args, err := shlex.Split(session.RawCommand())
	if err != nil {
		return fmt.Errorf("split: %w", err)
	}

	repositoryDirectory := filepath.Join(s.DataDirectory, args[len(args)-1])
	err = initDirAsBareRepository(repositoryDirectory)
	if err != nil {
		return fmt.Errorf("init dir as bar repo: %w", err)
	}

	args[len(args)-1] = repositoryDirectory

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	closers := []io.Closer{}

	var closeAll = func() {
		for _, closer := range closers {
			closer.Close()
		}
	}
	defer closeAll()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	closers = append(closers, stdout)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}
	closers = append(closers, stderr)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}
	closers = append(closers, stdin)

	eg, ctx := errgroup.WithContext(ctx)
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("cmd start: %w", err)
	}

	eg.Go(func() error {
		defer stdin.Close()

		if _, err := io.Copy(stdin, session); err != nil {
			return fmt.Errorf("copy write session to stdin: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer stdout.Close()

		if _, err := io.Copy(session, stdout); err != nil {
			return fmt.Errorf("write stdout to session: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer stderr.Close()

		if _, err := io.Copy(session.Stderr(), stderr); err != nil {
			return fmt.Errorf("write stderr to session: %s", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer closeAll()

		if err := cmd.Wait(); err != nil {
			session.Close()
			return fmt.Errorf("cmd wait: %w", err)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("errgroup wait: %w", err)
	}

	if err := session.Exit(exitCodeFromError(err)); err != nil {
		return fmt.Errorf("session exit: %w", err)
	}

	return nil
}

func (s *SSHServer) isAuthz(ctx ssh.Context, key ssh.PublicKey) bool {
	for _, authorizedKey := range s.authorizedKeys {
		if ssh.KeysEqual(key, authorizedKey) {
			return true
		}
	}

	return false
}

func exitCodeFromError(err error) int {
	if err == nil {
		return 0
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return 1
	}

	waitStatus, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		// This is a fallback and should at least let us return something useful
		// when running on Windows, even if it isn't completely accurate.
		if exitErr.Success() {
			return 0
		}

		return 1
	}

	return waitStatus.ExitStatus()
}
