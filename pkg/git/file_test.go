package git_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cirocosta/git-serve/pkg/git"
)

func TestFile(t *testing.T) {
	var join = func(sep string, v ...string) string {
		return strings.Join(v, sep)
	}

	type file struct {
		name string
		path string
	}

	repository := &fakeRepository{
		dir: "/tmp",
	}

	for _, tc := range []struct {
		scenario   string
		repository git.Repository
		input      string
		err        string
		expected   file
	}{
		{
			scenario: "empty",
			err:      `split: expected 4 but got 1 while splitting ''`,
		},

		{
			scenario: "bad separator",
			input:    "a b c d",
			err:      `split: expected 4 but got 1 while splitting 'a b c d'`,
		},

		{
			scenario: "not blob",
			input:    join(git.LSTreeSeparator, "a", "tree", "b", "c"),
			err:      `expected file type to be 'blob', got 'tree'`,
		},

		{
			scenario: "happy",
			input:    join(git.LSTreeSeparator, "a042389697", "blob", "12", "foo.md"),
			expected: file{
				name: "a042389697",
				path: "/tmp/foo.md",
			},
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			file, err := git.NewFile(repository, tc.input)
			if err != nil {
				if tc.err == "" {
					t.Fatalf("expected no err, got %v", err)
				}

				if d := cmp.Diff(tc.err, err.Error()); d != "" {
					t.Fatalf("%s", d)
				}

				return
			}

			expected, actual := tc.expected.name, file.Name()
			if actual != expected {
				t.Fatalf("name: expected '%s', got '%s'",
					expected, actual,
				)
			}

			expected, actual = tc.expected.path, file.Path()
			if actual != expected {
				t.Fatalf("path: expected '%s', got '%s'",
					expected, actual,
				)
			}
		})
	}
}
