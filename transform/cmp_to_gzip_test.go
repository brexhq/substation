package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var cmpToGzipTests = []struct {
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

func TestCmpToGzip(t *testing.T) {
	ctx := context.TODO()
	for _, test := range cmpToGzipTests {
		t.Run(test.name, func(t *testing.T) {
			msg := message.New().SetData(test.test)

			tf, err := newCmpToGzip(ctx, test.cfg)
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

func benchmarkCmpToGzip(b *testing.B, tf *cmpToGzip, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkCmpToGzip(b *testing.B) {
	for _, test := range cmpToGzipTests {
		tf, err := newCmpToGzip(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkCmpToGzip(b, tf, test.test)
			},
		)
	}
}
