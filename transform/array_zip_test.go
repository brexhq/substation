package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &arrayZip{}

var arrayZipTests = []struct {
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
			[]byte(`[["b",1],["c",2]]`),
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
			[]byte(`{"a":[["b",1],["c",2]]}`),
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

func FuzzTestArrayZip(f *testing.F) {
	testcases := [][]byte{
		[]byte(`[["b","c"],[1,2]]`),
		[]byte(`{"a":[["b","c"],[1,2]]}`),
		[]byte(`[["x","y","z"],[3,4,5]]`),
		[]byte(`{"a":[["x","y","z"],[3,4,5]]}`),
		[]byte(`[]`),
		[]byte(`{}`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Test with default settings
		tf, err := newArrayZip(ctx, config.Config{
			Settings: map[string]interface{}{},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}

		// Test with object settings
		tf, err = newArrayZip(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
