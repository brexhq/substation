package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var objToFloatTests = []struct {
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
					"key":     "a",
					"set_key": "a",
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

func TestToFloat(t *testing.T) {
	ctx := context.TODO()

	for _, test := range objToFloatTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjToFloat(ctx, test.cfg)
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

func benchmarkObjToFloat(b *testing.B, tf *objToFloat, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjToFloat(b *testing.B) {
	for _, test := range objToFloatTests {
		tf, err := newObjToFloat(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjToFloat(b, tf, test.test)
			},
		)
	}
}
