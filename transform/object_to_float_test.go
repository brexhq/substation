package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &objectToFloat{}

var objectToFloatTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"str to_float",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"type": "float",
			},
		},
		[]byte(`{"a":"1.1"}`),
		[][]byte{
			[]byte(`{"a":1.1}`),
		},
	},
}

func TestObjectToFloat(t *testing.T) {
	ctx := context.TODO()

	for _, test := range objectToFloatTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjectToFloat(ctx, test.cfg)
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

func benchmarkObjectToFloat(b *testing.B, tf *objectToFloat, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjectToFloat(b *testing.B) {
	for _, test := range objectToFloatTests {
		tf, err := newObjectToFloat(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjectToFloat(b, tf, test.test)
			},
		)
	}
}

func FuzzTestObjectToFloat(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"1.1"}`),
		[]byte(`{"a":"0.0"}`),
		[]byte(`{"a":"-1.1"}`),
		[]byte(`{"a":"1234567890.123456789"}`),
		[]byte(`{"a":"NaN"}`),
		[]byte(`{"a":"Infinity"}`),
		[]byte(`{"a":"-Infinity"}`),
		[]byte(`{"a":1}`),
		[]byte(`{"a":true}`),
		[]byte(`{"a":null}`),
		[]byte(`{"a":""}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Use a sample configuration for the transformer
		tf, err := newObjectToFloat(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"type": "float",
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
