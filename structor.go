package structor

import (
	"reflect"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"

	"github.com/illikainen/go-structor/defaults"
	"github.com/illikainen/go-structor/transform"
	"github.com/illikainen/go-structor/validate"
)

type Options struct {
	NoValidate  bool
	NoTransform bool
	NoDefaults  bool
}

func Apply(value any, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}

	validatorSet, err := validate.NewValidatorSet()
	if err != nil {
		return err
	}

	transformerSet, err := transform.NewTransformerSet()
	if err != nil {
		return err
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		err := apply(v.Elem(), validatorSet, transformerSet, opts)
		if err != nil {
			return err
		}
	} else if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice &&
		v.Elem().Type().Elem().Kind() == reflect.Struct {
		for i := range v.Elem().Len() {
			err := apply(v.Elem().Index(i), validatorSet, transformerSet, opts)
			if err != nil {
				return err
			}
		}
	} else {
		return errors.Errorf("%v must be a pointer to a struct or a slice of structs", value)
	}

	return nil
}

func apply(
	v reflect.Value,
	validatorSet *validate.ValidatorSet,
	transformerSet *transform.TransformerSet,
	opts *Options,
) error {
	t := v.Type()
	for i := range v.NumField() {
		fieldTyp := t.Field(i)
		fieldVal := v.Field(i)
		kind := fieldVal.Kind()

		if !fieldTyp.IsExported() {
			continue
		}

		if !opts.NoDefaults {
			_, err := defaults.SetDefault(&fieldTyp, fieldVal)
			if err != nil {
				return err
			}
		}

		if kind == reflect.Struct {
			err := apply(fieldVal, validatorSet, transformerSet, opts)
			if err != nil {
				return err
			}
		} else if kind == reflect.Ptr && fieldVal.Elem().Kind() == reflect.Struct {
			err := apply(fieldVal.Elem(), validatorSet, transformerSet, opts)
			if err != nil {
				return err
			}
		} else if kind == reflect.Slice {
			for j := range fieldVal.Len() {
				v := fieldVal.Index(j)
				if v.Kind() == reflect.Ptr {
					v = v.Elem()
				}

				if v.Kind() == reflect.Struct {
					err := apply(v, validatorSet, transformerSet, opts)
					if err != nil {
						return err
					}
				}
			}
		}

		if !opts.NoValidate {
			validators, err := ParseTag(fieldTyp.Tag.Get("validate"))
			if err != nil {
				return err
			}

			for _, validator := range validators {
				err := validatorSet.Validate(validator.Name, fieldTyp.Name, fieldVal, validator.Args, v)
				if err != nil {
					return err
				}
			}
		}

		if !opts.NoTransform {
			transformers, err := ParseTag(fieldTyp.Tag.Get("transform"))
			if err != nil {
				return err
			}

			for _, transformer := range transformers {
				err := transformerSet.Transform(transformer.Name, fieldTyp.Name, fieldVal, transformer.Args)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type Tag struct {
	Name string
	Args []string
}

func ParseTag(tag string) ([]*Tag, error) {
	if tag == "" {
		return nil, nil
	}

	var tags []*Tag
	for _, elt := range strings.Split(tag, ",") {
		parts := strings.SplitN(elt, "=", 2)
		args := []string{}
		if len(parts) == 2 {
			words, err := shellquote.Split(parts[1])
			if err != nil {
				return nil, errors.WithStack(err)
			}
			args = append(args, words...)
		}

		tags = append(tags, &Tag{
			Name: parts[0],
			Args: args,
		})
	}

	return tags, nil
}
