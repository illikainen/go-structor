package transform

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func expandPath(path string) (string, error) {
	if path == "" {
		return "", errors.Errorf("empty path")
	}

	sep := string(os.PathSeparator)
	if strings.HasPrefix(path, "~"+sep) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.WithStack(err)
		}

		path = filepath.Join(home, path[1+len(sep):])
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return path, nil
}
