package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &formatToBase64{}

var formatToBase64Tests = []struct {
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
		[]byte(`b`),
		[][]byte{
			[]byte(`Yg==`),
		},
		nil,
	},
	// object tests
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"src_key": "a",
					"dst_key": "a",
				},
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"Yg=="}`),
		},
		nil,
	},
}

func TestFormatToBase64(t *testing.T) {
	ctx := context.TODO()
	for _, test := range formatToBase64Tests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newFormatToBase64(ctx, test.cfg)
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

func benchmarkFormatToBase64(b *testing.B, tf *formatToBase64, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkFormatToBase64Encode(b *testing.B) {
	for _, test := range formatToBase64Tests {
		tf, err := newFormatToBase64(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkFormatToBase64(b, tf, test.test)
			},
		)
	}
}
