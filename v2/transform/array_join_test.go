package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &arrayJoin{}

var arrayJoinTests = []struct {
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
		[]byte(`["b","c","d"]`),
		[][]byte{
			[]byte(`b.c.d`),
		},
	},
	// object tests
	{
		"object from",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"separator": ".",
			},
		},
		[]byte(`{"a":["b","c","d"]}`),
		[][]byte{
			[]byte(`{"a":"b.c.d"}`),
		},
	},
}

func TestArrayJoin(t *testing.T) {
	ctx := context.TODO()
	for _, test := range arrayJoinTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newArrayJoin(ctx, test.cfg)
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

func benchmarkArrayJoin(b *testing.B, tf *arrayJoin, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkArrayJoin(b *testing.B) {
	for _, test := range arrayJoinTests {
		p, err := newArrayJoin(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkArrayJoin(b, p, test.test)
			},
		)
	}
}
