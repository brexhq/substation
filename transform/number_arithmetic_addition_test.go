package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var numberArithmeticAdditionTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{},
		[]byte(`[1,3]`),
		[][]byte{
			[]byte(`4`),
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
		[]byte(`{"a":[1,3]}`),
		[][]byte{
			[]byte(`{"a":4}`),
		},
	},
}

func TestNumberArithmeticAddition(t *testing.T) {
	ctx := context.TODO()
	for _, test := range numberArithmeticAdditionTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newNumberArithmeticAddition(ctx, test.cfg)
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

func benchmarkNumberArithmeticAddition(b *testing.B, tf *numberArithmeticAddition, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkNumberArithmeticAddition(b *testing.B) {
	for _, test := range numberArithmeticAdditionTests {
		tf, err := newNumberArithmeticAddition(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNumberArithmeticAddition(b, tf, test.test)
			},
		)
	}
}
