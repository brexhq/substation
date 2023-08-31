package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modBase64{}

var modBase64Tests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"data from",
		config.Config{
			Settings: map[string]interface{}{
				"direction": "from",
			},
		},
		[]byte(`Yg==`),
		[][]byte{
			[]byte(`b`),
		},
		nil,
	},
	{
		"data to",
		config.Config{
			Settings: map[string]interface{}{
				"direction": "to",
			},
		},
		[]byte(`b`),
		[][]byte{
			[]byte(`Yg==`),
		},
		nil,
	},
	{
		"object from",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"direction": "from",
			},
		},
		[]byte(`{"a":"Yg=="}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
		nil,
	},
	{
		"object to",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"direction": "to",
			},
		},
		[]byte(`{"a":"b"}`),
		[][]byte{
			[]byte(`{"a":"Yg=="}`),
		},
		nil,
	},
}

func TestModBase64(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modBase64Tests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModBase64(ctx, test.cfg)
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

func benchmarkmodBase64(b *testing.B, tf *modBase64, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModBase64(b *testing.B) {
	for _, test := range modBase64Tests {
		tf, err := newModBase64(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkmodBase64(b, tf, test.test)
			},
		)
	}
}
