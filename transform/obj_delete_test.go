package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var objDeleteTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "c",
				},
			},
		},
		[]byte(`{"a":"b","c":{"d":"e"}}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
	{
		"array",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "c",
				},
			},
		},
		[]byte(`{"a":"b","c":["d","e"]}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
	{
		"string",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "c",
				},
			},
		},
		[]byte(`{"a":"b","c":"d"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},

	{
		"int",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "c",
				},
			},
		},
		[]byte(`{"a":"b","c":1}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
}

func TestObjDelete(t *testing.T) {
	ctx := context.TODO()
	for _, test := range objDeleteTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjDelete(ctx, test.cfg)
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

func benchmarkObjDelete(b *testing.B, tf *objDelete, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjDelete(b *testing.B) {
	for _, test := range objDeleteTests {
		tf, err := newObjDelete(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjDelete(b, tf, test.test)
			},
		)
	}
}
