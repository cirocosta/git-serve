package git_test

import "github.com/cirocosta/git-serve/pkg/git"

type fakeFile struct {
	name string
	path string
}

func (f *fakeFile) Read(p []byte) (n int, err error) {
	panic("not implemented")
}

func (f *fakeFile) Name() string {
	return f.name
}

func (f *fakeFile) Path() string {
	return f.path
}

var _ git.File = (*fakeFile)(nil)
