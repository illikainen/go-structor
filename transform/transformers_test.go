package transform_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/illikainen/go-structor"
)

func TestFullPath(t *testing.T) {
	t.Parallel()

	type s struct {
		Path        string `transform:"fullpath"`
		DefaultPath string `transform:"fullpath" default:"~/foobar"`
	}

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	s0 := s{}
	require.NoError(t, structor.Apply(&s0, nil))
	assert.Equal(t, s{
		Path:        "",
		DefaultPath: filepath.Join(home, "foobar"),
	}, s0)

	s1 := s{Path: "~/asdf"}
	require.NoError(t, structor.Apply(&s1, nil))
	assert.Equal(t, s{
		Path:        filepath.Join(home, "asdf"),
		DefaultPath: filepath.Join(home, "foobar"),
	}, s1)

	s2 := s{DefaultPath: "~/asdf"}
	require.NoError(t, structor.Apply(&s2, nil))
	assert.Equal(t, s{
		Path:        "",
		DefaultPath: filepath.Join(home, "asdf"),
	}, s2)
}
