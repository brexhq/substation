package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &numberMaximum{}

var numberMaximumTests = []struct {
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
				"value": 1,
			},
		},
		[]byte(`0`),
		[][]byte{
			[]byte(`1`),
		},
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"value": -1,
			},
		},
		[]byte(`0`),
		[][]byte{
			[]byte(`0`),
		},
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"value": -1.1,
			},
		},
		[]byte(`0.1`),
		[][]byte{
			[]byte(`0.1`),
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
				"value": 1,
			},
		},
		[]byte(`{"a":0}`),
		[][]byte{
			[]byte(`{"a":1}`),
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
				"value": -1,
			},
		},
		[]byte(`{"a":0}`),
		[][]byte{
			[]byte(`{"a":0}`),
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
				"value": -1.1,
			},
		},
		[]byte(`{"a":0.1}`),
		[][]byte{
			[]byte(`{"a":0.1}`),
		},
	},
}

func TestNumberMaximum(t *testing.T) {
	ctx := context.TODO()
	for _, test := range numberMaximumTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newNumberMaximum(ctx, test.cfg)
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

func benchmarkNumberMaximum(b *testing.B, tf *numberMaximum, data []byte) {
	ctx := context.TODO()
	msg := message.New().SetData(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkNumberMaximum(b *testing.B) {
	for _, test := range numberMaximumTests {
		tf, err := newNumberMaximum(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNumberMaximum(b, tf, test.test)
			},
		)
	}
}
