package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &arrayZip{}

var arrayZipTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"array",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"src_key": "a",
					"dst_key": "a",
				},
			},
		},
		[]byte(`{"a":[["b","c"],[1,2]]}`),
		[][]byte{
			[]byte(`{"a":[["b",1],["c",2]]}`),
		},
	},
	{
		"array as_object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"src_key": "a",
					"dst_key": "a",
				},
				"as_object": true,
			},
		},
		[]byte(`{"a":[["b","c"],[1,2]]}`),
		[][]byte{
			[]byte(`{"a":{"b":1,"c":2}}`),
		},
	},
	{
		"array as_object multi_value",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"src_key": "a",
					"dst_key": "a",
				},
				"as_object": true,
			},
		},
		[]byte(`{"a":[["b","c"],[1,2],[3,4]]}`),
		[][]byte{
			[]byte(`{"a":{"b":[1,3],"c":[2,4]}}`),
		},
	},
	{
		"array as_object multi_value with_keys",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"src_key": "a",
					"dst_key": "a",
				},
				"as_object": true,
				"with_keys": []string{"x", "y"},
			},
		},
		[]byte(`{"a":[["b","c"],[1,2],[3,4]]}`),
		[][]byte{
			[]byte(`{"a":{"x":["b",1,3],"y":["c",2,4]}}`),
		},
	},
}

func TestArrayZip(t *testing.T) {
	ctx := context.TODO()
	for _, test := range arrayZipTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newArrayZip(ctx, test.cfg)
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

func benchmarkArrayZip(b *testing.B, tf *arrayZip, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkArrayZip(b *testing.B) {
	for _, test := range arrayZipTests {
		tf, err := newArrayZip(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkArrayZip(b, tf, test.test)
			},
		)
	}
}
