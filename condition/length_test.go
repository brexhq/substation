package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var lengthTests = []struct {
	name      string
	inspector length
	test      []byte
	expected  bool
}{
	{
		"pass",
		length{
			condition: condition{
				Key: "foo",
			},
			Options: lengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		length{
			Options: lengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		length{
			condition: condition{
				Key: "foo",
			},
			Options: lengthOptions{
				Value: 4,
				Type:  "equals",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		length{
			Options: lengthOptions{
				Value: 4,
				Type:  "equals",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		length{
			condition: condition{
				Key: "foo",
			},
			Options: lengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		length{
			Options: lengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		length{
			condition: condition{
				Key: "foo",
			},
			Options: lengthOptions{
				Value: 3,
				Type:  "less_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		length{
			Options: lengthOptions{
				Value: 3,
				Type:  "less_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		length{
			condition: condition{
				Key: "foo",
			},
			Options: lengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		length{
			Options: lengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		length{
			condition: condition{
				Key: "foo",
			},
			Options: lengthOptions{
				Value: 3,
				Type:  "greater_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		length{
			Options: lengthOptions{
				Value: 3,
				Type:  "greater_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		length{
			condition: condition{
				Key:    "foo",
				Negate: true,
			},
			Options: lengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		length{
			condition: condition{
				Negate: true,
			},
			Options: lengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		length{
			condition: condition{
				Key:    "foo",
				Negate: true,
			},
			Options: lengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		length{
			condition: condition{
				Negate: true,
			},
			Options: lengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		length{
			condition: condition{
				Key:    "foo",
				Negate: true,
			},
			Options: lengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		length{
			condition: condition{
				Negate: true,
			},
			Options: lengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"rune pass",
		length{
			Options: lengthOptions{
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
		length{
			condition: condition{
				Key: "foo",
			},
			Options: lengthOptions{
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

func benchmarkLengthByte(b *testing.B, inspector length, capsule config.Capsule) {
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
