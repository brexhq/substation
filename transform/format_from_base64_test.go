package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var formatFromBase64Tests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	// data tests
	{
		"data",
		config.Config{},
		[]byte(`Yg==`),
		[][]byte{
			[]byte(`b`),
		},
		nil,
	},
	// object tests
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
			},
		},
		[]byte(`{"a":"Yg=="}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
		nil,
	},
}

func TestFormatFromBase64(t *testing.T) {
	ctx := context.TODO()
	for _, test := range formatFromBase64Tests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newFormatFromBase64(ctx, test.cfg)
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

func benchmarkFormatFromBase64(b *testing.B, tf *formatFromBase64, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkFormatFromBase64(b *testing.B) {
	for _, test := range formatFromBase64Tests {
		tf, err := newFormatFromBase64(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkFormatFromBase64(b, tf, test.test)
			},
		)
	}
}
