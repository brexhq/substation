package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &stringReplace{}

var stringReplaceTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	// data tests
	{
		"data replace",
		config.Config{
			Settings: map[string]interface{}{
				"pattern":     "c",
				"replacement": "b",
			},
		},
		[]byte(`abc`),
		[][]byte{
			[]byte(`abb`),
		},
	},
	{
		"data remove",
		config.Config{
			Settings: map[string]interface{}{
				"pattern":     "c",
				"replacement": "",
			},
		},
		[]byte(`abc`),
		[][]byte{
			[]byte(`ab`),
		},
	},
	// object tests
	{
		"object replace",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"pattern":     "c",
				"replacement": "b",
			},
		},
		[]byte(`{"a":"bc"}`),
		[][]byte{
			[]byte(`{"a":"bb"}`),
		},
	},
	{
		"object remove",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"pattern": "c",
			},
		},
		[]byte(`{"a":"bc"}`),
		[][]byte{
			[]byte(`{"a":"b"}`),
		},
	},
}

func TestStringReplace(t *testing.T) {
	ctx := context.TODO()
	for _, test := range stringReplaceTests {
		t.Run(test.name, func(t *testing.T) {
			tf, err := newStringReplace(ctx, test.cfg)
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

func benchmarkStringReplace(b *testing.B, tf *stringReplace, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkStringReplace(b *testing.B) {
	for _, test := range stringReplaceTests {
		tf, err := newStringReplace(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkStringReplace(b, tf, test.test)
			},
		)
	}
}

func FuzzTestStringReplace(f *testing.F) {
	testcases := [][]byte{
		[]byte(`abc`),
		[]byte(`def`),
		[]byte(`ghi`),
		[]byte(`{"a":"abc"}`),
		[]byte(`{"a":"def"}`),
		[]byte(`{"a":"ghi"}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Test with default settings
		tf, err := newStringReplace(ctx, config.Config{
			Settings: map[string]interface{}{
				"pattern":     "c",
				"replacement": "b",
			},
		})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}

		// Test with object settings
		tf, err = newStringReplace(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
				"pattern":     "c",
				"replacement": "b",
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
