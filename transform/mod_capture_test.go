package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modCapture{}

var modCaptureTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"object find",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type":       "find",
				"expression": "^([^@]*)@.*$",
			},
		},
		[]byte(`{"a":"b@c"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
		nil,
	},
	{
		"object find_all",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type":       "find_all",
				"count":      3,
				"expression": "(.{1})",
			},
		},
		[]byte(`{"a":"bcd"}`),
		[][]byte{
			[]byte(`{"a":["b","c","d"]}`),
		},
		nil,
	},
	{
		"object named_group",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"type":       "named_group",
				"expression": "(?P<b>[a-zA-Z]+) (?P<d>[a-zA-Z]+)",
			},
		},
		[]byte(`{"a":"c e"}`),
		[][]byte{
			[]byte(`{"a":{"b":"c","d":"e"}}`),
		},
		nil,
	},
	{
		"data find",
		config.Config{
			Settings: map[string]interface{}{
				"type":       "find",
				"expression": "^([^@]*)@.*$",
			},
		},
		[]byte(`b@c`),
		[][]byte{
			[]byte(`b`),
		},
		nil,
	},
	{
		"data_named_group",
		config.Config{
			Settings: map[string]interface{}{
				"type":       "named_group",
				"expression": "(?P<b>[a-zA-Z]+) (?P<d>[a-zA-Z]+)",
			},
		},
		[]byte(`c e`),
		[][]byte{
			[]byte(`{"b":"c","d":"e"}`),
		},
		nil,
	},
}

func TestModCapture(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modCaptureTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newModCapture(ctx, test.cfg)
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

func benchmarkModCapture(b *testing.B, tf *modCapture, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModCapture(b *testing.B) {
	for _, test := range modCaptureTests {
		tf, err := newModCapture(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkModCapture(b, tf, test.test)
			},
		)
	}
}
