package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
)

var (
	// LSTreeSeparator is the character set to use to denote the
	// separation between the fields for each row of `git ls-tree`.
	//
	LSTreeSeparator = `%x00`

	// LSTreeFormatFields denotes the set of fields to include in the
	// `--format` flat passed to `git ls-tree`.
	//
	LSTreeFormatFields = []string{
		"%(objectname)",
		"%(objecttype)",
		"%(objectsize)",
		"%(path)",
	}

	// LSTreeFormat is the full value passue to `--format` to denote how
	// each row should be formatted.
	//
	LSTreeFormat = strings.Join(LSTreeFormatFields, LSTreeSeparator)
)

type repository struct {
	// runner is reponsible for running `git` commands.
	//
	runner    Runner
	directory string
}

// NewRepository instantiates a new repository whose contents should be stored
// under a particular directory.
//
func NewRepository(directory string) *repository {
	return &repository{
		runner:    NewRunner(),
		directory: directory,
	}
}

// WithRunner overrides the default Runner used for running `git` commands.
//
func (r *repository) WithRunner(runner Runner) *repository {
	r.runner = runner
	return r
}

// Init initializes the directory where this repository should be set up as a
// bare repository.
//
// ps.: if the directory has already been initialized, no further action will
// be taken.
//
func (r *repository) Init(ctx context.Context) error {
	return initDirAsBareRepository(r.directory)
}

// Files lists the files in a repository at a particular ref.
//
// e.g.: `Files(context.Background(), "master")` lists the files present in the
// repository at the latest commit under the `master` branch.
//
func (r *repository) Files(ctx context.Context, ref string) ([]File, error) {

	println("dir", r.directory)

	out, err := runAt(r.directory,
		`git ls-tree --full-tree -r %s --format="%s"`,
		ref, LSTreeFormat)
	if err != nil {
		return nil, fmt.Errorf("git ls-tree: %w", err)
	}

	println(out)

	return nil, nil
}

func initDirAsBareRepository(dir string) error {
	createdByUs, err := findOrCreateDir(dir)
	if err != nil {
		return fmt.Errorf("find or create dir '%s': %w", dir, err)
	}

	if !createdByUs {
		filesInDirectory, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("readdir '%s': %w", dir, err)
		}

		if n := len(filesInDirectory); n > 0 {
			return fmt.Errorf("dir '%s' already has %d files", dir, n)
		}
	}

	isBare, err := isBareRepository(dir)
	if err != nil {
		return fmt.Errorf("is bare check: %w", err)
	}

	if isBare {
		return nil
	}

	err = initBareRepository(dir)
	if err != nil {
		return fmt.Errorf("init dir as bare repo: %w", err)
	}

	return nil
}

func findOrCreateDir(dir string) (bool, error) {
	createdByUs := true
	notCreatedByUs := !createdByUs

	_, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("stat '%s': %w", dir, err)
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return false, fmt.Errorf("mkdir '%s': %w", dir, err)
		}

		return createdByUs, nil
	}

	return notCreatedByUs, nil
}

func initBareRepository(dir string) error {
	_, err := execAt(dir, "git", "init", "--bare", "--shared")
	if err != nil {
		return fmt.Errorf("init bare '%s': %w", dir, err)
	}

	return nil
}

func isBareRepository(dir string) (bool, error) {
	configFpath := filepath.Join(dir, "config")

	fbytes, err := os.ReadFile(configFpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("readfile '%s': %w", configFpath, err)
	}

	if strings.Contains(string(fbytes), "bare = true") {
		return true, nil
	}

	return false, nil
}

func execAt(dir string, name string, arg ...string) ([]byte, error) {
	c := exec.Command(name, arg...)
	c.Dir = dir
	return c.CombinedOutput()
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
