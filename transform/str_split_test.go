package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var strSplitTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"separator": ".",
			},
		},
		[]byte(`b.c.d`),
		[][]byte{
			[]byte(`["b","c","d"]`),
		},
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
				"separator": ".",
			},
		},
		[]byte(`{"a":"b.c.d"}`),
		[][]byte{
			[]byte(`{"a":["b","c","d"]}`),
		},
	},
}

func TestStrSplit(t *testing.T) {
	ctx := context.TODO()
	for _, test := range strSplitTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStrSplit(ctx, test.cfg)
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

func benchmarkStrSplit(b *testing.B, tf *strSplit, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarStrkSplit(b *testing.B) {
	for _, test := range strSplitTests {
		p, err := newStrSplit(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStrSplit(b, p, test.test)
			},
		)
	}
}
