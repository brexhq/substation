package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var lengthTests = []struct {
	name      string
	inspector inspLength
	test      []byte
	expected  bool
}{
	{
		"pass",
		inspLength{
			condition: condition{
				Key: "foo",
			},
			Options: inspLengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		inspLength{
			Options: inspLengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		inspLength{
			condition: condition{
				Key: "foo",
			},
			Options: inspLengthOptions{
				Value: 4,
				Type:  "equals",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		inspLength{
			Options: inspLengthOptions{
				Value: 4,
				Type:  "equals",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		inspLength{
			condition: condition{
				Key: "foo",
			},
			Options: inspLengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		inspLength{
			Options: inspLengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		inspLength{
			condition: condition{
				Key: "foo",
			},
			Options: inspLengthOptions{
				Value: 3,
				Type:  "less_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		inspLength{
			Options: inspLengthOptions{
				Value: 3,
				Type:  "less_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		inspLength{
			condition: condition{
				Key: "foo",
			},
			Options: inspLengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		inspLength{
			Options: inspLengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		inspLength{
			condition: condition{
				Key: "foo",
			},
			Options: inspLengthOptions{
				Value: 3,
				Type:  "greater_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		inspLength{
			Options: inspLengthOptions{
				Value: 3,
				Type:  "greater_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		inspLength{
			condition: condition{
				Key:    "foo",
				Negate: true,
			},
			Options: inspLengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		inspLength{
			condition: condition{
				Negate: true,
			},
			Options: inspLengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		inspLength{
			condition: condition{
				Key:    "foo",
				Negate: true,
			},
			Options: inspLengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		inspLength{
			condition: condition{
				Negate: true,
			},
			Options: inspLengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		inspLength{
			condition: condition{
				Key:    "foo",
				Negate: true,
			},
			Options: inspLengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		inspLength{
			condition: condition{
				Negate: true,
			},
			Options: inspLengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"rune pass",
		inspLength{
			Options: inspLengthOptions{
				Measurement: "rune",
				Value:       3,
				Type:        "equals",
			},
		},
		// 3 runes (characters), 4 bytes
		[]byte("a£c"),
		true,
	},
	{
		"array pass",
		inspLength{
			condition: condition{
				Key: "foo",
			},
			Options: inspLengthOptions{
				Measurement: "rune",
				Value:       3,
				Type:        "equals",
			},
		},
		[]byte(`{"foo":["bar",2,{"baz":"qux"}]}`),
		true,
	},
}

func TestLength(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range lengthTests {
		var _ Inspector = test.inspector

		capsule.SetData(test.test)

		check, err := test.inspector.Inspect(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if test.expected != check {
			t.Errorf("expected %v, got %v", test.expected, check)
			t.Errorf("settings: %+v", test.inspector)
			t.Errorf("test: %+v", string(test.test))
		}
	}
}

func benchmarkLengthByte(b *testing.B, inspector inspLength, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkLengthByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range lengthTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkLengthByte(b, test.inspector, capsule)
			},
		)
	}
}
