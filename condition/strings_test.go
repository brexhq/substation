package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var stringsTests = []struct {
	name      string
	inspector _strings
	test      []byte
	expected  bool
}{
	{
		"pass",
		_strings{
			condition: condition{
				Key: "foo",
			},
			Options: _stringsOptions{
				Type:       "starts_with",
				Expression: "Test",
			},
		},
		[]byte(`{"foo":"Test"}`),
		true,
	},
	{
		"pass",
		_strings{
			Options: _stringsOptions{
				Type:       "starts_with",
				Expression: "Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		_strings{
			Options: _stringsOptions{
				Type:       "starts_with",
				Expression: "Test",
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		_strings{
			Options: _stringsOptions{
				Type:       "equals",
				Expression: "Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		_strings{
			Options: _stringsOptions{
				Type:       "equals",
				Expression: "Test",
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		_strings{
			Options: _stringsOptions{
				Type:       "contains",
				Expression: "es",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		_strings{
			Options: _stringsOptions{
				Type:       "contains",
				Expression: "ABC",
			},
		},
		[]byte("Test"),
		false,
	},
	{
		"!fail",
		_strings{
			condition: condition{
				Negate: true,
			},
			Options: _stringsOptions{
				Type:       "starts_with",
				Expression: "XYZ",
			},
		},
		[]byte("ABC"),
		true,
	},
	{
		"!pass",
		_strings{
			condition: condition{
				Negate: true,
			},
			Options: _stringsOptions{
				Type:       "starts_with",
				Expression: "ABC",
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"!pass",
		_strings{
			condition: condition{
				Negate: true,
			},
			Options: _stringsOptions{
				Type:       "equals",
				Expression: "",
			},
		},
		[]byte(""),
		false,
	},
	{
		"!pass",
		_strings{
			condition: condition{
				Negate: true,
			},
			Options: _stringsOptions{
				Type:       "contains",
				Expression: "A",
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"pass",
		_strings{
			Options: _stringsOptions{
				Type:       "equals",
				Expression: "\"\"",
			},
		},
		[]byte("\"\""),
		true,
	},
	{
		"pass",
		_strings{
			Options: _stringsOptions{
				Type:       "equals",
				Expression: "",
			},
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

func benchmarkStringsByte(b *testing.B, inspector _strings, capsule config.Capsule) {
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
