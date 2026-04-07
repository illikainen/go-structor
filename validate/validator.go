package validate

import (
	"reflect"

	"github.com/pkg/errors"
)

var ErrValidation = errors.New("validation failed")

type Validator interface {
	Init() error
	Validate(name string, value reflect.Value, args []string, root reflect.Value) error
	NumArgs() (int, int)
}

type ValidatorSet struct {
	validators map[string]Validator
}

func NewValidatorSet() (*ValidatorSet, error) {
	validators := map[string]Validator{
		"alphanumunicode":  &validateAlphanumUnicode{},
		"base64":           &validateBase64{},
		"cidr":             &validateCIDR{},
		"cidrv4":           &validateCIDRv4{},
		"cidrv6":           &validateCIDRv6{},
		"dir":              &validateDir{},
		"required_without": &validateRequiredWithout{},
		"email":            &validateEmail{},
		"file":             &validateFile{},
		"hexadecimal":      &validateHexadecimal{},
		"hostname":         &validateHostname{},
		"ip":               &validateIP{},
		"ipv4":             &validateIPv4{},
		"ipv6":             &validateIPv6{},
		"max":              &validateMax{},
		"min":              &validateMin{},
		"oneof":            &validateOneOf{},
		"port":             &validatePort{},
		"printascii":       &validatePrintASCII{},
		"required":         &validateRequired{},
		"url":              &validateURL{},
	}

	for _, validator := range validators {
		err := validator.Init()
		if err != nil {
			return nil, err
		}
	}

	return &ValidatorSet{validators: validators}, nil
}

func (v *ValidatorSet) Validate(name string, fieldName string, value reflect.Value,
	args []string, root reflect.Value,
) error {
	validator, ok := v.validators[name]
	if !ok {
		return errors.Errorf("invalid validator: %s", name)
	}

	minArgs, maxArgs := validator.NumArgs()
	actual := len(args)
	if actual < minArgs || (maxArgs >= 0 && actual > maxArgs) {
		return errors.Errorf("%s: %d arguments provided to %s but only %d-%d arguments accepted",
			fieldName, actual, name, minArgs, maxArgs)
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

			err := validator.Validate(fieldName, v, args, root)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return validator.Validate(fieldName, value, args, root)
}
