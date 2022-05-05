package condition

import (
	"testing"
)

var stringsTests = []struct {
	name      string
	inspector Strings
	test      []byte
	expected  bool
}{
	{
		"pass",
		Strings{
			Function:   "startswith",
			Expression: "Test",
			Key:        "foo",
		},
		[]byte(`{"foo":"Test"}`),
		true,
	},
	{
		"pass",
		Strings{
			Function:   "startswith",
			Expression: "Test",
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		Strings{
			Function:   "startswith",
			Expression: "Test",
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		Strings{
			Function:   "equals",
			Expression: "Test",
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		Strings{
			Function:   "equals",
			Expression: "Test",
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		Strings{
			Function:   "contains",
			Expression: "es",
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		Strings{
			Function:   "contains",
			Expression: "ABC",
		},
		[]byte("Test"),
		false,
	},
	{
		"!fail",
		Strings{
			Function:   "startswith",
			Negate:     true,
			Expression: "XYZ",
		},
		[]byte("ABC"),
		true,
	},
	{
		"!pass",
		Strings{
			Function:   "startswith",
			Negate:     true,
			Expression: "ABC",
		},
		[]byte("ABC"),
		false,
	},
	{
		"!pass",
		Strings{
			Function:   "equals",
			Negate:     true,
			Expression: "",
		},
		[]byte(""),
		false,
	},
	{
		"!pass",
		Strings{
			Function:   "contains",
			Negate:     true,
			Expression: "A",
		},
		[]byte("ABC"),
		false,
	},
	{
		"pass",
		Strings{
			Function:   "equals",
			Expression: "\"\"",
		},
		[]byte("\"\""),
		true,
	},
}

func TestStrings(t *testing.T) {
	for _, testing := range stringsTests {
		check, _ := testing.inspector.Inspect(testing.test)

		if testing.expected != check {
			t.Logf("expected %v, got %v", testing.expected, check)
			t.Fail()
		}
	}
}

func benchmarkStringsByte(b *testing.B, inspector Strings, test []byte) {
	for i := 0; i < b.N; i++ {
		inspector.Inspect(test)
	}
}

func BenchmarkStringsByte(b *testing.B) {
	for _, test := range stringsTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkStringsByte(b, test.inspector, test.test)
			},
		)
	}
}
