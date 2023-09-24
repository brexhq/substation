package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var numberArithmeticDivTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data",
		config.Config{},
		[]byte(`[6,2]`),
		[][]byte{
			[]byte(`3`),
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
		[]byte(`{"a":[6,2]}`),
		[][]byte{
			[]byte(`{"a":3}`),
		},
	},
}

func TestDiv(t *testing.T) {
	ctx := context.TODO()
	for _, test := range numberArithmeticDivTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newNumberArithmeticDiv(ctx, test.cfg)
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

func benchmarkNumberArithmeticDiv(b *testing.B, tf *numberArithmeticDiv, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkNumberArithmeticDiv(b *testing.B) {
	for _, test := range numberArithmeticDivTests {
		tf, err := newNumberArithmeticDiv(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNumberArithmeticDiv(b, tf, test.test)
			},
		)
	}
}