package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var cmpFromGzipTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
}{
	{
		"data",
		config.Config{},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		[][]byte{
			[]byte(`foo`),
		},
	},
}

func TestCmpFromGzip(t *testing.T) {
	ctx := context.TODO()
	for _, test := range cmpFromGzipTests {
		t.Run(test.name, func(t *testing.T) {
			msg := message.New().SetData(test.test)

			tf, err := newCmpFromGzip(ctx, test.cfg)
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

func benchmarkCmpFromGzip(b *testing.B, tf *cmpFromGzip, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkCmpFromGzip(b *testing.B) {
	for _, test := range cmpFromGzipTests {
		tf, err := newCmpFromGzip(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkCmpFromGzip(b, tf, test.test)
			},
		)
	}
}
