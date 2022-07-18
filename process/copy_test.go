package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var copyTests = []struct {
	name     string
	proc     Copy
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON",
		Copy{
			InputKey:  "original",
			OutputKey: "copy",
		},
		[]byte(`{"original":"foo"}`),
		[]byte(`{"original":"foo","copy":"foo"}`),
		nil,
	},
	{
		"from JSON",
		Copy{
			InputKey: "copy",
		},
		[]byte(`{"copy":"foo"}`),
		[]byte(`foo`),
		nil,
	},
	{
		"to JSON utf8",
		Copy{
			OutputKey: "copy",
		},
		[]byte(`foo`),
		[]byte(`{"copy":"foo"}`),
		nil,
	},
	{
		"to JSON zlib",
		Copy{
			OutputKey: "copy",
		},
		[]byte{120, 156, 5, 192, 33, 13, 0, 0, 0, 128, 176, 182, 216, 247, 119, 44, 6, 2, 130, 1, 69},
		[]byte(`{"copy":"eJwFwCENAAAAgLC22Pd3LAYCggFF"}`),
		nil,
	},
}

func TestCopy(t *testing.T) {
	ctx := context.TODO()
	for _, test := range copyTests {
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
