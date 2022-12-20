package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var lengthTests = []struct {
	name      string
	inspector _length
	test      []byte
	expected  bool
}{
	{
		"pass",
		_length{
			condition: condition{
				Key: "foo",
			},
			Options: _lengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		_length{
			Options: _lengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		_length{
			condition: condition{
				Key: "foo",
			},
			Options: _lengthOptions{
				Value: 4,
				Type:  "equals",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		_length{
			Options: _lengthOptions{
				Value: 4,
				Type:  "equals",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		_length{
			condition: condition{
				Key: "foo",
			},
			Options: _lengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		_length{
			Options: _lengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		_length{
			condition: condition{
				Key: "foo",
			},
			Options: _lengthOptions{
				Value: 3,
				Type:  "less_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		_length{
			Options: _lengthOptions{
				Value: 3,
				Type:  "less_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		_length{
			condition: condition{
				Key: "foo",
			},
			Options: _lengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		_length{
			Options: _lengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		_length{
			condition: condition{
				Key: "foo",
			},
			Options: _lengthOptions{
				Value: 3,
				Type:  "greater_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		_length{
			Options: _lengthOptions{
				Value: 3,
				Type:  "greater_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		_length{
			condition: condition{
				Key:    "foo",
				Negate: true,
			},
			Options: _lengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		_length{
			condition: condition{
				Negate: true,
			},
			Options: _lengthOptions{
				Value: 3,
				Type:  "equals",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		_length{
			condition: condition{
				Key:    "foo",
				Negate: true,
			},
			Options: _lengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		_length{
			condition: condition{
				Negate: true,
			},
			Options: _lengthOptions{
				Value: 4,
				Type:  "less_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		_length{
			condition: condition{
				Key:    "foo",
				Negate: true,
			},
			Options: _lengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		_length{
			condition: condition{
				Negate: true,
			},
			Options: _lengthOptions{
				Value: 2,
				Type:  "greater_than",
			},
		},
		[]byte(`bar`),
		false,
	},
	{
		"rune pass",
		_length{
			Options: _lengthOptions{
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
		_length{
			condition: condition{
				Key: "foo",
			},
			Options: _lengthOptions{
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

func benchmarkLengthByte(b *testing.B, inspector _length, capsule config.Capsule) {
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
