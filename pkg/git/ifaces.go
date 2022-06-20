package git

import (
	"context"
	"io"
)

type Repository interface {
	// Directory gives back the name of the directory where the
	// repository's contents are stored at.
	//
	Directory() string
}

type File interface {
	io.Reader

	// Names returns the name of the blob that stores the contents of this
	// file.
	//
	Name() string

	// Path returns the path relative to the repository where the file can
	// be found.
	//
	Path() string
}

type Runner interface {
	// RunAt runs a command per `argv` with the current working directory
	// set to `dir`.
	//
	RunAt(
		ctx context.Context, dir string, argv ...string,
	) (stdout, stderr string, err error)
}
