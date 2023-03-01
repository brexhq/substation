package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var expandTests = []struct {
	name     string
	proc     procExpand
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"objects",
		procExpand{
			process: process{
				Key: "a",
			},
		},
		[]byte(`{"a":[{"b":"c"}]}`),
		[][]byte{
			[]byte(`{"b":"c"}`),
		},
		nil,
	},
	{
		"objects with key",
		procExpand{
			process: process{
				Key: "a",
			},
		},
		[]byte(`{"a":[{"b":"c"},{"d":"e"}],"x":"y"}`),
		[][]byte{
			[]byte(`{"x":"y","b":"c"}`),
			[]byte(`{"x":"y","d":"e"}`),
		},
		nil,
	},
	{
		"non-objects with key",
		procExpand{
			process: process{
				Key: "a",
			},
		},
		[]byte(`{"a":["b","c"],"d":"e"}`),
		[][]byte{
			[]byte(`{"d":"e"}`),
			[]byte(`{"d":"e"}`),
		},
		nil,
	},
	{
		"objects with set key",
		procExpand{
			process: process{
				Key:    "a",
				SetKey: "a",
			},
		},
		[]byte(`{"a":[{"b":"c"},{"d":"e"}],"x":"y"}`),
		[][]byte{
			[]byte(`{"x":"y","a":{"b":"c"}}`),
			[]byte(`{"x":"y","a":{"d":"e"}}`),
		},
		nil,
	},
	{
		"strings with key",
		procExpand{
			process: process{
				Key:    "a",
				SetKey: "a",
			},
		},
		[]byte(`{"a":["b","c"],"d":"e"}`),
		[][]byte{
			[]byte(`{"d":"e","a":"b"}`),
			[]byte(`{"d":"e","a":"c"}`),
		},
		nil,
	},
	{
		"objects with deeply nested set key",
		procExpand{
			process: process{
				Key:    "a.b",
				SetKey: "a.b.c.d",
			},
		},
		[]byte(`{"a":{"b":[{"g":"h"},{"i":"j"}],"x":"y"}}`),
		[][]byte{
			[]byte(`{"a":{"x":"y","b":{"c":{"d":{"g":"h"}}}}}`),
			[]byte(`{"a":{"x":"y","b":{"c":{"d":{"i":"j"}}}}}`),
		},
		nil,
	},
	{
		"objects overwriting set key",
		procExpand{
			process: process{
				Key:    "a.b",
				SetKey: "a",
			},
		},
		[]byte(`{"a":{"b":[{"c":"d"},{"e":"f"}],"x":"y"}}`),
		[][]byte{
			[]byte(`{"a":{"c":"d"}}`),
			[]byte(`{"a":{"e":"f"}}`),
		},
		nil,
	},
	{
		"data array",
		procExpand{},
		[]byte(`[{"a":"b"},{"c":"d"}]`),
		[][]byte{
			[]byte(`{"a":"b"}`),
			[]byte(`{"c":"d"}`),
		},
		nil,
	},
}

func TestExpand(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range expandTests {
		var _ Batcher = test.proc

		slice := make([]config.Capsule, 1)
		capsule.SetData(test.test)
		slice[0] = capsule

		result, err := test.proc.Batch(ctx, slice...)
		if err != nil {
			t.Error(err)
		}

		for i, res := range result {
			expected := test.expected[i]
			if !bytes.Equal(expected, res.Data()) {
				t.Errorf("expected %s, got %s", expected, res)
			}
		}
	}
}

func benchmarkExpand(b *testing.B, slicer procExpand, slice []config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = slicer.Batch(ctx, slice...)
	}
}

func BenchmarkExpand(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range expandTests {
		slice := make([]config.Capsule, 1)
		capsule.SetData(test.test)
		slice[0] = capsule

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkExpand(b, test.proc, slice)
			},
		)
	}
}
