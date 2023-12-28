package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &arrayGroup{}

var arrayGroupTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"tuples",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"src_key": "a",
					"dst_key": "a",
				},
			},
		},
		[]byte(`{"a":[["b","c"],[1,2]]}`),
		[][]byte{
			[]byte(`{"a":[["b",1],["c",2]]}`),
		},
	},
	{
		"objects",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"src_key": "a",
					"dst_key": "a",
				},
				"group_keys": []string{"d.e", "f"},
			},
		},
		[]byte(`{"a":[["b","c"],[1,2]]}`),
		[][]byte{
			[]byte(`{"a":[{"d":{"e":"b"},"f":1},{"d":{"e":"c"},"f":2}]}`),
		},
	},
}

func TestArrayGroup(t *testing.T) {
	ctx := context.TODO()
	for _, test := range arrayGroupTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newArrayGroup(ctx, test.cfg)
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

func benchmarkArrayGroup(b *testing.B, tf *arrayGroup, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkArrayGroup(b *testing.B) {
	for _, test := range arrayGroupTests {
		tf, err := newArrayGroup(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkArrayGroup(b, tf, test.test)
			},
		)
	}
}
