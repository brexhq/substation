package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
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
	{
		"pass",
		Strings{
			Function:   "equals",
			Expression: "",
		},
		[]byte(``),
		true,
	},
}

func TestStrings(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range stringsTests {
		capsule.SetData(test.test)

		check, err := test.inspector.Inspect(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if test.expected != check {
			t.Errorf("expected %v, got %v", test.expected, check)
		}
	}
}

func benchmarkStringsByte(b *testing.B, inspector Strings, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkStringsByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range stringsTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkStringsByte(b, test.inspector, capsule)
			},
		)
	}
}
