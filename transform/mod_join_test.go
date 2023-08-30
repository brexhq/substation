package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modJoin{}

var modJoinTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "x",
					"set_key": "x",
				},
				"separator": ".",
			},
		},
		[]byte(`{"x":["a","b","c"]}`),
		[][]byte{
			[]byte(`{"x":"a.b.c"}`),
		},
		nil,
	},
}

func TestModJoin(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modJoinTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModJoin(ctx, test.cfg)
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

func benchmarkModJoin(b *testing.B, tf *modJoin, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModJoin(b *testing.B) {
	for _, test := range modJoinTests {
		tf, err := newModJoin(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModJoin(b, tf, test.test)
			},
		)
	}
}
