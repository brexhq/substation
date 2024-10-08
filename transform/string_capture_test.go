package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &stringCapture{}

var stringCaptureTests = []struct {
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
				"pattern": "^([^@]*)@.*$",
			},
		},
		[]byte(`b@c`),
		[][]byte{
			[]byte(`b`),
		},
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"count":   3,
				"pattern": "(.{1})",
			},
		},
		[]byte(`bcd`),
		[][]byte{
			[]byte(`["b","c","d"]`),
		},
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"count":   1,
				"pattern": "(.{1})",
			},
		},
		[]byte(`bcd`),
		[][]byte{
			[]byte(`["b"]`),
		},
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"pattern": "(?P<b>[a-zA-Z]+) (?P<d>[a-zA-Z]+)",
			},
		},
		[]byte(`c e`),
		[][]byte{
			[]byte(`{"b":"c","d":"e"}`),
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
				"pattern": "^([^@]*)@.*$",
			},
		},
		[]byte(`{"a":"b@c"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
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
				"count":   3,
				"pattern": "(.{1})",
			},
		},
		[]byte(`{"a":"bcd"}`),
		[][]byte{
			[]byte(`{"a":["b","c","d"]}`),
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
				"count":   1,
				"pattern": "(.{1})",
			},
		},
		[]byte(`{"a":"bcd"}`),
		[][]byte{
			[]byte(`{"a":["b"]}`),
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
				"pattern": "(?P<b>[a-zA-Z]+) (?P<d>[a-zA-Z]+)",
			},
		},
		[]byte(`{"a":"c e"}`),
		[][]byte{
			[]byte(`{"a":{"b":"c","d":"e"}}`),
		},
	},
}

func TestStringCapture(t *testing.T) {
	ctx := context.TODO()
	for _, test := range stringCaptureTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStringCapture(ctx, test.cfg)
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

func benchmarkStringCapture(b *testing.B, tf *stringCapture, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStringCapture(b *testing.B) {
	for _, test := range stringCaptureTests {
		tf, err := newStringCapture(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStringCapture(b, tf, test.test)
			},
		)
	}
}

func FuzzTestStringCapture(f *testing.F) {
	testcases := [][]byte{
		[]byte(`b@c`),
		[]byte(`bcd`),
		[]byte(`example@example.com`),
		[]byte(`123-456-7890`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Test with default settings
		tf, err := newStringCapture(ctx, config.Config{
			Settings: map[string]interface{}{
				"pattern": "^([^@]*)@.*$",
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}

		// Test with count settings
		tf, err = newStringCapture(ctx, config.Config{
			Settings: map[string]interface{}{
				"count":   3,
				"pattern": "(.{1})",
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
