package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &numberMathAddition{}

var numberMathAdditionTests = []struct {
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
			[]byte(`8`),
		},
	},
	{
		"data",
		config.Config{},
		[]byte(`[0.123456789,10]`),
		[][]byte{
			[]byte(`10.123456789`),
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
			},
		},
		[]byte(`{"a":[6,2]}`),
		[][]byte{
			[]byte(`{"a":8}`),
		},
	},
	{
		"object",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":[0.123456789,10]}`),
		[][]byte{
			[]byte(`{"a":10.123456789}`),
		},
	},
}

func TestNumberMathAddition(t *testing.T) {
	ctx := context.TODO()
	for _, test := range numberMathAdditionTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newNumberMathAddition(ctx, test.cfg)
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

func benchmarkNumberMathAddition(b *testing.B, tf *numberMathAddition, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkNumberMathAddition(b *testing.B) {
	for _, test := range numberMathAdditionTests {
		tf, err := newNumberMathAddition(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNumberMathAddition(b, tf, test.test)
			},
		)
	}
}
