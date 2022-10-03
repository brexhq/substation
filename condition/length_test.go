package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var lengthTests = []struct {
	name      string
	inspector Length
	test      []byte
	expected  bool
}{
	{
		"pass",
		Length{
			Key:      "foo",
			Value:    3,
			Function: "equals",
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		Length{
			Value:    3,
			Function: "equals",
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		Length{
			Key:      "foo",
			Value:    4,
			Function: "equals",
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		Length{
			Value:    4,
			Function: "equals",
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		Length{
			Key:      "foo",
			Value:    4,
			Function: "lessthan",
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		Length{
			Value:    4,
			Function: "lessthan",
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		Length{
			Key:      "foo",
			Value:    3,
			Function: "lessthan",
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		Length{
			Value:    3,
			Function: "lessthan",
		},
		[]byte(`bar`),
		false,
	},
	{
		"pass",
		Length{
			Key:      "foo",
			Value:    2,
			Function: "greaterthan",
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		Length{
			Value:    2,
			Function: "greaterthan",
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		Length{
			Key:      "foo",
			Value:    3,
			Function: "greaterthan",
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		Length{
			Value:    3,
			Function: "greaterthan",
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		Length{
			Key:      "foo",
			Value:    3,
			Function: "equals",
			Negate:   true,
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		Length{
			Value:    3,
			Function: "equals",
			Negate:   true,
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		Length{
			Key:      "foo",
			Value:    4,
			Function: "lessthan",
			Negate:   true,
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		Length{
			Value:    4,
			Function: "lessthan",
			Negate:   true,
		},
		[]byte(`bar`),
		false,
	},
	{
		"!pass",
		Length{
			Key:      "foo",
			Value:    2,
			Function: "greaterthan",
			Negate:   true,
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		Length{
			Value:    2,
			Function: "greaterthan",
			Negate:   true,
		},
		[]byte(`bar`),
		false,
	},
	{
		"rune pass",
		Length{
			Type:     "rune",
			Value:    3,
			Function: "equals",
		},
		// 3 runes (characters), 4 bytes
		[]byte("a£c"),
		true,
	},
	{
		"array pass",
		Length{
			Key:      "foo",
			Value:    3,
			Function: "equals",
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

func benchmarkLengthByte(b *testing.B, inspector Length, capsule config.Capsule) {
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
