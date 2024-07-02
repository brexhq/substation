package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &numberMinimum{}

var numberMinimumTests = []struct {
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
			[]byte(`0`),
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
			[]byte(`-1`),
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
			[]byte(`-1.1`),
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
				"value": -1,
			},
		},
		[]byte(`{"a":0}`),
		[][]byte{
			[]byte(`{"a":-1}`),
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
			[]byte(`{"a":-1.1}`),
		},
	},
}

func TestNumberMinimum(t *testing.T) {
	ctx := context.TODO()
	for _, test := range numberMinimumTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newNumberMinimum(ctx, test.cfg)
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

func benchmarkNumberMinimum(b *testing.B, tf *numberMinimum, data []byte) {
	ctx := context.TODO()
	msg := message.New().SetData(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkNumberMinimum(b *testing.B) {
	for _, test := range numberMinimumTests {
		tf, err := newNumberMinimum(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkNumberMinimum(b, tf, test.test)
			},
		)
	}
}
