package validate

import (
	"encoding/base64"
	"encoding/hex"
	"math"
	"net"
	"net/mail"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

// ------
// Base64
// ------
type validateBase64 struct{}

func (v *validateBase64) Init() error {
	return nil
}

func (v *validateBase64) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return errors.Wrapf(ErrValidation, "%s: invalid base64", name)
	}

	return nil
}

func (v *validateBase64) NumArgs() (int, int) {
	return 0, 0
}

// ---------------
// RequiredWithout
// ---------------
type validateRequiredWithout struct{}

func (v *validateRequiredWithout) Init() error {
	return nil
}

func (v *validateRequiredWithout) Validate(name string, value reflect.Value, args []string, root reflect.Value) error {
	if value.IsValid() && !value.IsZero() {
		return nil
	}

	for _, arg := range args {
		v := root.FieldByName(arg) // nosemgrep
		if v.IsValid() && !v.IsZero() {
			return nil
		}
	}

	return errors.Wrapf(ErrValidation, "one of the following fields are required: %s, %s",
		name, strings.Join(args, ", "))
}

func (v *validateRequiredWithout) NumArgs() (int, int) {
	return 1, -1
}

// -----
// Email
// -----
type validateEmail struct{}

func (v *validateEmail) Init() error {
	return nil
}

func (v *validateEmail) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	_, err := mail.ParseAddress(s)
	if err != nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid email", name, s)
	}

	return nil
}

func (v *validateEmail) NumArgs() (int, int) {
	return 0, 0
}

// ----
// File
// ----
type validateFile struct{}

func (v *validateFile) Init() error {
	return nil
}

func (v *validateFile) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	info, err := os.Stat(s)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.Wrapf(ErrValidation, "%s: '%s' does not exist", name, s)
		}
		return errors.WithStack(err)
	}

	if info.IsDir() {
		return errors.Wrapf(ErrValidation, "%s: '%s' is not a file", name, s)
	}

	return nil
}

func (v *validateFile) NumArgs() (int, int) {
	return 0, 0
}

// ----
// Dir
// ----
type validateDir struct{}

func (v *validateDir) Init() error {
	return nil
}

func (v *validateDir) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	info, err := os.Stat(s)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.Wrapf(ErrValidation, "%s: '%s' does not exist", name, s)
		}
		return errors.WithStack(err)
	}

	if !info.IsDir() {
		return errors.Wrapf(ErrValidation, "%s: '%s' is not a directory", name, s)
	}

	return nil
}

func (v *validateDir) NumArgs() (int, int) {
	return 0, 0
}

// -----------
// Hexadecimal
// -----------
type validateHexadecimal struct{}

func (v *validateHexadecimal) Init() error {
	return nil
}

func (v *validateHexadecimal) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	_, err := hex.DecodeString(s)
	if err != nil {
		return errors.Wrapf(ErrValidation, "%s: value is not valid hex", name)
	}

	return nil
}

func (v *validateHexadecimal) NumArgs() (int, int) {
	return 0, 0
}

// --------
// Hostname
// --------
type validateHostname struct {
	rx *regexp.Regexp
}

func (v *validateHostname) Init() error {
	rx, err := regexp.Compile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)
	if err != nil {
		return errors.WithStack(err)
	}
	v.rx = rx
	return nil
}

func (v *validateHostname) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	if len(s) > 253 {
		return errors.Wrapf(ErrValidation, "%s: '%s' is not valid hostname", name, s)
	}

	for _, elt := range strings.Split(s, ".") {
		if !v.rx.MatchString(elt) {
			return errors.Wrapf(ErrValidation, "%s: '%s' is not valid hostname", name, s)
		}
	}

	return nil
}

func (v *validateHostname) NumArgs() (int, int) {
	return 0, 0
}

// ------
// Oneof
// -----
type validateOneOf struct{}

func (v *validateOneOf) Init() error {
	return nil
}

func (v *validateOneOf) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	elemType := value.Type()
	if elemType.Kind() == reflect.Slice {
		elemType = elemType.Elem()
	}

	valid := reflect.New(reflect.SliceOf(elemType))
	switch elemType.Kind() { //nolint:exhaustive
	case reflect.String:
		valid.Elem().Set(reflect.ValueOf(args))
	case reflect.Int:
		var elts []int
		for _, arg := range args {
			n, err := strconv.Atoi(arg)
			if err != nil {
				return errors.WithStack(err)
			}
			elts = append(elts, n)
		}
		valid.Elem().Set(reflect.ValueOf(elts))
	default:
		return errors.Errorf("%s: %s is not a valid type", name, value.Type())
	}

	if value.Type().Kind() == reflect.Slice { // revive:disable-line
		for i := range value.Len() {
			ok := false
			for j := 0; j < valid.Elem().Len() && !ok; j++ {
				if reflect.DeepEqual(value.Index(i), valid.Elem().Index(j)) {
					ok = true
				}
			}

			if !ok {
				return errors.Wrapf(ErrValidation, "%s: '%v' is not among %s",
					name, value, strings.Join(args, ", "))
			}
		}
	} else {
		for i := range valid.Elem().Len() {
			if reflect.DeepEqual(value.Interface(), valid.Elem().Index(i).Interface()) {
				return nil
			}
		}
		return errors.Wrapf(ErrValidation, "%s: '%v' is not among %s", name, value, strings.Join(args, ", "))
	}

	return nil
}

func (v *validateOneOf) NumArgs() (int, int) {
	return 1, -1
}

// --
// IP
// --
type validateIP struct{}

func (v *validateIP) Init() error {
	return nil
}

func (v *validateIP) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	if net.ParseIP(s) == nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IP", name, s)
	}
	return nil
}

func (v *validateIP) NumArgs() (int, int) {
	return 0, -1
}

// ----
// IPv4
// ----
type validateIPv4 struct{}

func (v *validateIPv4) Init() error {
	return nil
}

func (v *validateIPv4) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IP", name, s)
	}

	if ip.To4() == nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IPv4", name, s)
	}
	return nil
}

func (v *validateIPv4) NumArgs() (int, int) {
	return 0, -1
}

// ----
// IPv6
// ----
type validateIPv6 struct{}

func (v *validateIPv6) Init() error {
	return nil
}

func (v *validateIPv6) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IP", name, s)
	}

	if ip.To4() != nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IPv6", name, s)
	}

	return nil
}

func (v *validateIPv6) NumArgs() (int, int) {
	return 0, -1
}

// ----
// CIDR
// ----
type validateCIDR struct{}

func (v *validateCIDR) Init() error {
	return nil
}

func (v *validateCIDR) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	_, _, err := net.ParseCIDR(s)
	if err != nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid CIDR", name, s)
	}

	return nil
}

func (v *validateCIDR) NumArgs() (int, int) {
	return 0, -1
}

// ------
// CIDRv4
// ------
type validateCIDRv4 struct{}

func (v *validateCIDRv4) Init() error {
	return nil
}

func (v *validateCIDRv4) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	cidr, _, err := net.ParseCIDR(s)
	if err != nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid CIDR", name, s)
	}

	if cidr.To4() == nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IPv4", name, s)
	}

	return nil
}

func (v *validateCIDRv4) NumArgs() (int, int) {
	return 0, -1
}

// ------
// CIDRv6
// ------
type validateCIDRv6 struct{}

func (v *validateCIDRv6) Init() error {
	return nil
}

func (v *validateCIDRv6) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	cidr, _, err := net.ParseCIDR(s)
	if err != nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid CIDR", name, s)
	}

	if cidr.To4() != nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IPv6", name, s)
	}

	return nil
}

func (v *validateCIDRv6) NumArgs() (int, int) {
	return 0, -1
}

// ---
// Min
// ---
type validateMin struct{}

func (v *validateMin) Init() error {
	return nil
}

//nolint:dupl
func (v *validateMin) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	switch value.Kind() { //nolint:exhaustive
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(args[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return err
		}

		if value.Int() < n {
			return errors.Wrapf(ErrValidation, "%s: %d must be above %d", name, value.Int(), n)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(args[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return err
		}

		if value.Uint() < n {
			return errors.Wrapf(ErrValidation, "%s: %d must be above %d", name, value.Uint(), n)
		}
		return nil
	}

	return errors.Errorf("%s: %s cannot be used with min", name, value.Type())
}

func (v *validateMin) NumArgs() (int, int) {
	return 1, 1
}

// ---
// Max
// ---
type validateMax struct{}

func (v *validateMax) Init() error {
	return nil
}

//nolint:dupl
func (v *validateMax) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	switch value.Kind() { //nolint:exhaustive
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(args[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return err
		}

		if value.Int() > n {
			return errors.Wrapf(ErrValidation, "%s: %d must be below %d", name, value.Int(), n)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(args[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return err
		}

		if value.Uint() > n {
			return errors.Wrapf(ErrValidation, "%s: %d must be below %d", name, value.Uint(), n)
		}
		return nil
	}

	return errors.Errorf("%s: %s cannot be used with min", name, value.Type())
}

func (v *validateMax) NumArgs() (int, int) {
	return 1, 1
}

// ----
// Port
// ----
type validatePort struct{}

func (v *validatePort) Init() error {
	return nil
}

func (v *validatePort) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	switch value.Kind() { //nolint:exhaustive
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := value.Int()
		if v < 0 || v > math.MaxUint16 {
			return errors.Wrapf(ErrValidation, "%s: %d is not a valid port", name, v)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := value.Uint()
		if v > math.MaxUint16 {
			return errors.Wrapf(ErrValidation, "%s: %d is not a valid port", name, v)
		}
		return nil
	case reflect.String:
		v := value.String()
		elts := strings.SplitN(v, "-", 2)

		start, err := strconv.ParseUint(elts[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return errors.WithStack(err)
		}
		if start > math.MaxUint16 {
			return errors.Wrapf(ErrValidation, "%s: %s is not a valid port", name, v)
		}

		if len(elts) == 2 {
			end, err := strconv.ParseUint(elts[1], 10, reflect.TypeOf(int64(0)).Bits())
			if err != nil {
				return errors.WithStack(err)
			}

			if end < start || end > math.MaxUint16 {
				return errors.Wrapf(ErrValidation, "%s: %s is not a valid port rage", name, v)
			}
		}
		return nil
	}

	return errors.Errorf("%s: %s cannot be used with port", name, value.Type())
}

func (v *validatePort) NumArgs() (int, int) {
	return 0, 0
}

// ----------
// PrintASCII
// ----------
type validatePrintASCII struct{}

func (v *validatePrintASCII) Init() error {
	return nil
}

func (v *validatePrintASCII) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	if !isPrintable(s) {
		return errors.Wrapf(ErrValidation, "%s: value is not printable ascii", name)
	}
	return nil
}

func (v *validatePrintASCII) NumArgs() (int, int) {
	return 0, 1
}

// ---------------
// AlphanumUnicode
// ---------------
type validateAlphanumUnicode struct{}

func (v *validateAlphanumUnicode) Init() error {
	return nil
}

func (v *validateAlphanumUnicode) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	for _, r := range s {
		if !unicode.IsPrint(r) {
			return errors.Wrapf(ErrValidation, "%s: value is not printable unicode", name)
		}
	}
	return nil
}

func (v *validateAlphanumUnicode) NumArgs() (int, int) {
	return 0, 1
}

// --------
// Required
// --------
type validateRequired struct{}

func (v *validateRequired) Init() error {
	return nil
}

func (v *validateRequired) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return errors.Wrapf(ErrValidation, "missing required field: %s", name)
	}
	return nil
}

func (v *validateRequired) NumArgs() (int, int) {
	return 0, 0
}

// ---
// URL
// ---
type validateURL struct{}

func (v *validateURL) Init() error {
	return nil
}

func (v *validateURL) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	var uri *url.URL
	switch v := value.Interface().(type) {
	case *url.URL:
		uri = v
	case string:
		u, err := url.ParseRequestURI(v)
		if err != nil {
			return errors.Wrapf(ErrValidation, "%s: '%s' not a valid URL", name, v)
		}
		uri = u
	default:
		return errors.Errorf("%s: %s is not a string or a *net.URL", name, value.Type())
	}

	if uri.Scheme != "https" {
		return errors.Wrapf(ErrValidation, "%s: '%s' does not have a valid URL scheme", name, uri)
	}

	return nil
}

func (v *validateURL) NumArgs() (int, int) {
	return 0, -1
}
