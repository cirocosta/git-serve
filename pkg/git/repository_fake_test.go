package git_test

import "github.com/cirocosta/git-serve/pkg/git"

type fakeRepository struct {
	dir string
}

func (r *fakeRepository) Directory() string {
	return r.dir
}

var _ git.Repository = (*fakeRepository)(nil)
