package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &formatToGzip{}

var formatToGzipTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"data",
		config.Config{},
		[]byte(`foo`),
		[][]byte{
			{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		},
		nil,
	},
}

func TestFormatToGzip(t *testing.T) {
	ctx := context.TODO()
	for _, test := range formatToGzipTests {
		t.Run(test.name, func(t *testing.T) {
			msg := message.New().SetData(test.test)

			tf, err := newFormatToGzip(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

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

func benchmarkFormatToGzip(b *testing.B, tf *formatToGzip, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkFormatToGzip(b *testing.B) {
	for _, test := range formatToGzipTests {
		tf, err := newFormatToGzip(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkFormatToGzip(b, tf, test.test)
			},
		)
	}
}

func FuzzTestFormatToGzip(f *testing.F) {
	testcases := [][]byte{
		[]byte(`b`),
		[]byte(`{"a":"b"}`),
		[]byte(``),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		// Test with default settings
		tf, err := newFormatToGzip(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}

		// Test with object settings
		tf, err = newFormatToGzip(ctx, config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "a",
					"target_key": "a",
				},
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
