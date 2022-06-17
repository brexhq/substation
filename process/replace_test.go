package process

import (
	"bytes"
	"context"
	"testing"
)

var replaceTests = []struct {
	name     string
	proc     Replace
	test     []byte
	expected []byte
}{
	{
		"json",
		Replace{
			Input:  "replace",
			Output: "replace",
			Options: ReplaceOptions{
				Old: "r",
				New: "z",
			},
		},
		[]byte(`{"replace":"bar"}`),
		[]byte(`{"replace":"baz"}`),
	},
	{
		"json array",
		Replace{
			Input:  "replace",
			Output: "replace",
			Options: ReplaceOptions{
				Old: "r",
				New: "z",
			},
		},
		[]byte(`{"replace":["bar","bard"]}`),
		[]byte(`{"replace":["baz","bazd"]}`),
	},
	{
		"data",
		Replace{
			Options: ReplaceOptions{
				Old: "r",
				New: "z",
			},
		},
		[]byte(`bar`),
		[]byte(`baz`),
	},
}

func TestReplace(t *testing.T) {
	ctx := context.TODO()
	for _, test := range replaceTests {
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

func benchmarkReplaceByte(b *testing.B, byter Replace, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkReplaceByte(b *testing.B) {
	for _, test := range replaceTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkReplaceByte(b, test.proc, test.test)
			},
		)
	}
}
