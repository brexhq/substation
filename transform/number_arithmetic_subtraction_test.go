package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var numberArithmeticSubtractionTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{},
		[]byte(`[3,1]`),
		[][]byte{
			[]byte(`2`),
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
		[]byte(`{"a":[3,1]}`),
		[][]byte{
			[]byte(`{"a":2}`),
		},
	},
}

func TestNumberArithmeticSubtraction(t *testing.T) {
	ctx := context.TODO()
	for _, test := range numberArithmeticSubtractionTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newNumberArithmeticSubtraction(ctx, test.cfg)
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

func benchmarkNumberArithmeticSubtraction(b *testing.B, tf *numberArithmeticSubtraction, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkNumberArithmeticSubtraction(b *testing.B) {
	for _, test := range numberArithmeticSubtractionTests {
		tf, err := newNumberArithmeticSubtraction(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNumberArithmeticSubtraction(b, tf, test.test)
			},
		)
	}
}
