package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modDelete{}

var modDeleteTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"string",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "baz",
				},
			},
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "baz",
				},
			},
		},
		[]byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
		[][]byte{
			[]byte(`{"foo":"bar"}`),
		},
		nil,
	},
}

func TestModRemove(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modDeleteTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModRemove(ctx, test.cfg)
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

func benchmarkModRemove(b *testing.B, tf *modDelete, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModRemove(b *testing.B) {
	for _, test := range modDeleteTests {
		tf, err := newModRemove(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModRemove(b, tf, test.test)
			},
		)
	}
}
