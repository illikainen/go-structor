package defaults

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

func SetDefault(field *reflect.StructField, value reflect.Value) (bool, error) {
	if !value.IsValid() || !value.IsZero() {
		return false, nil
	}

	def, err := getDefault(field)
	if err != nil {
		return false, err
	}

	if def.IsValid() {
		value.Set(def)
		return true, nil
	}
	return false, nil
}

func getDefault(field *reflect.StructField) (reflect.Value, error) {
	tag := field.Tag.Get("default")
	if tag != "" {
		var data []byte
		kind := field.Type.Kind()
		if kind == reflect.String || (kind == reflect.Ptr && field.Type.Elem().Kind() == reflect.String) {
			data = []byte(`"` + tag + `"`)
		} else {
			data = []byte(tag)
		}

		def := reflect.New(field.Type)
		err := json.Unmarshal(data, def.Interface())
		if err != nil {
			return reflect.ValueOf(nil), errors.WithStack(err)
		}

		return def.Elem(), nil
	}

	return reflect.ValueOf(nil), nil
}
