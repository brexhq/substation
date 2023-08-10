package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procExpand{}

var procExpandTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"objects",
		config.Config{
			Settings: map[string]interface{}{
				"key": "a",
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
		config.Config{
			Settings: map[string]interface{}{
				"key": "a",
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
		config.Config{
			Settings: map[string]interface{}{
				"key": "a",
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
		config.Config{
			Settings: map[string]interface{}{
				"key":     "a",
				"set_key": "a",
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
		config.Config{
			Settings: map[string]interface{}{
				"key":     "a",
				"set_key": "a",
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
		config.Config{
			Settings: map[string]interface{}{
				"key":     "a.b",
				"set_key": "a.b.c.d",
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
		config.Config{
			Settings: map[string]interface{}{
				"key":     "a.b",
				"set_key": "a",
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
		config.Config{},
		[]byte(`[{"a":"b"},{"c":"d"}]`),
		[][]byte{
			[]byte(`{"a":"b"}`),
			[]byte(`{"c":"d"}`),
		},
		nil,
	},
}

func TestProcExpand(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procExpandTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcExpand(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, message)
			if err != nil {
				t.Error(err)
			}

			var data [][]byte
			for _, c := range result {
				data = append(data, c.Data())
			}

			if !reflect.DeepEqual(data, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, data)
			}
		})
	}
}

func benchmarkProcExpand(b *testing.B, tform *procExpand, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcExpand(b *testing.B) {
	for _, test := range procExpandTests {
		proc, err := newProcExpand(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcExpand(b, proc, test.test)
			},
		)
	}
}
