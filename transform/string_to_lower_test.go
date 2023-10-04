package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &stringToLower{}

var stringToLowerTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{},
		[]byte(`B`),
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
			},
		},
		[]byte(`{"a":"B"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
}

func TestStringToLower(t *testing.T) {
	ctx := context.TODO()
	for _, test := range stringToLowerTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStringToLower(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msg := message.New().SetData(test.test)
			result, err := tf.Transform(ctx, msg)
			if err != nil {
				t.Error(err)
			}

			var r [][]byte
			for _, c := range result {
				r = append(r, c.Data())
			}

			if !reflect.DeepEqual(r, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, r)
			}
		})
	}
}

func benchmarkStringToLower(b *testing.B, tf *stringToLower, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStringToLower(b *testing.B) {
	for _, test := range stringToLowerTests {
		tf, err := newStringToLower(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStringToLower(b, tf, test.test)
			},
		)
	}
}
