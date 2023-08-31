package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modConvert{}

var modConvertTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"bool true",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "bool",
			},
		},
		[]byte(`{"a":"true"}`),
		[][]byte{
			[]byte(`{"a":true}`),
		},
		nil,
	},
	{
		"bool false",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "bool",
			},
		},
		[]byte(`{"a":"false"}`),
		[][]byte{
			[]byte(`{"a":false}`),
		},
		nil,
	},
	{
		"int",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "int",
			},
		},
		[]byte(`{"a":"-1"}`),
		[][]byte{
			[]byte(`{"a":-1}`),
		},
		nil,
	},
	{
		"float",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "float",
			},
		},
		[]byte(`{"a":"1.2"}`),
		[][]byte{
			[]byte(`{"a":1.2}`),
		},
		nil,
	},
	{
		"uint",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "uint",
			},
		},
		[]byte(`{"a":"1"}`),
		[][]byte{
			[]byte(`{"a":1}`),
		},
		nil,
	},
	{
		"string",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "string",
			},
		},
		[]byte(`{"a":1}`),
		[][]byte{
			[]byte(`{"a":"1"}`),
		},
		nil,
	},
	{
		"int",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type": "int",
			},
		},
		[]byte(`{"a":1.2}`),
		[][]byte{
			[]byte(`{"a":1}`),
		},
		nil,
	},
}

func TestModConvert(t *testing.T) {
	ctx := context.TODO()

	for _, test := range modConvertTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModConvert(ctx, test.cfg)
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

func benchmarkModConvert(b *testing.B, tf *modConvert, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModConvert(b *testing.B) {
	for _, test := range modConvertTests {
		tf, err := newModConvert(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModConvert(b, tf, test.test)
			},
		)
	}
}
