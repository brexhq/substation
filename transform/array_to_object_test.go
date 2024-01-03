package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &arrayToObject{}

var arrayToObjectTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{},
		},
		[]byte(`[["b","c"],[1,2]]`),
		[][]byte{
			[]byte(`{"b":1,"c":2}`),
		},
	},
	{
		"data multi_value",
		config.Config{
			Settings: map[string]interface{}{},
		},
		[]byte(`[["b","c"],[1,2],[3,4]]`),
		[][]byte{
			[]byte(`{"b":[1,3],"c":[2,4]}`),
		},
	},
	{
		"data multi_value keys",
		config.Config{
			Settings: map[string]interface{}{
				"object_keys": []string{"x", "y"},
			},
		},
		[]byte(`[["b","c"],[1,2],[3,4]]`),
		[][]byte{
			[]byte(`{"x":["b",1,3],"y":["c",2,4]}`),
		},
	},
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":[["b","c"],[1,2]]}`),
		[][]byte{
			[]byte(`{"a":{"b":1,"c":2}}`),
		},
	},
	{
		"object multi_value",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":[["b","c"],[1,2],[3,4]]}`),
		[][]byte{
			[]byte(`{"a":{"b":[1,3],"c":[2,4]}}`),
		},
	},
	{
		"object multi_value keys",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"object_keys": []string{"x", "y"},
			},
		},
		[]byte(`{"a":[["b","c"],[1,2],[3,4]]}`),
		[][]byte{
			[]byte(`{"a":{"x":["b",1,3],"y":["c",2,4]}}`),
		},
	},
}

func TestArrayToObject(t *testing.T) {
	ctx := context.TODO()
	for _, test := range arrayToObjectTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newArrayToObject(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.test)
			result, err := tf.Transform(ctx, msg)
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

func benchmarkArrayToObject(b *testing.B, tf *arrayToObject, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkArrayToObject(b *testing.B) {
	for _, test := range arrayToObjectTests {
		tf, err := newArrayToObject(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkArrayToObject(b, tf, test.test)
			},
		)
	}
}
