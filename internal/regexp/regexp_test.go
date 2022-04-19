package regexp

import (
	"errors"
	"testing"
)

var regexpTests = []struct {
	test     string
	expected error
}{
	{
		"foo",
		nil,
	},
	{
		`(\d+):\1`,
		errors.New("regexp did not compile"),
	},
}

func TestRegexp(t *testing.T) {
	for _, test := range regexpTests {
		_, err := Compile(test.test)
		if test.expected == nil && err != nil {
			t.Logf("expected %+v, got %+v", test.expected, err)
			t.Fail()
		} else if test.expected != nil && err == nil {
			t.Logf("expected %+v, got %+v", test.expected, err)
			t.Fail()
		}
	}
}
