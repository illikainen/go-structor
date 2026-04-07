package transform

import (
	"reflect"

	"github.com/pkg/errors"
)

type Transformer interface {
	Init() error
	Transform(name string, value reflect.Value, args []string) error
}

type TransformerSet struct {
	transformers map[string]Transformer
}

func NewTransformerSet() (*TransformerSet, error) {
	transformers := map[string]Transformer{
		"fullpath": &transformFullPath{},
	}

	for _, transformer := range transformers {
		err := transformer.Init()
		if err != nil {
			return nil, err
		}
	}

	return &TransformerSet{transformers: transformers}, nil
}

func (v *TransformerSet) Transform(name string, fieldName string, value reflect.Value, args []string) error {
	transformer, ok := v.transformers[name]
	if !ok {
		return errors.Errorf("invalid transformer: %s", name)
	}

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() == reflect.Slice {
		for i := range value.Len() {
			v := value.Index(i)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			err := transformer.Transform(fieldName, v, args)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return transformer.Transform(fieldName, value, args)
}
