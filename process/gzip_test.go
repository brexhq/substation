package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var gzipTests = []struct {
	name     string
	proc     Gzip
	err      error
	test     []byte
	expected []byte
}{
	{
		"from",
		Gzip{
			Options: GzipOptions{
				Direction: "from",
			},
		},
		nil,
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
		[]byte(`foo`),
	},
	{
		"to",
		Gzip{
			Options: GzipOptions{
				Direction: "to",
			},
		},
		nil,
		[]byte(`foo`),
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0},
	},
	{
		"missing required options",
		Gzip{},
		ProcessorInvalidSettings,
		[]byte{},
		[]byte{},
	},
	{
		"unsupported direction",
		Gzip{},
		ProcessorInvalidDirection,
		[]byte(`foo`),
		[]byte{},
	},
}

func TestGzip(t *testing.T) {
	ctx := context.TODO()
	for _, test := range gzipTests {
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.As(err, &test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkGzipByte(b *testing.B, byter Gzip, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkGzipByte(b *testing.B) {
	for _, test := range gzipTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkGzipByte(b, test.proc, test.test)
			},
		)
	}
}
