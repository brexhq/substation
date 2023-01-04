package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var gzipTests = []struct {
	name     string
	proc     procGzip
	test     []byte
	expected []byte
	err      error
}{
	{
		"from",
		procGzip{
			Options: procGzipOptions{
				Direction: "from",
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		[]byte(`foo`),
		nil,
	},
	{
		"to",
		procGzip{
			Options: procGzipOptions{
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
	capsule := config.NewCapsule()

	for _, test := range gzipTests {
		var _ Applier = test.proc
		var _ Batcher = test.proc

		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkGzip(b *testing.B, applier procGzip, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkGzip(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range gzipTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkGzip(b, test.proc, capsule)
			},
		)
	}
}
