package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modGroup{}

var modGroupTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"tuples",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
			},
		},
		[]byte(`{"a":[["b","c"],[1,2]]}`),
		[][]byte{
			[]byte(`{"a":[["b",1],["c",2]]}`),
		},
		nil,
	},
	{
		"objects",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"keys": []string{"d.e", "f"},
			},
		},
		[]byte(`{"a":[["b","c"],[1,2]]}`),
		[][]byte{
			[]byte(`{"a":[{"d":{"e":"b"},"f":1},{"d":{"e":"c"},"f":2}]}`),
		},
		nil,
	},
}

func TestModGroup(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modGroupTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModGroup(ctx, test.cfg)
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

func benchmarkModGroup(b *testing.B, tf *modGroup, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModGroup(b *testing.B) {
	for _, test := range modGroupTests {
		tf, err := newModGroup(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModGroup(b, tf, test.test)
			},
		)
	}
}
