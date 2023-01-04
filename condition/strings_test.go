package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var stringsTests = []struct {
	name      string
	inspector inspStrings
	test      []byte
	expected  bool
}{
	{
		"pass",
		inspStrings{
			condition: condition{
				Key: "foo",
			},
			Options: inspStringsOptions{
				Type:       "starts_with",
				Expression: "Test",
			},
		},
		[]byte(`{"foo":"Test"}`),
		true,
	},
	{
		"pass",
		inspStrings{
			Options: inspStringsOptions{
				Type:       "starts_with",
				Expression: "Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		inspStrings{
			Options: inspStringsOptions{
				Type:       "starts_with",
				Expression: "Test",
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		inspStrings{
			Options: inspStringsOptions{
				Type:       "equals",
				Expression: "Test",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		inspStrings{
			Options: inspStringsOptions{
				Type:       "equals",
				Expression: "Test",
			},
		},
		[]byte("-Test"),
		false,
	},
	{
		"pass",
		inspStrings{
			Options: inspStringsOptions{
				Type:       "contains",
				Expression: "es",
			},
		},
		[]byte("Test"),
		true,
	},
	{
		"fail",
		inspStrings{
			Options: inspStringsOptions{
				Type:       "contains",
				Expression: "ABC",
			},
		},
		[]byte("Test"),
		false,
	},
	{
		"!fail",
		inspStrings{
			condition: condition{
				Negate: true,
			},
			Options: inspStringsOptions{
				Type:       "starts_with",
				Expression: "XYZ",
			},
		},
		[]byte("ABC"),
		true,
	},
	{
		"!pass",
		inspStrings{
			condition: condition{
				Negate: true,
			},
			Options: inspStringsOptions{
				Type:       "starts_with",
				Expression: "ABC",
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"!pass",
		inspStrings{
			condition: condition{
				Negate: true,
			},
			Options: inspStringsOptions{
				Type:       "equals",
				Expression: "",
			},
		},
		[]byte(""),
		false,
	},
	{
		"!pass",
		inspStrings{
			condition: condition{
				Negate: true,
			},
			Options: inspStringsOptions{
				Type:       "contains",
				Expression: "A",
			},
		},
		[]byte("ABC"),
		false,
	},
	{
		"pass",
		inspStrings{
			Options: inspStringsOptions{
				Type:       "equals",
				Expression: "\"\"",
			},
		},
		[]byte("\"\""),
		true,
	},
	{
		"pass",
		inspStrings{
			Options: inspStringsOptions{
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
		var _ Inspector = test.inspector

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

func benchmarkStringsByte(b *testing.B, inspector inspStrings, capsule config.Capsule) {
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
