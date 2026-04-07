package transform

import (
	"reflect"

	"github.com/pkg/errors"
)

// --------
// FullPath
// --------
type transformFullPath struct{}

func (t *transformFullPath) Init() error {
	return nil
}

func (t *transformFullPath) Transform(name string, value reflect.Value, _ []string) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	if value.Kind() != reflect.String {
		return errors.Errorf("%s must be a string", name)
	}

	path, err := expandPath(value.String())
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(path))
	return nil
}
