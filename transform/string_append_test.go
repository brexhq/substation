package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &stringAppend{}

var stringAppendTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"suffix": "c",
			},
		},
		[]byte(`ab`),
		[][]byte{
			[]byte(`abc`),
		},
	},
	// object tests
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"suffix": "c",
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"bc"}`),
		},
	},
	{
		"object_dynamic_suffix",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "input",
					"target_key": "output",
				},
				"suffix_key": "suffix",
			},
		},
		[]byte(`{"input":"hello","suffix":"_world"}`),
		[][]byte{
			[]byte(`{"input":"hello","suffix":"_world","output":"hello_world"}`),
		},
	},
}

func TestStringAppend(t *testing.T) {
	ctx := context.TODO()
	for _, test := range stringAppendTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStringAppend(ctx, test.cfg)
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

func benchmarkStringAppend(b *testing.B, tf *stringAppend, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStringAppend(b *testing.B) {
	for _, test := range stringAppendTests {
		tf, err := newStringAppend(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStringAppend(b, tf, test.test)
			},
		)
	}
}

func FuzzTestStringAppend(f *testing.F) {
	testcases := [][]byte{
		[]byte(`ab`),
		[]byte(`{"a":"b"}`),
		[]byte(``),
		[]byte(`{"a":""}`),
		[]byte(`{"a":123}`),
		[]byte(`{"input":"hello","suffix":"_world"}`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Test with default settings
		tf, err := newStringAppend(ctx, config.Config{
			Settings: map[string]interface{}{
				"suffix": "c",
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}

		// Test with object settings
		tf, err = newStringAppend(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"suffix": "c",
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}

		// Test with dynamic suffix settings
		tf, err = newStringAppend(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "input",
					"target_key": "output",
				},
				"suffix_key": "suffix",
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
