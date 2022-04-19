package condition

import (
	"testing"
)

func TestStrings(t *testing.T) {
	var tests = []struct {
		condition Strings
		test      []byte
		expected  bool
	}{
		{
			Strings{
				Function:   "startswith",
				Expression: "Test",
				Key:        "foo",
			},
			[]byte(`{"foo":"Test"}`),
			true,
		},
		{
			Strings{
				Function:   "startswith",
				Expression: "Test",
			},
			[]byte("Test"),
			true,
		},
		{
			Strings{
				Function:   "startswith",
				Expression: "Test",
			},
			[]byte("-Test"),
			false,
		},
		{
			Strings{
				Function:   "equals",
				Expression: "Test",
			},
			[]byte("Test"),
			true,
		},
		{
			Strings{
				Function:   "equals",
				Expression: "Test",
			},
			[]byte("-Test"),
			false,
		},
		{
			Strings{
				Function:   "contains",
				Expression: "es",
			},
			[]byte("Test"),
			true,
		},
		{
			Strings{
				Function:   "contains",
				Expression: "ABC",
			},
			[]byte("Test"),
			false,
		},
		{
			Strings{
				Function:   "startswith",
				Negate:     true,
				Expression: "XYZ",
			},
			[]byte("ABC"),
			true,
		},
		{
			Strings{
				Function:   "startswith",
				Negate:     true,
				Expression: "ABC",
			},
			[]byte("ABC"),
			false,
		},
		{
			Strings{
				Function:   "equals",
				Negate:     true,
				Expression: "",
			},
			[]byte(""),
			false,
		},
		{
			Strings{
				Function:   "contains",
				Negate:     true,
				Expression: "A",
			},
			[]byte("ABC"),
			false,
		},
		{
			Strings{
				Function: "equals",
				// Negate:     true,
				Expression: "\"\"",
			},
			[]byte("\"\""),
			true,
		},
	}

	for _, testing := range tests {
		check, _ := testing.condition.Inspect(testing.test)

		if testing.expected != check {
			t.Logf("expected %v, got %v", testing.expected, check)
			t.Fail()
		}
	}
}
