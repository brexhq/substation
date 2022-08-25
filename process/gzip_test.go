package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var gzipTests = []struct {
	name     string
	proc     Gzip
	test     []byte
	expected []byte
	err      error
}{
	{
		"from",
		Gzip{
			Options: GzipOptions{
				Direction: "from",
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		[]byte(`foo`),
		nil,
	},
	{
		"to",
		Gzip{
			Options: GzipOptions{
				Direction: "to",
			},
		},
		[]byte(`foo`),
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		nil,
	},
}

func TestGzip(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()

	for _, test := range gzipTests {
		cap.SetData(test.test)

		result, err := test.proc.Apply(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result.GetData(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, result.GetData())
			t.Fail()
		}
	}
}

func benchmarkGzip(b *testing.B, applicator Gzip, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkGzip(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range gzipTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkGzip(b, test.proc, cap)
			},
		)
	}
}
