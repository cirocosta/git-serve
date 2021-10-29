package main

import (
	"context"
	"os"
	"os/exec"
)

func runSSHDaemon(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "/usr/sbin/sshd")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		return cmd.Process.Kill()
	case err := <-doneCh:
		return err
	}
}
