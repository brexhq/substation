package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &stringSplit{}

var stringSplitTests = []struct {
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
					"source_key": "a",
					"target_key": "a",
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

func TestStringSplit(t *testing.T) {
	ctx := context.TODO()
	for _, test := range stringSplitTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStringSplit(ctx, test.cfg)
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

func benchmarkStringSplit(b *testing.B, tf *stringSplit, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarStrkSplit(b *testing.B) {
	for _, test := range stringSplitTests {
		p, err := newStringSplit(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStringSplit(b, p, test.test)
			},
		)
	}
}

func FuzzTestStringSplit(f *testing.F) {
	testcases := [][]byte{
		[]byte(`b.c.d`),
		[]byte(`{"a":"b.c.d"}`),
		[]byte(`a.b.c.d.e.f.g`),
		[]byte(`{"a":"a.b.c.d.e.f.g"}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Test with default settings
		tf, err := newStringSplit(ctx, config.Config{
			Settings: map[string]interface{}{
				"separator": ".",
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}

		// Test with object settings
		tf, err = newStringSplit(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"separator": ".",
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
