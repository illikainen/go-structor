package validate_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/illikainen/go-structor"
	"github.com/illikainen/go-structor/validate"
)

func TestBase64(t *testing.T) {
	t.Parallel()
	type s struct {
		Base64 string `validate:"base64"`
	}

	require.NoError(t, structor.Apply(&s{
		Base64: base64.StdEncoding.EncodeToString([]byte("foobar")),
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Base64: "invalid",
	}, nil), validate.ErrValidation)
}

func TestRequiredWithout(t *testing.T) {
	t.Parallel()
	type s struct {
		First  string
		Second string `validate:"required_without=First"`
		Third  string
	}

	require.NoError(t, structor.Apply(&s{
		First: "foobar",
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Second: "foobar",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Third: "foobar",
	}, nil), validate.ErrValidation)
}

func TestMultipleRequiredWithout(t *testing.T) {
	t.Parallel()
	type s struct {
		First  string
		Second string `validate:"required_without=First Third"`
		Third  string
	}

	require.NoError(t, structor.Apply(&s{
		First: "foobar",
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Second: "foobar",
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Third: "foobar",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{}, nil), validate.ErrValidation)
}

func TestEmail(t *testing.T) {
	t.Parallel()
	type s struct {
		Email string `validate:"email"`
	}

	require.NoError(t, structor.Apply(&s{
		Email: "foo@example.invalid",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Email: "invalid",
	}, nil), validate.ErrValidation)
}

func TestFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	type s struct {
		File string `validate:"file"`
		Dir  string `validate:"dir"`
	}

	require.ErrorIs(t, structor.Apply(&s{
		File: tmp,
	}, nil), validate.ErrValidation)

	require.ErrorIs(t, structor.Apply(&s{
		File: filepath.Join(tmp, "enoent"),
	}, nil), validate.ErrValidation)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "file"), []byte{}, 0o600))
	require.NoError(t, structor.Apply(&s{
		File: filepath.Join(tmp, "file"),
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Dir: tmp,
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Dir: filepath.Join(tmp, "enoent"),
	}, nil), validate.ErrValidation)
}

func TestHexadecimal(t *testing.T) {
	t.Parallel()
	type s struct {
		Hex string `validate:"hexadecimal"`
	}

	require.NoError(t, structor.Apply(&s{
		Hex: "0123456789abcdef",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Hex: "0123456789abcdefg",
	}, nil), validate.ErrValidation)
}

func TestHost(t *testing.T) {
	t.Parallel()
	type s struct {
		Host string `validate:"hostname"`
	}

	require.NoError(t, structor.Apply(&s{
		Host: "foobar",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Host: "_foobar",
	}, nil), validate.ErrValidation)

	require.NoError(t, structor.Apply(&s{
		Host: "example.invalid",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Host: "foobar.invalid_",
	}, nil), validate.ErrValidation)
}

func TestOneOf(t *testing.T) {
	t.Parallel()
	type s struct {
		In string `validate:"oneof=foo bar"`
	}

	require.NoError(t, structor.Apply(&s{
		In: "foo",
	}, nil))

	require.NoError(t, structor.Apply(&s{
		In: "bar",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		In: "foobar",
	}, nil), validate.ErrValidation)
}

func TestIP(t *testing.T) {
	t.Parallel()
	type s struct {
		IP     string `validate:"ip"`
		IPv4   string `validate:"ipv4"`
		IPv6   string `validate:"ipv6"`
		CIDR   string `validate:"cidr"`
		CIDRv4 string `validate:"cidrv4"`
		CIDRv6 string `validate:"cidrv6"`
	}

	require.NoError(t, structor.Apply(&s{
		IP:     "127.0.0.1",
		IPv4:   "127.0.0.1",
		IPv6:   "::1",
		CIDR:   "127.0.0.1/8",
		CIDRv4: "127.0.0.1/32",
		CIDRv6: "::1/128",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		IPv6: "127.0.0.1",
	}, nil), validate.ErrValidation)

	require.ErrorIs(t, structor.Apply(&s{
		CIDRv6: "127.0.0.1/8",
	}, nil), validate.ErrValidation)

	require.ErrorIs(t, structor.Apply(&s{
		IPv4: "::1",
	}, nil), validate.ErrValidation)

	require.ErrorIs(t, structor.Apply(&s{
		CIDRv4: "::1/128",
	}, nil), validate.ErrValidation)

	require.ErrorIs(t, structor.Apply(&s{
		IP: "foobar",
	}, nil), validate.ErrValidation)

	require.ErrorIs(t, structor.Apply(&s{
		CIDR: "foobar",
	}, nil), validate.ErrValidation)
}

func TestMinMax(t *testing.T) {
	t.Parallel()
	type s struct {
		Value int `validate:"min=2,max=5"`
	}

	require.ErrorIs(t, structor.Apply(&s{
		Value: 1,
	}, nil), validate.ErrValidation)

	require.NoError(t, structor.Apply(&s{
		Value: 2,
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Value: 3,
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Value: 5,
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Value: 6,
	}, nil), validate.ErrValidation)
}

func TestPort(t *testing.T) {
	t.Parallel()
	type s struct {
		Value int `validate:"port"`
	}

	require.ErrorIs(t, structor.Apply(&s{
		Value: -1,
	}, nil), validate.ErrValidation)

	require.NoError(t, structor.Apply(&s{
		Value: 0,
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Value: 22,
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Value: 65535,
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Value: 65536,
	}, nil), validate.ErrValidation)
}

func TestPortrange(t *testing.T) {
	t.Parallel()
	type s struct {
		Value string `validate:"port"`
	}

	require.NoError(t, structor.Apply(&s{
		Value: "0",
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Value: "65535",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Value: "65536",
	}, nil), validate.ErrValidation)

	require.NoError(t, structor.Apply(&s{
		Value: "1-65535",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Value: "1-65536",
	}, nil), validate.ErrValidation)

	require.ErrorIs(t, structor.Apply(&s{
		Value: "1024-22",
	}, nil), validate.ErrValidation)
}

func TestPrintable(t *testing.T) {
	t.Parallel()
	type s struct {
		ASCII   string `validate:"printascii"`
		Unicode string `validate:"alphanumunicode"`
	}

	require.NoError(t, structor.Apply(&s{
		ASCII: "foo",
	}, nil))

	require.NoError(t, structor.Apply(&s{
		ASCII: "",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		ASCII: "foo\x1b",
	}, nil), validate.ErrValidation)

	require.NoError(t, structor.Apply(&s{
		Unicode: "foo",
	}, nil))

	require.NoError(t, structor.Apply(&s{
		Unicode: "",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		Unicode: "foo\x1b",
	}, nil), validate.ErrValidation)
}

func TestRequired(t *testing.T) {
	t.Parallel()
	type s struct {
		Value string `validate:"required"`
	}

	require.ErrorIs(t, structor.Apply(&s{}, nil), validate.ErrValidation)

	require.NoError(t, structor.Apply(&s{
		Value: "foo",
	}, nil))
}

func TestURL(t *testing.T) {
	t.Parallel()
	type s struct {
		URL string `validate:"url"`
	}

	require.NoError(t, structor.Apply(&s{
		URL: "https://example.invalid",
	}, nil))

	require.ErrorIs(t, structor.Apply(&s{
		URL: "http://example.invalid",
	}, nil), validate.ErrValidation)

	require.ErrorIs(t, structor.Apply(&s{
		URL: "example.invalid",
	}, nil), validate.ErrValidation)
}
