package process

import (
	"bytes"
	"context"
	"testing"
)

var copyTests = []struct {
	name     string
	proc     Copy
	test     []byte
	expected []byte
}{
	{
		"json",
		Copy{
			Input:  "original",
			Output: "copy",
		},
		[]byte(`{"original":"hello"}`),
		[]byte(`{"original":"hello","copy":"hello"}`),
	},
	{
		"from json",
		Copy{
			Input: "copy",
		},
		[]byte(`{"copy":"hello"}`),
		[]byte(`hello`),
	},
	{
		"to json",
		Copy{
			Output: "copy",
		},
		[]byte(`hello`),
		[]byte(`{"copy":"hello"}`),
	},
}

func TestCopy(t *testing.T) {
	for _, test := range copyTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkCopyByte(b *testing.B, byter Copy, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkCopyByte(b *testing.B) {
	for _, test := range copyTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkCopyByte(b, test.proc, test.test)
			},
		)
	}
}
