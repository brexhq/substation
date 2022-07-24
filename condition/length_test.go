package condition

import (
	"testing"
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
			Value:    3,
			Function: "lessthaneq",
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		Length{
			Value:    3,
			Function: "lessthaneq",
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		Length{
			Key:      "foo",
			Value:    2,
			Function: "lessthaneq",
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		Length{
			Value:    2,
			Function: "lessthaneq",
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
		"pass",
		Length{
			Key:      "foo",
			Value:    3,
			Function: "greaterthaneq",
		},
		[]byte(`{"foo":"bar"}`),
		true,
	},
	{
		"pass",
		Length{
			Value:    3,
			Function: "greaterthaneq",
		},
		[]byte(`bar`),
		true,
	},
	{
		"fail",
		Length{
			Key:      "foo",
			Value:    4,
			Function: "greaterthaneq",
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"fail",
		Length{
			Value:    4,
			Function: "greaterthaneq",
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
			Value:    3,
			Function: "lessthaneq",
			Negate:   true,
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		Length{
			Value:    3,
			Function: "lessthaneq",
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
		"!pass",
		Length{
			Key:      "foo",
			Value:    3,
			Function: "greaterthaneq",
			Negate:   true,
		},
		[]byte(`{"foo":"bar"}`),
		false,
	},
	{
		"!pass",
		Length{
			Value:    3,
			Function: "greaterthaneq",
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
	for _, testing := range lengthTests {
		check, _ := testing.inspector.Inspect(testing.test)

		if testing.expected != check {
			t.Logf("expected %v, got %v", testing.expected, check)
			t.Fail()
		}
	}
}

func benchmarkLengthByte(b *testing.B, inspector Length, test []byte) {
	for i := 0; i < b.N; i++ {
		inspector.Inspect(test)
	}
}

func BenchmarkLengthByte(b *testing.B) {
	for _, test := range lengthTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkLengthByte(b, test.inspector, test.test)
			},
		)
	}
}
