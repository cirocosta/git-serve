package git_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/shlex"

	"github.com/cirocosta/git-serve/pkg/git"
)

func TestRepositoryInit_fresh(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}

	repository := git.NewRepository(dir)

	if err := repository.Init(context.Background()); err != nil {
		t.Fatalf("init: expected no errors, got %v", err)
	}

	isBare := mustRunAt(dir, "git rev-parse --is-bare-repository")
	if isBare != "true" {
		t.Fatalf("is bare: '%v'", isBare)
	}

	// ensure idempotency
	if err := repository.Init(context.Background()); err != nil {
		t.Fatalf("init: expected no err after re-init, got %v", err)
	}
}

func TestRepositoryInit_nonEmpty(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}

	fpath := filepath.Join(dir, "foo.txt")
	content := []byte("hello")
	if err := os.WriteFile(fpath, content, 0755); err != nil {
		t.Fatalf("write file '%s': %v", fpath, err)
	}

	err = git.NewRepository(dir).Init(context.Background())
	if err == nil {
		t.Fatalf("expected err, got nil")
	}

	expected, actual := "already has 1 files", err.Error()
	if !strings.Contains(actual, expected) {
		t.Fatalf("init: expected err '%s', got '%s'", expected, actual)
	}
}

func TestRepositoryFiles_empty(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}

	ctx := context.Background()
	repository := git.NewRepository(dir)

	if err := repository.Init(ctx); err != nil {
		t.Fatalf("init: expect no err, got '%v'", err)
	}

	// with no files at all, we expect listing to fail as the ref wouldn't
	// exist just yet
	//
	files, err := repository.Files(ctx, "main")
	if err == nil {
		t.Fatalf("files: expected err, got '%v'", err)
	}
	if !strings.Contains(err.Error(), "Not a valid object name main") {
		t.Fatalf("files: expected err to mention `main` not valid name")
	}

	if len(files) != 0 {
		t.Fatalf("expected no files, got %v", files)
	}

	// push something to it

	scratch, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("mkdirtemp: %v", err)
	}

	mustRunAt(scratch, `/bin/bash -c "
	git clone %s .
	echo foo > file.txt
	git add --all .
	git commit -m foo
	git push origin HEAD"`, dir)

	files, err = repository.Files(ctx, "main")
	if err != nil {
		t.Fatalf("files: expected no err, got '%v'", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %v", files)
	}
}

func mustRunAt(dir, format string, args ...interface{}) string {
	res, err := runAt(dir, format, args...)
	if err != nil {
		cmd := fmt.Sprintf(format, args...)
		panic(fmt.Errorf("run '%s' at '%s': %w",
			dir, cmd, err,
		))
	}

	return res
}

func runAt(dir, format string, args ...interface{}) (string, error) {
	argv, err := shlex.Split(fmt.Sprintf(format, args...))
	if err != nil {
		return "", fmt.Errorf("shlex split: %w", err)
	}

	c := exec.Command(argv[0], argv[1:]...)
	c.Dir = dir

	b, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("combinet output: %w", err)
	}

	return strings.TrimSpace(string(b)), nil
}
