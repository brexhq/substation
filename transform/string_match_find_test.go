package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &stringMatchFind{}

var stringMatchFindTests = []struct {
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
				"pattern": "^([^@]*)@.*$",
			},
		},
		[]byte(`b@c`),
		[][]byte{
			[]byte(`b`),
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
				"pattern": "^([^@]*)@.*$",
			},
		},
		[]byte(`{"a":"b@c"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
}

func TestStringMatchFind(t *testing.T) {
	ctx := context.TODO()
	for _, test := range stringMatchFindTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStringMatchFind(ctx, test.cfg)
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

func benchmarkStringMatchFind(b *testing.B, tf *stringMatchFind, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStringMatchFind(b *testing.B) {
	for _, test := range stringMatchFindTests {
		tf, err := newStringMatchFind(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStringMatchFind(b, tf, test.test)
			},
		)
	}
}
