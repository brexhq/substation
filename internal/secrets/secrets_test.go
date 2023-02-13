package secrets

import (
	"context"
	"errors"
	"testing"
)

var interpolateTest = []struct {
	name     string
	test     string
	expected string
	err      error
}{
	{
		"zero secrets",
		"/path/to/secret/",
		"/path/to/secret/",
		nil,
	},
	{
		"one secrets",
		"/path/to/secret/${SECRETS_ENV:SECRET_FOO}",
		"/path/to/secret/foo",
		nil,
	},
	{
		"two secrets",
		"/path/to/secret/${SECRETS_ENV:SECRET_FOO}/${SECRETS_ENV:SECRET_BAR}",
		"/path/to/secret/foo/bar",
		nil,
	},
	{
		"secrets not found",
		"/path/to/secret/${SECRETS_NIL:SECRET_FOO}/${SECRETS_NIL:SECRET_BAR}",
		"",
		errSecretNotFound,
	},
}

func TestInterpolate(t *testing.T) {
	t.Setenv("SECRET_FOO", "foo")
	t.Setenv("SECRET_BAR", "bar")

	for _, test := range interpolateTest {
		interp, err := Interpolate(context.TODO(), test.test)
		if test.err != nil && !errors.Is(err, test.err) {
			continue
		}

		if err != nil {
			t.Errorf("got error %v", err)
		}

		if interp != test.expected {
			t.Errorf("expected %s, got %s", test.expected, interp)
		}
	}
}
