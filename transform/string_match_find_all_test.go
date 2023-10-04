package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &stringMatchFindAll{}

var stringMatchFindAllTests = []struct {
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
				"count":   3,
				"pattern": "(.{1})",
			},
		},
		[]byte(`bcd`),
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
				"count":   3,
				"pattern": "(.{1})",
			},
		},
		[]byte(`{"a":"bcd"}`),
		[][]byte{
			[]byte(`{"a":["b","c","d"]}`),
		},
	},
}

func TestStringMatchFindAll(t *testing.T) {
	ctx := context.TODO()
	for _, test := range stringMatchFindAllTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStringMatchFindAll(ctx, test.cfg)
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

func benchmarkStringMatchFindAll(b *testing.B, tf *stringMatchFindAll, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStringMatchFindAll(b *testing.B) {
	for _, test := range stringMatchFindAllTests {
		tf, err := newStringMatchFindAll(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStringMatchFindAll(b, tf, test.test)
			},
		)
	}
}
