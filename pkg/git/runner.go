package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-multierror"
)

type runner struct{}

func NewRunner() *runner {
	return &runner{}
}

func (r *runner) RunAt(
	ctx context.Context, dir string, argv ...string,
) (string, string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		if stdout.Len() > 0 {
			err = multierror.Append(err, fmt.Errorf(
				"stdout: %s", stdout.String(),
			))
		}

		if stderr.Len() > 0 {
			err = multierror.Append(err, fmt.Errorf(
				"stderr: %s", stderr.String(),
			))
		}

		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}
