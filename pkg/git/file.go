package git

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type file struct {
	repository Repository

	name  string
	ttype string
	size  string
	path  string
}

var _ io.Reader = (*file)(nil)
var _ File = (*file)(nil)

func NewFile(repository Repository, v string) (*file, error) {
	fields := strings.Split(v, LSTreeSeparator)

	if actual, expected := len(fields), len(LSTreeFormatFields); expected != actual {
		return nil, fmt.Errorf("split: "+
			"expected %d but got %d while splitting '%s'",
			expected, actual, v,
		)
	}

	f := &file{
		name:  fields[0],
		ttype: fields[1],
		size:  fields[2],
		path:  fields[3],

		repository: repository,
	}

	const expectedType = "blob"
	if f.ttype != expectedType {
		return nil, fmt.Errorf("expected file type to be '%s', got '%s'",
			expectedType, f.ttype,
		)
	}

	return f, nil
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Path() string {
	return filepath.Join(f.repository.Directory(), f.path)
}

func (f *file) Read(p []byte) (int, error) {
	// execat (
	// 	f.repository.directory,
	// 	git cat-file blob 3b18e512dba79e4c8300dd08aeb37f8e728b8dad
	// )

	return -1, nil
}
