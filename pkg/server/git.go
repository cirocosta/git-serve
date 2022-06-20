package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func listFiles(dir, rev string) ([]string, error) {
	// objectmode
	//     The mode of the object.
	//
	// objecttype
	//     The type of the object (commit, blob or tree).
	//
	// objectname
	//     The name of the object.
	//
	// objectsize[:padded]
	//     The size of a blob object ("-" if it’s a commit or tree).
	//
	// path
	//     The pathname of the object.
	//
	const format = `%(path)`

	output, err := execAt(dir, "git",
		"ls-tree", "--full-tree", "-r", "--format="+format, rev,
	)
	if err != nil {
		return nil, fmt.Errorf("ls-tree '%s' '%s: %w", dir, rev, err)
	}

	files := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		files = append(files, line)
	}

	return files, nil
}

func initDirAsBareRepository(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat '%s': %w", dir, err)
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdir '%s': %w", dir, err)
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
