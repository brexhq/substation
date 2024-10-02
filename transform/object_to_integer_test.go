package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &objectToInteger{}

var objectToIntegerTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"float to_int",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":1.1}`),
		[][]byte{
			[]byte(`{"a":1}`),
		},
	},
	{
		"str to_int",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":"-1"}`),
		[][]byte{
			[]byte(`{"a":-1}`),
		},
	},
}

func TestObjectToInteger(t *testing.T) {
	ctx := context.TODO()

	for _, test := range objectToIntegerTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjectToInteger(ctx, test.cfg)
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

func benchmarkObjectToInteger(b *testing.B, tf *objectToInteger, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjectToInteger(b *testing.B) {
	for _, test := range objectToIntegerTests {
		tf, err := newObjectToInteger(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjectToInteger(b, tf, test.test)
			},
		)
	}
}

func FuzzTestObjectToInteger(f *testing.F) {
	testcases := [][]byte{
		[]byte(`{"a":"1.1"}`),
		[]byte(`{"a":"0"}`),
		[]byte(`{"a":"-1.1"}`),
		[]byte(`{"a":"1234567890"}`),
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
		tf, err := newObjectToInteger(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"type": "integer",
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
