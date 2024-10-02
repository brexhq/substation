package transform

import (
	"context"
	"slices"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &formatFromZip{}

var formatFromZipTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []string
}{
	{
		"data",
		config.Config{},
		// This is a zip file containing two files with the contents "bar" and "qux" (no newlines).
		[]byte{80, 75, 3, 4, 10, 0, 0, 0, 0, 0, 57, 63, 251, 88, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 28, 0, 116, 109, 112, 47, 85, 84, 9, 0, 3, 238, 10, 165, 102, 239, 10, 165, 102, 117, 120, 11, 0, 1, 4, 246, 1, 0, 0, 4, 20, 0, 0, 0, 80, 75, 3, 4, 10, 0, 0, 0, 0, 0, 55, 63, 251, 88, 200, 175, 228, 166, 3, 0, 0, 0, 3, 0, 0, 0, 11, 0, 28, 0, 116, 109, 112, 47, 98, 97, 122, 46, 116, 120, 116, 85, 84, 9, 0, 3, 233, 10, 165, 102, 234, 10, 165, 102, 117, 120, 11, 0, 1, 4, 246, 1, 0, 0, 4, 20, 0, 0, 0, 113, 117, 120, 80, 75, 3, 4, 10, 0, 0, 0, 0, 0, 44, 63, 251, 88, 170, 140, 255, 118, 3, 0, 0, 0, 3, 0, 0, 0, 11, 0, 28, 0, 116, 109, 112, 47, 102, 111, 111, 46, 116, 120, 116, 85, 84, 9, 0, 3, 212, 10, 165, 102, 214, 10, 165, 102, 117, 120, 11, 0, 1, 4, 246, 1, 0, 0, 4, 20, 0, 0, 0, 98, 97, 114, 80, 75, 1, 2, 30, 3, 10, 0, 0, 0, 0, 0, 57, 63, 251, 88, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 24, 0, 0, 0, 0, 0, 0, 0, 16, 0, 237, 65, 0, 0, 0, 0, 116, 109, 112, 47, 85, 84, 5, 0, 3, 238, 10, 165, 102, 117, 120, 11, 0, 1, 4, 246, 1, 0, 0, 4, 20, 0, 0, 0, 80, 75, 1, 2, 30, 3, 10, 0, 0, 0, 0, 0, 55, 63, 251, 88, 200, 175, 228, 166, 3, 0, 0, 0, 3, 0, 0, 0, 11, 0, 24, 0, 0, 0, 0, 0, 1, 0, 0, 0, 164, 129, 62, 0, 0, 0, 116, 109, 112, 47, 98, 97, 122, 46, 116, 120, 116, 85, 84, 5, 0, 3, 233, 10, 165, 102, 117, 120, 11, 0, 1, 4, 246, 1, 0, 0, 4, 20, 0, 0, 0, 80, 75, 1, 2, 30, 3, 10, 0, 0, 0, 0, 0, 44, 63, 251, 88, 170, 140, 255, 118, 3, 0, 0, 0, 3, 0, 0, 0, 11, 0, 24, 0, 0, 0, 0, 0, 1, 0, 0, 0, 164, 129, 134, 0, 0, 0, 116, 109, 112, 47, 102, 111, 111, 46, 116, 120, 116, 85, 84, 5, 0, 3, 212, 10, 165, 102, 117, 120, 11, 0, 1, 4, 246, 1, 0, 0, 4, 20, 0, 0, 0, 80, 75, 5, 6, 0, 0, 0, 0, 3, 0, 3, 0, 236, 0, 0, 0, 206, 0, 0, 0, 0, 0},
		[]string{
			"bar",
			"qux",
		},
	},
}

func TestFormatFromZip(t *testing.T) {
	ctx := context.TODO()
	for _, test := range formatFromZipTests {
		t.Run(test.name, func(t *testing.T) {
			msg := message.New().SetData(test.test)

			tf, err := newFormatFromZip(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			msgs, err := tf.Transform(ctx, msg)
			if err != nil {
				t.Error(err)
			}

			// The order of the output is not guaranteed, so we need to
			// check that the expected values are present anywhere in the
			// result.
			var results []string
			for _, m := range msgs {
				results = append(results, string(m.Data()))
			}

			for _, r := range results {
				if !slices.Contains(test.expected, r) {
					t.Errorf("expected %s, got %s", test.expected, r)
				}
			}
		})
	}
}

func benchmarkFormatFromZip(b *testing.B, tf *formatFromZip, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkFormatFromZip(b *testing.B) {
	for _, test := range formatFromZipTests {
		tf, err := newFormatFromZip(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkFormatFromZip(b, tf, test.test)
			},
		)
	}
}

func FuzzTestFormatFromZip(f *testing.F) {
	testcases := [][]byte{
		{80, 75, 3, 4, 10, 0, 0, 0, 0, 0, 57, 63, 251, 88, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 28, 0, 116, 109, 112, 47, 85, 84, 9, 0, 3, 238, 10, 165, 102, 239, 10, 165, 102, 117, 120, 11, 0, 1, 4, 246, 1, 0, 0, 4, 20, 0, 0, 0, 80, 75, 3, 4, 10, 0, 0, 0, 0, 0, 55, 63, 251, 88, 200, 175, 228, 166, 3, 0, 0, 0, 3, 0, 0, 0, 11, 0, 28, 0, 116, 109, 112, 47, 98, 97, 122, 46, 116, 120, 116, 85, 84, 9, 0, 3, 233, 10, 165, 102, 234, 10, 165, 102, 117, 120, 11, 0, 1, 4, 246, 1, 0, 0, 4, 20, 0, 0, 0, 113, 117, 120},
		{80, 75, 3, 4, 10, 0, 0, 0, 0, 0, 44, 63, 251, 88, 170, 140, 255, 118, 3, 0, 0, 0, 3, 0, 0, 0, 11, 0, 28, 0, 116, 109, 112, 47, 102, 111, 111, 46, 116, 120, 116, 85, 84, 9, 0, 3, 212, 10, 165, 102, 214, 10, 165, 102, 117, 120, 11, 0, 1, 4, 246, 1, 0, 0, 4, 20, 0, 0, 0, 98, 97, 114},
		{},
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		msg := message.New().SetData(data)

		tf, err := newFormatFromZip(ctx, config.Config{})
		if err != nil {
			return
		}

		_, err = tf.Transform(ctx, msg)
		if err != nil {
			return
		}
	})
}
