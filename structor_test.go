package structor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/illikainen/go-structor"
)

func TestStructor(t *testing.T) {
	t.Parallel()

	type last struct {
		Ints           []int   `default:"[9,8,7]" validate:"min=1,max=9"`
		String         *string `default:"foobar"`
		StringChoice   *string `default:"foo" validate:"oneof=foo bar"`
		IntChoice      *int    `default:"3" validate:"oneof=3 4 5"`
		SliceIntChoice []int   `default:"[5,7]" validate:"oneof=5 6 7 8"`
	}
	type inner struct {
		IP              *string `default:"127.0.0.1" validate:"ip"`
		Last            *last
		LastWithDefault *last `default:"{\"String\": \"ohai\"}"`
	}
	type good struct {
		String *string `default:"a string"`
		Inner  inner
	}

	actual := good{}
	expected := good{
		String: ptr("a string"),
		Inner: inner{
			IP: ptr("127.0.0.1"),
			LastWithDefault: &last{
				Ints:           []int{9, 8, 7},
				String:         ptr("ohai"),
				StringChoice:   ptr("foo"),
				IntChoice:      ptr(3),
				SliceIntChoice: []int{5, 7},
			},
		},
	}
	require.NoError(t, structor.Apply(&actual, nil))
	assert.Equal(t, expected, actual)
}

func ptr[T int | string](v T) *T {
	return &v
}

func TestParseTag(t *testing.T) {
	t.Parallel()

	tags, err := structor.ParseTag("foo")
	require.NoError(t, err)
	assert.Equal(t, []*structor.Tag{
		{
			Name: "foo",
			Args: []string{},
		},
	}, tags)

	tags, err = structor.ParseTag("foo=bar")
	require.NoError(t, err)
	assert.Equal(t, []*structor.Tag{
		{
			Name: "foo",
			Args: []string{"bar"},
		},
	}, tags)

	tags, err = structor.ParseTag("foo=bar baz")
	require.NoError(t, err)
	assert.Equal(t, []*structor.Tag{
		{
			Name: "foo",
			Args: []string{"bar", "baz"},
		},
	}, tags)

	tags, err = structor.ParseTag("foo=bar baz,other='111 222'")
	require.NoError(t, err)
	assert.Equal(t, []*structor.Tag{
		{
			Name: "foo",
			Args: []string{"bar", "baz"},
		},
		{
			Name: "other",
			Args: []string{"111 222"},
		},
	}, tags)
}
