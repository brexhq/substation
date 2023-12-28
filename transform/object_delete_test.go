package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &objectDelete{}

var objectDeleteTests = []struct {
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
					"src_key": "c",
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
					"src_key": "c",
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
					"src_key": "c",
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
					"src_key": "c",
				},
			},
		},
		[]byte(`{"a":"b","c":1}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
}

func TestObjectDelete(t *testing.T) {
	ctx := context.TODO()
	for _, test := range objectDeleteTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjectDelete(ctx, test.cfg)
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

func benchmarkObjectDelete(b *testing.B, tf *objectDelete, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjectDelete(b *testing.B) {
	for _, test := range objectDeleteTests {
		tf, err := newObjectDelete(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjectDelete(b, tf, test.test)
			},
		)
	}
}
