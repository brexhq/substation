package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &objectToString{}

var objectToStringTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"bool to_str",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":true}`),
		[][]byte{
			[]byte(`{"a":"true"}`),
		},
	},
	{
		"float to_str",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":1.1}`),
		[][]byte{
			[]byte(`{"a":"1.1"}`),
		},
	},
	{
		"int to_str",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
			},
		},
		[]byte(`{"a":1}`),
		[][]byte{
			[]byte(`{"a":"1"}`),
		},
	},
}

func TestObjectToString(t *testing.T) {
	ctx := context.TODO()

	for _, test := range objectToStringTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newObjectToString(ctx, test.cfg)
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

func benchmarkObjectToString(b *testing.B, tf *objectToString, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkObjectToString(b *testing.B) {
	for _, test := range objectToStringTests {
		tf, err := newObjectToString(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkObjectToString(b, tf, test.test)
			},
		)
	}
}
