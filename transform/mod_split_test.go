package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modSplit{}

var modSplitTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"split",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"separator": ".",
			},
		},
		[]byte(`{"a":"b.c.d"}`),
		[][]byte{
			[]byte(`{"a":["b","c","d"]}`),
		},
		nil,
	},
}

func TestModSplit(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modSplitTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModSplit(ctx, test.cfg)
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

func benchmarkModSplit(b *testing.B, tf *modSplit, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModSplit(b *testing.B) {
	for _, test := range modSplitTests {
		p, err := newModSplit(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModSplit(b, p, test.test)
			},
		)
	}
}
